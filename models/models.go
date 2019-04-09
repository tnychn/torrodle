/* Reference: https://stackoverflow.com/questions/26027350/go-interface-fields
- In order to specify fields that would be required on anything that implements `ProviderInterface`, a struct type `Provider` is used.
- This allows us to access methods and even fields of an interface.*/
package models

import "fmt"

// ProviderInterface is an interface that provides all the methods a `Provider` struct type has.
type ProviderInterface interface {
	String() string
	Search(string, int, CategoryURL) ([]Source, error) // search for torrents with a given (query, count, categoryURL) -> returns a slice of sources found
	// Getters
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

// Category is a custom type which represents a URL of a Category.
type CategoryURL string

// Categories is a collection of CategoryURL types.
// See also: `utils.GetCategoryURL`
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
