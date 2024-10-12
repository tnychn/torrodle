# Torrodle Library Usage

[⬅️ Back to Main](./README.md)

The reference for developers to work with **torrodle** by using it as a library.

## Index

1. [Constants](#constants)
    * [Categories](#categories)
    * [Sorts](#sorts)
    * [Providers](#providers)
2. [Functions](#functions)
    * [ListProviderResults](#functions)
    * [ListResults](#functions)
3. [Models](#models)
    * [Source](#source)
    * [Provider](#provider)

---

## Constants

### Categories

* `CategoryAll`
* `CategoryMovie`
* `CategoryTV`
* `CategoryAnime`
* `CategoryPorn`

### Sorts

* `SortByDefault`
* `SortBySeeders`
* `SortByLeechers`
* `SortBySize`

### Providers

**Interface: `models.ProviderInterface`**

* `SukebeiProvider` (`Sukebei`)
* `ThePirateBayProvider` (`The Pirate Bay`)
* `LimeTorrentsProvider` (`LimeTorrents`)
* `Torrentz2Provider` (`Torrentz2`)
* `RarbgProvider` (`RARBG`)
* `LeetxProvider` (`1337x`)
* `YifyProvider` (`YTS`)

`AllProviders` *`[...]models.ProviderInterface`* An array that holds all the above providers

## Functions

```go
func ListProviderResults(provider models.ProviderInterface, query string, count int, category string, sortBy string) []models.Source
```
**ListProviderResults** lists all results queried from this specific provider only.
It sorts the results and returns at most {count} results.

<details>
  <summary>Example</summary>
  <pre><code>sources := torrodle.ListProviderResults(torrodle.LeetxProvider, "the great gatsby", 50, torrodle.CategoryMovie, torrodle.SortBySeeders)</code></pre>
</details>

<br>

```go
func ListResults(providers []interface{}, query string, count int, category string, sortBy string) []models.Source
```
**ListResults** lists all results queried from all the specified providers.
It sorts the results after collected all the sorted results from different providers.
Returns at most {count} results.

<details>
  <summary>Example</summary>
  <sub>You can pass in a slice of strings which are the names of the providers.</sub>
  <code>sources := torrodle.ListResults([]string{"1337x", "RARBG"}, "the great gatsby", 50, torrodle.CategoryMovie, torrodle.SortBySeeders)</code>
  <sub>You can also directly import <code>torrodle/models</code> package and pass in a slice of the provider interfaces.</sub>
  <code>sources := torrodle.ListResults([]models.ProviderInterface{torrodle.LeetxProvider, torrodle.RarbgProvider}, "the great gatsby", 50, torrodle.CategoryMovie, torrodle.SortBySeeders)</code>
</details>

## Models

### Source

```go
// Source provides informational fields for a torrent source.
type Source struct {
    From     string // which provider this source is from
    Title    string // title name of this source
    URL      string // URL to the info page of this source
    Seeders  int    // amount of seeders
    Leechers int    // amount of leechers
    FileSize int64  // file size of this source in bytes
    Magnet   string // magnet uri of this source
}
```

### Provider

```go
// ProviderInterface is an interface that provides all the methods a `Provider` struct type has.
type ProviderInterface interface {
    String() string // stringer
    Search(string, int, CategoryURL) ([]Source, error) // search for torrents with a given (query, count, categoryURL) -> returns a slice of sources found
    GetName() string // GetName returns the name of this provider.
    GetSite() string // GetSite returns the URL (site domain) of this provider.
    GetCategories() Categories // GetCategories returns the categories of this provider.
}
```

```go
// Provider is a struct type that exposes fields for the `ProviderInterface`.
type Provider struct {
    Name       string
    Site       string
    Categories Categories
}
```
