package providers

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"

	"github.com/a1phat0ny/torrodle/models"
	"github.com/a1phat0ny/torrodle/request"
	"github.com/a1phat0ny/torrodle/utils"
)

type ThePirateBayProvider struct {
	models.Provider
}

func NewThePirateBayProvider() models.ProviderInterface {
	provider := &ThePirateBayProvider{}
	provider.Name = "ThePirateBay"
	provider.Site = "https://thepiratebay.org"
	provider.Categories = models.Categories{
		All:   "/search/%v/%d/99/0",
		Movie: "/search/%v/%d/99/200",
		TV:    "/search/%v/%d/99/200",
		Porn:  "/search/%v/%d/99/500",
	}
	return provider
}

func (provider *ThePirateBayProvider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	results := []models.Source{}
	if count <= 0 {
		return results, nil
	}
	if categoryURL == "" {
		categoryURL = provider.Categories.All
	}
	query = url.QueryEscape(query)
	pages := utils.ComputePageCount(count, 30)
	logrus.Debugf("ThePirateBay: pages=%d\n", pages)
	// asynchronize
	wg := sync.WaitGroup{}
	for page := 0; page < pages; page++ {
		surl := fmt.Sprintf(string(categoryURL), query, page)
		wg.Add(1)
		go provider.extractResults(provider.Site+surl, page, &results, &wg)
	}
	wg.Wait()
	// Ending up
	logrus.Infof("ThePirateBay: Found %d results\n", len(results))
	if len(results) < count {
		count = len(results)
	}
	return results[:count], nil
}

func (provider *ThePirateBayProvider) extractResults(surl string, page int, results *[]models.Source, wg *sync.WaitGroup) {
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
			URL:      provider.Site + URL,
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
