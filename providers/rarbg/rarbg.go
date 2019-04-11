package rarbg

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/a1phat0ny/torrodle/models"
	"github.com/a1phat0ny/torrodle/request"
)

const (
	Name = "RARBG"
	Site = "https://rarbg.to"

	apiURL   = "https://torrentapi.org"
	tokenURL = "https://torrentapi.org/pubapi_v2.php?get_token=get_token&app_id=torrodle"
)

var (
	token     string
	tokenFile string
)

type RarbgProvider struct {
	models.Provider
}

func New() models.ProviderInterface {
	provider := &RarbgProvider{}
	provider.Name = Name
	provider.Site = Site
	provider.Categories = models.Categories{
		All:   "/pubapi_v2.php?mode=search&app_id=torrodle&format=json_extended&search_string=%v&sort=seeders&limit=%d&token=",
		Movie: "/pubapi_v2.php?mode=search&app_id=torrodle&format=json_extended&search_string=%v&category=14;17;42;44;45;46;47;48;50;51;52&sort=seeders&limit=%d&token=",
		TV:    "/pubapi_v2.php?mode=search&app_id=torrodle&format=json_extended&search_string=%v&category=1;18;41;49&sort=seeders&limit=%d&token=",
		Porn:  "/pubapi_v2.php?mode=search&app_id=torrodle&format=json_extended&search_string=%v&category=1;4&sort=seeders&limit=%d&token=",
	}

	var cacheDir, _ = os.UserCacheDir()
	var dir = filepath.Join(cacheDir, "torrodle")
	os.Mkdir(dir, os.ModePerm)
	tokenFile = filepath.Join(dir, "rarbg_token.txt")
	return provider
}

func (provider *RarbgProvider) Search(query string, count int, categoryURL models.CategoryURL) ([]models.Source, error) {
	results := []models.Source{}
	if count <= 0 {
		return results, nil
	}
	// limit: only 25, 50 and 100 are valid
	if count < 25 {
		count = 25
	} else if count < 50 {
		count = 50
	} else {
		count = 100
	}
	logrus.Debugf("RARBG: count=%d\n", count)
	if categoryURL == "" {
		categoryURL = provider.Categories.All
	}
	escaped := url.QueryEscape(query)
	surl := fmt.Sprintf(string(categoryURL), escaped, count)

	// Check the cache directory for token
	logrus.Debugf("RARBG: tokenFile=%v", tokenFile)

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		// rarbg_token.txt does not exist -> get a new token
		token, err = newToken()
		if err != nil {
			return results, err
		}
	} else {
		// read token from rarbg_token.txt
		token := getToken()
		// file is empty -> get a new token
		if token == "" {
			token, err = newToken()
			if err != nil {
				return results, err
			}
		}
	}
	logrus.Debugf("RARBG: token=%v\n", token)

	surl = apiURL + surl + token
	logrus.Debugf("RARBG: surl=%v\n", surl)

	logrus.Infoln("RARBG: Getting search results...")
	_, json, err := request.Get(nil, surl, nil)
	if err != nil {
		return results, err
	}
	if json == "" {
		// empty response -> update token
		token, err = newToken()
		if err != nil {
			return results, err
		}
		// retry with the new updated token
		return provider.Search(query, count, categoryURL)
	}

	logrus.Infoln("RARBG: Extracting sources...")
	data := gjson.Get(json, "torrent_results").Array()
	for _, result := range data {
		source := models.Source{
			From:     provider.Name,
			Title:    result.Get("title").String(),
			URL:      result.Get("info_page").String(),
			Seeders:  int(result.Get("seeders").Int()),
			Leechers: int(result.Get("leechers").Int()),
			FileSize: result.Get("size").Int(),
			Magnet:   result.Get("download").String(),
		}
		if source.Title == "" || source.URL == "" || source.Seeders == 0 {
			continue
		}
		results = append(results, source)
	}
	logrus.Infof("RARBG: Found %d results\n", len(results))
	if count > len(results) {
		count = len(results)
	}
	return results[:count], nil
}

func newToken() (string, error) {
	logrus.Infoln("RARBG: Getting API token...")
	_, res, err := request.Get(nil, tokenURL, nil)
	if err != nil {
		return "", err
	}
	token := gjson.Get(string(res), "token").String()
	if token == "" {
		return "", errors.New("RARBG: error getting API token")
	}
	ioutil.WriteFile(tokenFile, []byte(token), 0777)
	return token, nil
}

func getToken() string {
	// read token from rarbg_token.txt
	t, _ := ioutil.ReadFile(tokenFile)
	token = string(t)
	return token
}
