package sukebei

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"

	"github.com/tnychn/torrodle/models"
	"github.com/tnychn/torrodle/request"
)

const (
	Name = "Sukebei"
	Site = "https://sukebei.nyaa.si"
)

type provider struct {
	models.Provider
}

func New() models.ProviderInterface {
	provider := &provider{}
	provider.Name = Name
	provider.Site = Site
	provider.Categories = models.Categories{
		All:  "/?f=0&c=0_0&q=%v&s=seeders&o=desc&p=%d",
		Porn: "/?f=0&c=0_0&q=%v&s=seeders&o=desc&p=%d",
	}
	return provider
}

func (provider *provider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	results, err := provider.Query(query, categoryURL, count, 75, 1, extractor)
	return results, err
}

func extractor(surl string, page int, results *[]models.Source, wg *sync.WaitGroup) {
	logrus.Infof("Sukebei: [%d] Extracting results...\n", page)
	_, html, err := request.Get(nil, surl, nil)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("Sukebei: [%d]", page), err)
		wg.Done()
		return
	}
	var sources []models.Source
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	table := doc.Find("table.table.table-bordered.table-hover.table-striped.torrent-list")
	table.Find("tr.default").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td.text-center")
		a := tr.Find("td[colspan]")
		// title
		title := a.Text()
		// seeders
		s := tds.Eq(3).Text()
		seeders, _ := strconv.Atoi(strings.TrimSpace(s))
		// leechers
		l := tds.Eq(4).Text()
		leechers, _ := strconv.Atoi(strings.TrimSpace(l))
		// filesize
		fs := tds.Eq(1).Text()
		filesize, _ := humanize.ParseBytes(strings.TrimSpace(fs)) // convert human words to bytes number
		// url
		URL, _ := a.Attr("href")
		// magnet
		magnet, _ := tds.Eq(0).Find("a").Eq(1).Attr("href")
		// ---
		source := models.Source{
			From:     "Sukebei",
			Title:    strings.TrimSpace(title),
			URL:      Site + URL,
			Seeders:  seeders,
			Leechers: leechers,
			FileSize: int64(filesize),
			Magnet:   magnet,
		}
		sources = append(sources, source)
	})
	logrus.Debugf("Sukebei: [%d] Amount of results: %d", page, len(sources))
	*results = append(*results, sources...)
	wg.Done()
}
