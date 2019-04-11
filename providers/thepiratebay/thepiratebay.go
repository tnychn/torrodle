package thepiratebay

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"

	"github.com/a1phat0ny/torrodle/models"
	"github.com/a1phat0ny/torrodle/request"
)

const (
	Name = "ThePirateBay"
	Site = "https://thepiratebay.org"
)

type ThePirateBayProvider struct {
	models.Provider
}

func New() models.ProviderInterface {
	provider := &ThePirateBayProvider{}
	provider.Name = Name
	provider.Site = Site
	provider.Categories = models.Categories{
		All:   "/search/%v/%d/99/0",
		Movie: "/search/%v/%d/99/200",
		TV:    "/search/%v/%d/99/200",
		Porn:  "/search/%v/%d/99/500",
	}
	return provider
}

func (provider *ThePirateBayProvider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	results, err := provider.Query(query, categoryURL, count, 30, 0, extractor)
	return results, err
}

func extractor(surl string, page int, results *[]models.Source, wg *sync.WaitGroup) {
	logrus.Infof("ThePirateBay: [%d] Extracting results...\n", page)
	_, html, err := request.Get(nil, surl, nil)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("ThePirateBay: [%d]", page), err)
		wg.Done()
		return
	}
	sources := []models.Source{}
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	table := doc.Find("table#searchResult").Find("tbody")
	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		a := tds.Eq(1).Find("a.detLink")
		// title
		title := a.Text()
		// seeders
		s := tds.Eq(2).Text()
		seeders, _ := strconv.Atoi(strings.TrimSpace(s))
		// leechers
		l := tds.Eq(3).Text()
		leechers, _ := strconv.Atoi(strings.TrimSpace(l))
		// filesize
		re := regexp.MustCompile(`Size\s(.*?),`)
		text := tds.Eq(1).Find("font").Text()
		fs := re.FindStringSubmatch(text)[1]
		filesize, _ := humanize.ParseBytes(strings.TrimSpace(fs)) // convert human words to bytes number
		// url
		URL, _ := a.Attr("href")
		// magnet
		magnet, _ := tds.Eq(1).Find(`a[title="Download this torrent using magnet"]`).Attr("href")
		// ---
		source := models.Source{
			From:     "ThePirateBay",
			Title:    strings.TrimSpace(title),
			URL:      Site + URL,
			Seeders:  seeders,
			Leechers: leechers,
			FileSize: int64(filesize),
			Magnet:   magnet,
		}
		sources = append(sources, source)
	})

	logrus.Debugf("ThePirateBay: [%d] Amount of results: %d", page, len(sources))
	*results = append(*results, sources...)
	wg.Done()
}
