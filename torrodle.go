package torrodle

import (
	"sort"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	"github.com/tnychn/torrodle/models"
	"github.com/tnychn/torrodle/providers/leetx"
	"github.com/tnychn/torrodle/providers/limetorrents"
	"github.com/tnychn/torrodle/providers/rarbg"
	"github.com/tnychn/torrodle/providers/sukebei"
	"github.com/tnychn/torrodle/providers/thepiratebay"
	"github.com/tnychn/torrodle/providers/torrentz"
	"github.com/tnychn/torrodle/providers/yify"
)

type Category string
type SortBy string

const (
	CategoryAll   Category = "ALL"
	CategoryMovie Category = "MOVIE"
	CategoryTV    Category = "TV"
	CategoryAnime Category = "ANIME"
	CategoryPorn  Category = "PORN"

	SortByDefault  SortBy = "default"
	SortBySeeders  SortBy = "seeders"
	SortByLeechers SortBy = "leechers"
	SortBySize     SortBy = "size"
)

// Expose all the providers
var (
	SukebeiProvider      = sukebei.New()
	ThePirateBayProvider = thepiratebay.New()
	LimeTorrentsProvider = limetorrents.New()
	Torrentz2Provider    = torrentz.New()
	RarbgProvider        = rarbg.New()
	LeetxProvider        = leetx.New()
	YifyProvider         = yify.New()
)

var AllProviders = [...]models.ProviderInterface{
	SukebeiProvider,
	ThePirateBayProvider,
	LimeTorrentsProvider,
	Torrentz2Provider,
	RarbgProvider,
	LeetxProvider,
	YifyProvider,
}

// ListProviderResults lists all results queried from this specific provider only.
// It sorts the results and returns at most {count} results.
func ListProviderResults(provider models.ProviderInterface, query string, count int, category Category, sortBy SortBy) []models.Source {
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
func ListResults(providers []interface{}, query string, count int, category Category, sortBy SortBy) []models.Source {
	var argProviders []models.ProviderInterface
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
	var results []models.Source
	for _, provider := range argProviders {
		if showSpinner {
			c := color.New(color.FgYellow, color.Bold)
			s = spinner.New(spinner.CharSets[33], 100*time.Millisecond)
			_ = s.Color("fgBlue")
			s.Suffix = c.Sprint(" Waiting for ") + color.GreenString(provider.GetName()) + c.Sprint(" ...")
			s.Start()
		}

		sources := ListProviderResults(provider, query, count, category, sortBy)
		results = append(results, sources...)
		if showSpinner && s != nil {
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
func GetCategoryURL(category Category, categories models.Categories) models.CategoryURL {
	var caturl models.CategoryURL
	switch category {
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
		logrus.Fatalln("Invalid category")
	}
	return caturl
}

func GetSortedResults(results []models.Source, sortBy SortBy) []models.Source {
	// Sort results
	switch sortBy {
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
		logrus.Fatalln("Invalid SortBy")
	}
	return results
}
