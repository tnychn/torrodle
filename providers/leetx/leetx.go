package leetx

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/avast/retry-go"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tnychn/torrodle/models"
	"github.com/tnychn/torrodle/request"
)

const (
	Name = "1337x"
	Site = "https://1337x.to"
)

type provider struct {
	models.Provider
	*rate.Limiter
}

func New() models.ProviderInterface {
	provider := &provider{}
	provider.Name = Name
	provider.Site = Site
	provider.Categories = models.Categories{
		All:   "/search/%v/%d/",
		Movie: "/category-search/%v/Movies/%d/",
		TV:    "/category-search/%v/TV/%d/",
		Anime: "/category-search/%v/Anime/%d/",
		Porn:  "/category-search/%v/XXX/%d/",
	}
	provider.Limiter = rate.NewLimiter(rate.Every(time.Second), 1)
	return provider
}

func (provider *provider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	perPage := 40
	if categoryURL == provider.Categories.All {
		perPage = 20
	}
	results, err := provider.Query(query, categoryURL, count, perPage, 0, provider.extractor)
	return results, err
}

func (provider *provider) extractor(surl string, page int, results *[]models.Source, wg *sync.WaitGroup) {
	logrus.Infof("1337x: [%d] Extracting results...\n", page)
	var html string
	err := retry.Do(func() (err error) {
		_ = provider.Wait(context.Background())
		_, html, err = request.Get(nil, surl, nil)
		return
	},
		retry.RetryIf(func(err error) bool {
			return err.Error() == http.StatusText(503)
		}),
		retry.Attempts(3),
	)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("1337x: [%d]", page), err)
		wg.Done()
		return
	}

	var sources []models.Source // Temporary array for storing models.Source(s) but without magnet and torrent links
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	table := doc.Find("table.table-list.table.table-responsive.table-striped")
	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		// title
		title := tr.Find("td.coll-1.name").Text()
		// seeders
		s := tr.Find("td.coll-2.seeds").Text()
		seeders, _ := strconv.Atoi(strings.TrimSpace(s))
		// leechers
		l := tr.Find("td.coll-3.leeches").Text()
		leechers, _ := strconv.Atoi(strings.TrimSpace(l))
		// filesize
		tr.Find("td.coll-4.size").Find("span.seeds").Remove()
		filesize, _ := humanize.ParseBytes(strings.TrimSpace(tr.Find("td.coll-4.size").Text())) // convert human words to bytes number
		// url
		URL, _ := tr.Find(`a[href^="/torrent"]`).Attr("href")
		if title == "" || URL == "" || seeders == 0 {
			return
		}
		// ---
		source := models.Source{
			From:     "1337x",
			Title:    strings.TrimSpace(title),
			URL:      Site + URL,
			Seeders:  seeders,
			Leechers: leechers,
			FileSize: int64(filesize),
		}
		sources = append(sources, source)
	})

	logrus.Debugf("1337x: [%d] Amount of results: %d", page, len(sources))
	logrus.Debugf("1337x: [%d] Getting sources in parallel...", page)
	group := sync.WaitGroup{}
	for _, source := range sources {
		group.Add(1)
		go func(source models.Source) {
			var magnet string

			_, html, err := request.Get(nil, source.URL, nil)
			if err != nil {
				logrus.Errorln(err)
				group.Done()
				return
			}
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			dropdown := doc.Find("ul.dropdown-menu")
			li := dropdown.Find("li")
			if li != nil {
				if val, ok := li.Last().Find("a").Attr("href"); ok {
					magnet = val
				}
			}
			// Assignment
			source.Magnet = magnet
			*results = append(*results, source)
			group.Done()
		}(source)
	}
	group.Wait()
	wg.Done()
}
