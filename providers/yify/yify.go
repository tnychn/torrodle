package yify

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"

	"github.com/tnychn/torrodle/models"
	"github.com/tnychn/torrodle/request"
	"github.com/tnychn/torrodle/utils"
)

const (
	Name = "YIFY"
	Site = "https://yts.am"

	apiURL = "https://yts.am/api"
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

type provider struct {
	models.Provider
}

func New() models.ProviderInterface {
	provider := &provider{}
	provider.Name = Name
	provider.Site = Site
	provider.Categories = models.Categories{
		All:   "/v2/list_movies.json?query_term=%v&limit=50&page=%d",
		Movie: "/v2/list_movies.json?query_term=%v&limit=50&page=%d",
	} // this provider can only search for movies
	return provider
}

type apiResponse struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
	Data          struct {
		Movies []struct {
			URL       string `json:"url"`
			Title     string `json:"title"`
			TitleLong string `json:"title_long"`
			Torrents  []struct {
				URL       string `json:"url"`
				Hash      string `json:"hash"`
				Quality   string `json:"quality"`
				Type      string `json:"type"`
				Seeds     int    `json:"seeds"`
				Peers     int    `json:"peers"`
				SizeBytes int64  `json:"size_bytes"`
			} `json:"torrents"`
		} `json:"movies"`
	} `json:"data"`
}

func (provider *provider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	// categoryURL will be ignored since this provider only searches for movies
	var results []models.Source
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
	_, resp, err := request.Get(nil, apiURL+surl, nil)
	if err != nil {
		return results, err
	}

	response := apiResponse{}
	if err = json.Unmarshal([]byte(resp), &response); err != nil {
		return results, err
	}

	status := response.Status
	msg := response.StatusMessage
	logrus.Debugln("YIFY: Message ->", msg)
	if status != "ok" {
		return results, errors.New("YIFY: returned a non-ok")
	}

	logrus.Infoln("YIFY: Extracting sources...")
	data := response.Data
	movies := data.Movies
	for _, movie := range movies {
		source := models.Source{
			From:  provider.Name,
			Title: movie.TitleLong,
			URL:   movie.URL,
		}
		torrents := movie.Torrents
		for _, torrent := range torrents {
			s := source
			s.Title += " " + torrent.Quality + " " + torrent.Type + " YIFY"
			s.Seeders = torrent.Seeds
			s.Leechers = torrent.Peers
			s.FileSize = torrent.SizeBytes
			// filter out invalid sources
			if s.Seeders == 0 {
				continue
			}
			// build magnet uri
			hash := torrent.Hash
			encodedName := url.PathEscape(movie.Title)
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
