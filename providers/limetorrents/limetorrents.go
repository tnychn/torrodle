package limetorrents

import (
	"fmt"

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
	Name = "LimeTorrents"
	Site = "https://www.limetorrents.info"
)

type LimeTorrentsProvider struct {
	models.Provider
}

func New() models.ProviderInterface {
	provider := &LimeTorrentsProvider{}
	provider.Name = Name
	provider.Site = Site
	provider.Categories = models.Categories{
		All:   "/search/all/%v/seeds/%d",
		Movie: "/search/movies/%v/seeds/%d",
		TV:    "/search/tv/%v/seeds/%d",
		Anime: "/search/anime/%v/seeds/%d",
	}
	return provider
}

func (provider *LimeTorrentsProvider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	results, err := provider.Query(query, categoryURL, count, 50, 1, extractor)
	return results, err
}

func extractor(surl string, page int, results *[]models.Source, wg *sync.WaitGroup) {
	logrus.Infof("LimeTorrents: [%d] Extracting results...\n", page)
	_, html, err := request.Get(nil, surl, nil)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("LimeTorrents: [%d]", page), err)
		wg.Done()
		return
	}

	sources := []models.Source{}
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	table := doc.Find("table.table2")
	table.Find(`tr[bgcolor="#F4F4F4"]`).Each(func(_ int, tr *goquery.Selection) {
		// title and url
		var magnet, title, URL string
		tr.Find("div.tt-name").Find("a").Each(func(i int, a *goquery.Selection) {
			cls, _ := a.Attr("class")
			if cls == "csprite_dl14" {
				torrent, _ := a.Attr("href")
				torrent = strings.Replace(torrent, "http://itorrents.org/torrent/", "", 1)
				torrentFile := strings.Split(torrent, "?")[0]
				hash := strings.TrimSuffix(torrentFile, ".torrent")
				magnet = fmt.Sprintf("magnet:?xt=urn:btih:%v", hash)
			} else {
				title = strings.TrimSpace(a.Text())
				URL, _ = a.Attr("href")
			}
		})
		// filesize
		filesize, _ := humanize.ParseBytes(strings.TrimSpace(tr.Find("td.tdnormal").Eq(1).Text())) // convert human words to bytes number
		// seeders
		s := tr.Find("td.tdseed").Text()
		seeders, _ := strconv.Atoi(strings.Replace(s, ",", "", -1))
		// leechers
		l := tr.Find("td.tdleech").Text()
		leechers, _ := strconv.Atoi(strings.Replace(l, ",", "", -1))
		// ---
		if title == "" || URL == "" || seeders == 0 {
			return
		}
		source := models.Source{
			From:     "LimeTorrents",
			Title:    title,
			URL:      Site + URL,
			Seeders:  seeders,
			Leechers: leechers,
			FileSize: int64(filesize),
			Magnet:   magnet,
		}
		sources = append(sources, source)
	})

	logrus.Debugf("LimeTorrents: [%d] Amount of results: %d", page, len(sources))
	*results = append(*results, sources...)
	wg.Done()
}
