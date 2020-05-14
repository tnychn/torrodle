/* Reference: https://stackoverflow.com/questions/26027350/go-interface-fields
- In order to specify fields that would be required on anything that implements `ProviderInterface`, a struct type `Provider` is used.
- This allows us to access methods and even fields of an interface.*/
package models

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/tnychn/torrodle/utils"
)

// ProviderInterface is an interface that exposes all the methods a `Provider` struct type has.
type ProviderInterface interface {
	String() string
	Search(string, int, CategoryURL) ([]Source, error) // search for torrents with a given (query, count, categoryURL) -> returns a slice of sources found
	Query(string, CategoryURL, int, int, int, func(string, int, *[]Source, *sync.WaitGroup)) ([]Source, error)
	GetName() string
	GetSite() string
	GetCategories() Categories
}

// Provider is a struct type that exposes fields for the `ProviderInterface`.
type Provider struct {
	Name       string
	Site       string
	Categories Categories
}

func (provider *Provider) String() string {
	return fmt.Sprintf("<Provider(name=%v, url=%v)>", provider.Name, provider.Site)
}

// Search queries the provider and returns the results (sources).
func (provider *Provider) Search(string, int, CategoryURL) ([]Source, error) {
	return []Source{}, nil
}

// GetName returns the name of this provider.
func (provider *Provider) GetName() string {
	return provider.Name
}

// GetSite returns the URL (site domain) of this provider.
func (provider *Provider) GetSite() string {
	return provider.Site
}

// GetCategories returns the categories of this provider.
func (provider *Provider) GetCategories() Categories {
	return provider.Categories
}

// Query is a universal base function for querying webpages asynchronusly.
func (provider *Provider) Query(query string, categoryURL CategoryURL, count int, perPage int, start int, extractor func(string, int, *[]Source, *sync.WaitGroup)) ([]Source, error) {
	var results []Source
	if count <= 0 {
		return results, nil
	}

	query = url.QueryEscape(query)
	if categoryURL == "" {
		categoryURL = provider.Categories.All
	}
	logrus.Infof("%v: Getting search results in parallel...\n", provider.Name)
	pages := utils.ComputePageCount(count, perPage)
	logrus.Debugf("%v: pages=%d\n", provider.Name, pages)

	// asynchronize
	wg := sync.WaitGroup{}
	for page := start; page <= pages; page++ {
		surl := fmt.Sprintf(string(categoryURL), query, page)
		wg.Add(1)
		go extractor(provider.Site+surl, page, &results, &wg)
	}
	wg.Wait()

	// Ending up
	logrus.Infof("%v: Found %d results\n", provider.Name, len(results))
	if len(results) < count {
		count = len(results)
	}
	return results[:count], nil
}

// Category is a custom type which represents a URL of a Category.
type CategoryURL string

// Categories is a collection of CategoryURL types.
type Categories struct {
	All   CategoryURL
	Movie CategoryURL
	TV    CategoryURL
	Anime CategoryURL
	Porn  CategoryURL
}

// Source provides informational fields for a torrent source.
type Source struct {
	From     string
	Title    string
	URL      string
	Seeders  int
	Leechers int
	FileSize int64
	Magnet   string
}

func (source Source) String() string {
	return fmt.Sprintf("<Source(title=%v)>", source.Title)
}
