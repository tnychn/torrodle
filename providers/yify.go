package providers

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/a1phat0ny/torrodle/models"
	"github.com/a1phat0ny/torrodle/request"
	"github.com/a1phat0ny/torrodle/utils"
)

var trackers = [...]string{
	"udp://open.demonii.com:1337/announce",
	"udp://tracker.openbittorrent.com:80",
	"udp://tracker.coppersurfer.tk:6969",
	"udp://glotorrents.pw:6969/announce",
	"udp://tracker.opentrackr.org:1337/announce",
	"udp://torrent.gresille.org:80/announce",
	"udp://p4p.arenabg.com:1337",
	"udp://tracker.leechers-paradise.org:6969",
	"http://track.one:1234/announce",
	"udp://track.two:80",
}

type YifyProvider struct {
	models.Provider
	apiURL string
}

func NewYifyProvider() models.ProviderInterface {
	provider := &YifyProvider{}
	provider.Name = "YIFY"
	provider.Site = "https://yts.am"
	provider.Categories = models.Categories{
		All:   "/v2/list_movies.json?query_term=%v&limit=50&page=%d",
		Movie: "/v2/list_movies.json?query_term=%v&limit=50&page=%d",
	} // this provider can only search for movies
	provider.apiURL =  "https://yts.am/api"
	return provider
}

func (provider *YifyProvider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	// categoryURL will be ignored since this provider only searches for movies
	results := []models.Source{}
	if count <= 0 {
		return results, nil
	}
	if count > 20 { // query was successful but no movie is returned -> YIFY API bug ?
		count = 20
	}
	// Build URL
	if categoryURL == "" {
		categoryURL = provider.Categories.Movie
	}
	query = url.QueryEscape(query)
	pages := utils.ComputePageCount(count, 50)
	surl := fmt.Sprintf(string(categoryURL), query, pages)
	logrus.Debugf("YIFY: pages=%d\n", pages)

	// Extract sources
	logrus.Infoln("YIFY: Getting search results...")
	_, json, err := request.Get(nil, provider.apiURL+surl, nil)
	if err != nil {
		return results, err
	}

	status := gjson.Get(json, "status").String()
	msg := gjson.Get(json, "status_message").String()
	logrus.Debugln("YIFY: Message ->", msg)
	if status != "ok" {
		return results, errors.New("YIFY: returned a non-ok")
	}

	logrus.Infoln("YIFY: Extracting sources...")
	data := gjson.Get(json, "data")
	movies := data.Get("movies").Array()
	for _, movie := range movies {
		title := movie.Get("title_long").String()
		URL := movie.Get("url").String()
		source := models.Source{
			From:  provider.Name,
			Title: title,
			URL:   URL,
		}
		torrents := movie.Get("torrents").Array()
		for _, torrent := range torrents {
			s := source
			s.Title += " " + torrent.Get("quality").String() + " " + torrent.Get("type").String() + " YIFY"
			s.Seeders = int(torrent.Get("seeds").Int())
			s.Leechers = int(torrent.Get("peers").Int())
			s.FileSize = torrent.Get("size_bytes").Int()
			// filter out invalid sources
			if s.Seeders == 0 {
				continue
			}
			// build magnet uri
			hash := torrent.Get("hash").String()
			encodedName := url.PathEscape(movie.Get("title").String())
			magnet := fmt.Sprintf("magnet:?xt=urn:btih:%v&dn=%v", hash, encodedName)
			for _, tracker := range trackers {
				magnet += "&tr=" + tracker
			}
			s.Magnet = magnet
			results = append(results, s)
		}
	}
	logrus.Infof("YIFY: Found %d results\n", len(results))
	if count > len(results) {
		count = len(results)
	}
	return results[:count], nil
}
