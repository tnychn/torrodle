package torrentz

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"

	"github.com/tnychn/torrodle/models"
	"github.com/tnychn/torrodle/request"
)

const (
	Name = "Torrentz2"
	Site = "https://torrentz2.eu"
)

type TorrentzProvider struct {
	models.Provider
}

func New() models.ProviderInterface {
	provider := &TorrentzProvider{}
	provider.Name = Name
	provider.Site = Site
	provider.Categories = models.Categories{
		All:   "/search?f=%v&p=%d",
		Movie: "/search?f=%v&p=%d",
		TV:    "/search?f=%v&p=%d",
		Anime: "/search?f=%v&p=%d",
		Porn:  "/search?f=%v&p=%d",
	}
	return provider
}

func (provider *TorrentzProvider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	results, err := provider.Query(query, categoryURL, count, 50, 0, extractor)
	return results, err
}

func extractor(surl string, page int, results *[]models.Source, wg *sync.WaitGroup) {
	logrus.Infof("Torrentz2: [%d] Extracting results...\n", page)
	_, html, err := request.Get(nil, surl, nil)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("Torrentz2: [%d]", page), err)
		wg.Done()
		return
	}
	sources := []models.Source{}
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	div := doc.Find("div.results")
	div.Find("dl").Each(func(i int, s *goquery.Selection) {
		// title
		title := s.Find("dt").Find("a").Text()

		spans := s.Find("dd").Find("span")
		// filesize
		filesize, _ := humanize.ParseBytes(strings.TrimSpace(spans.Eq(2).Text()))
		// seeders
		seeders, _ := strconv.Atoi(spans.Eq(3).Text())
		// leechers
		leechers, _ := strconv.Atoi(spans.Eq(4).Text())
		// url
		URL, _ := s.Find("dt").Find("a").Attr("href")
		// magnet
		hash := strings.TrimLeft(URL, "/")
		magnet := fmt.Sprintf("magnet:?xt=urn:btih:%v", hash)

		if title == "" || URL == "" || seeders == 0 {
			return
		}
		// ---
		source := models.Source{
			From:     "Torrentz2",
			Title:    strings.TrimSpace(title),
			URL:      Site + URL,
			Seeders:  seeders,
			Leechers: leechers,
			FileSize: int64(filesize),
			Magnet:   magnet,
		}
		sources = append(sources, source)
	})
	logrus.Debugf("Torrentz2: [%d] Amount of results: %d", page, len(sources))
	*results = append(*results, sources...)
	wg.Done()
}
