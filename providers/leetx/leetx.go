package leetx

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
	Name = "1337x"
	Site = "https://1337x.to"
)

type LeetxProvider struct {
	models.Provider
}

func New() models.ProviderInterface {
	provider := &LeetxProvider{}
	provider.Name = Name
	provider.Site = Site
	provider.Categories = models.Categories{
		All:   "/search/%v/%d/",
		Movie: "/category-search/%v/Movies/%d/",
		TV:    "/category-search/%v/TV/%d/",
		Anime: "/category-search/%v/Anime/%d/",
		Porn:  "/category-search/%v/XXX/%d/",
	}
	return provider
}

func (provider *LeetxProvider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	perPage := 40
	if categoryURL == provider.Categories.All {
		perPage = 20
	}
	results, err := provider.Query(query, categoryURL, count, perPage, 0, extractor)
	return results, err
}

func extractor(surl string, page int, results *[]models.Source, wg *sync.WaitGroup) {
	logrus.Infof("1337x: [%d] Extracting results...\n", page)
	_, html, err := request.Get(nil, surl, nil)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("1337x: [%d]", page), err)
		wg.Done()
		return
	}

	sources := []models.Source{} // Temporary array for storing models.Source(s) but without magnet and torrent links
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
		go getSourceWorker(source, results, &group) // directly append the completed sources to `results`
	}
	group.Wait()
	wg.Done()
}

func getSourceWorker(source models.Source, results *[]models.Source, group *sync.WaitGroup) {
	var magnet string
	var torrents []string

	_, html, err := request.Get(nil, source.URL, nil)
	if err != nil {
		logrus.Errorln(err)
		group.Done()
		return
	}
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	list := doc.Find("ul.download-links-dontblock.btn-wrap-list")
	// Get all urls
	torrentsList := list.Find("ul.dropdown-menu")
	torrentsList.Find("a.btn").Each(func(i int, s *goquery.Selection) {
		torrent, _ := s.Attr("href")
		if torrent != "" {
			torrents = append(torrents, torrent)
		}
	})
	if len(torrents) != 0 {
		magnet = torrents[len(torrents)-1]
	}
	// Assignment
	source.Magnet = magnet
	*results = append(*results, source)
	group.Done()
}
