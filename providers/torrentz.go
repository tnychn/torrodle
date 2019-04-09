package providers

import (
	"fmt"

	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"

	"github.com/a1phat0ny/torrodle/models"
	"github.com/a1phat0ny/torrodle/request"
	"github.com/a1phat0ny/torrodle/utils"
)

type TorrentzProvider struct {
	models.Provider
}

func NewTorrentzProvider() models.ProviderInterface {
	provider := &TorrentzProvider{}
	provider.Name = "Torrentz2"
	provider.Site = "https://torrentz2.eu"
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
	results := []models.Source{}
	query = url.QueryEscape(query)
	if count <= 0 {
		return results, nil
	}
	if categoryURL == "" {
		categoryURL = provider.Categories.All
	}

	logrus.Infoln("Torrentz2: Getting search results in parallel...")
	pages := utils.ComputePageCount(count, 50)
	logrus.Debugf("Torrentz2: pages=%d\n", pages)

	// asynchronize
	wg := sync.WaitGroup{}
	for page := 0; page <= pages; page++ {
		surl := fmt.Sprintf(string(categoryURL), query, page)
		if page == 0 {
			surl = strings.TrimSuffix(surl, "&p=0")
		}
		logrus.Debugf("Torrentz2: surl=%v\n", provider.Site+surl)
		wg.Add(1)
		go provider.extractResults(provider.Site+surl, page, &results, &wg)
	}
	wg.Wait()
	// Ending up
	logrus.Infof("Torrentz2: Found %d results\n", len(results))
	if len(results) < count {
		count = len(results)
	}
	return results[:count], nil
}

func (provider *TorrentzProvider) extractResults(surl string, page int, results *[]models.Source, wg *sync.WaitGroup) {
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
			URL:      provider.Site + URL,
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
