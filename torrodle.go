package torrodle

import (
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	"github.com/a1phat0ny/torrodle/models"
	"github.com/a1phat0ny/torrodle/providers/leetx"
	"github.com/a1phat0ny/torrodle/providers/limetorrents"
	"github.com/a1phat0ny/torrodle/providers/rarbg"
	"github.com/a1phat0ny/torrodle/providers/thepiratebay"
	"github.com/a1phat0ny/torrodle/providers/torrentz"
	"github.com/a1phat0ny/torrodle/providers/yify"
)

const (
	CategoryAll   string = "ALL"
	CategoryMovie string = "MOVIE"
	CategoryTV    string = "TV"
	CategoryAnime string = "ANIME"
	CategoryPorn  string = "PORN"

	SortByDefault  string = "default"
	SortBySeeders  string = "seeders"
	SortByLeechers string = "leechers"
	SortBySize     string = "size"
)

// Expose all the providers
var ThePirateBayProvider = thepiratebay.New()
var LimeTorrentsProvider = limetorrents.New()
var Torrentz2Provider = torrentz.New()
var RarbgProvider = rarbg.New()
var LeetxProvider = leetx.New()
var YifyProvider = yify.New()

var AllProviders = [...]models.ProviderInterface{
	ThePirateBayProvider,
	LimeTorrentsProvider,
	Torrentz2Provider,
	RarbgProvider,
	LeetxProvider,
	YifyProvider,
}

// ListProviderResults lists all results queried from this specific provider only.
// It sorts the results and returns at most {count} results.
func ListProviderResults(provider models.ProviderInterface, query string, count int, category string, sortBy string) []models.Source {
	var sources []models.Source
	categories := provider.GetCategories()
	caturl := GetCategoryURL(category, categories)
	if caturl == "" {
		logrus.Warningf("'%v' provider does not support category '%v', getting default category (ALL)...", provider.GetName(), category)
	}
	sources, err := provider.Search(query, count, caturl)
	if err != nil {
		logrus.Fatalln(err)
	}
	if len(sources) == 0 {
		logrus.Warningf("No torrents found via '%v'\n", provider.GetName())
	}
	results := GetSortedResults(sources, sortBy)
	if count > len(results) {
		count = len(results)
	}
	return results[:count]
}

// ListResults lists all results queried from all the specified providers.
// It sorts the results after collected all the sorted results from different providers.
// Returns at most {count} results.
func ListResults(providers []interface{}, query string, count int, category string, sortBy string) []models.Source {
	argProviders := []models.ProviderInterface{}
	for _, p := range providers {
		switch p.(type) {
		case string:
			for _, p2 := range AllProviders {
				if p2.GetName() == p.(string) {
					argProviders = append(argProviders, p2)
				}
			}
		case models.ProviderInterface:
			argProviders = append(argProviders, p.(models.ProviderInterface))
		default:
			logrus.Fatalln("Invalid interface type in 'providers': only 'string' and 'models.ProviderInterface' are accepted")
		}
	}

	// Init spinner
	var s *spinner.Spinner
	showSpinner := logrus.GetLevel() <= logrus.WarnLevel

	if count > 500 {
		logrus.Warningln("'count' should not be larger than 500, set to 500 automatically")
		count = 500
	}

	// Get results from providers
	results := []models.Source{}
	for _, provider := range argProviders {
		if showSpinner {
			c := color.New(color.FgYellow, color.Bold)
			s = spinner.New(spinner.CharSets[33], 100*time.Millisecond)
			s.Color("fgBlue")
			s.Suffix = c.Sprint(" Waiting for ") + color.GreenString(provider.GetName()) + c.Sprint(" ...")
			s.Start()
		}

		sources := ListProviderResults(provider, query, count, category, sortBy)
		results = append(results, sources...)
		if showSpinner {
			s.Stop()
		}
	}
	logrus.Infof("Returning %d results in total...\n", len(results))

	results = GetSortedResults(results, sortBy)
	if count > len(results) {
		count = len(results)
	}
	return results[:count]
}

// GetCategoryURL returns CategoryURL according to the category name (constant).
func GetCategoryURL(category string, categories models.Categories) models.CategoryURL {
	var caturl models.CategoryURL
	switch strings.ToUpper(category) {
	case CategoryAll:
		caturl = categories.All
	case CategoryMovie:
		caturl = categories.Movie
	case CategoryTV:
		caturl = categories.TV
	case CategoryAnime:
		caturl = categories.Anime
	case CategoryPorn:
		caturl = categories.Porn
	default:
		logrus.Fatalf("Invalid category: %v\n", category)
	}
	return caturl
}

func GetSortedResults(results []models.Source, sortBy string) []models.Source {
	// Sort results
	switch strings.ToLower(sortBy) {
	case SortByDefault:
		// nothing to do
	case SortBySeeders:
		sort.Slice(results, func(i, j int) bool {
			return results[i].Seeders > results[j].Seeders
		})
	case SortByLeechers:
		sort.Slice(results, func(i, j int) bool {
			return results[i].Leechers > results[j].Leechers
		})
	case SortBySize:
		sort.Slice(results, func(i, j int) bool {
			return results[i].FileSize > results[j].FileSize
		})
	default:
		logrus.Fatalf("Invalid SortBy: %v", sortBy)
	}
	return results
}
