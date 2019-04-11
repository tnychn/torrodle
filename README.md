<h1 align="center">Torrodle</h1>

<p align="center"><em>Watch anything instantly just with a single command</em></p>

<p align="center"><img src="demo.gif" width=70%></p>

<p align="center">
    <a href="./LICENSE.txt"><img src="https://img.shields.io/github/license/a1phat0ny/noteboard.svg"></a>
    <a href="https://github.com/a1phat0ny"><img src="https://img.shields.io/badge/dev-a1phat0ny-orange.svg?style=flat-square&logo=github"></a>
</p>

**Torrodle** is a command-line program which search and gather magnet links of movies, tv shows, animes and porns from [providers](#available-providers).
It then streams the video via HTTP (along with its subtitles) and play it with a user preferred video player (such as *vlc* and *mpv*).

> If you don't know what BitTorrent is, you shouldn't be using **Torrodle**. There are some copyrighted content which might be illegal downloading them in your country.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
  - [Binary](#binary)
  - [Go Get](#go-get)
    - [Command-line tool](#command-line-tool)
    - [Library](#library)
  - [Build From Source](#build-from-source)
  - [Dependencies](#dependencies)
- [Usage](#usage)
- [Configurations](#configurations)
- [Providers](#available-providers)
  - [1337x](#1337x-)
  - [RARBG](#rarbg-)
  - [The Pirate Bay](#the-pirate-bay-)
  - [YIFY](#yify-)
  - [Torrentz2](#torrentz2)
  - [LimeTorrents](#limetorrents)
  - [OpenSubtitles](#opensubtitles)
- [Credit](#credit)

## Features

* Lightning fast ⚡️
* Simple to use 🚸
* Built-in torrent streaming client via HTTP (refined from `go-peerflix`)
* Watch the video while it is being downloaded 🔰 
* Query multiple providers in a single search 🔎
* Sorted results from 5 different providers at once 🚀 
* Along with subtitles fetching for the video 📄

## Installation

### Binary

**For installing the command-line tool only**

Download the latest stable release of the binary at [releases](https://github.com/a1phat0ny/torrodle/releases).

### Go Get

Make sure you have Go installed on your machine.

#### Command-line tool

`$ go get github.com/a1phat0ny/torrodle/cmd/...`

#### Library

`$ go get github.com/a1phat0ny/torrodle`

### Build From Source

**For installing the command-line tool only**

```bash
$ git clone github.com/a1phat0ny/torrodle
$ cd torrodle
$ go build cmd/torrodle/main.go
```

### Dependencies

Dependencies are listed in [`go.mod`](./go.mod) file.

1. [logrus](https://github.com/sirupsen/logrus) -- better logging
2. [goquery](https://github.com/PuerkitoBio/goquery) -- HTML parsing
3. [gjson](https://github.com/tidwall/gjson) -- JSON parsing
4. [torrent](https://github.com/anacrolix/torrent) -- torrent streaming
5. [osdb](https://github.com/oz/osdb) -- subtitles fetching from OpenSubtitles
5. [go-humanize](https://github.com/dustin/go-humanize) -- humanizing file size words
6. [color](https://github.com/fatih/color) -- colorized output
7. [tablewriter](https://github.com/olekukonko/tablewriter) -- table rendering
8. [survey](https://github.com/AlecAivazis/survey) -- pretty prompting

## Usage

Enter `torrodle` in your terminal. That's all !

This command will launch a *wizard* that will help you search for torrents.

For auto executing of video players, only **MPV** and **VLC** are supported (for now).
For other video players, you can choose `None` in video player options prompt and open your video player with the stream url.

## Configurations

**Path to the config file:** `~/.torrodle.json`

 ```json
{
	"DataDir": "",
	"ResultsLimit": 100,
	"TorrentPort": 9999,
	"HostPort": 8080,
	"Debug": false
}
```

* `DataDir` (default: `$TMPDIR/torrodle/`) : Directory where the directories of download files (and subtitles) will be stored.
* `ResultsLimit` (default: `100`) : Maximum count of results will be fetched from provider(s).
* `TorrentPort` (default: `9999`) : Listen port for the torrent client.
* `HostPort` (default: `8080`) : Listen port for HTTP localhost video streaming (`http://localhost:<port>`).
* `Debug` (default: `false`) : Detailed debug messages will be printed to output if `true`.

## API

### Constants

#### Categories

Type: *`string`*

* `CategoryAll`
* `CategoryMovie`
* `CategoryTV`
* `CategoryAnime`
* `CategoryPorn`

#### Sorts

Type: *`string`*

* `SortByDefault`
* `SortBySeeders`
* `SortByLeechers`
* `SortBySize`

#### Providers

Type: *`models.ProviderInterface`*

* `ThePirateBayProvider` (`The Pirate Bay`)
* `LimeTorrentsProvider` (`LimeTorrents`)
* `Torrentz2Provider` (`Torrentz2`)
* `RarbgProvider` (`RARBG`)
* `LeetxProvider` (`1337x`)
* `YifyProvider` (`YIFY`)

**`AllProviders`** *`[...]models.ProviderInterface`* -- an array that holds all the above providers

### Functions

```go
func ListProviderResults(provider models.ProviderInterface, query string, count int, category string, sortBy string) []models.Source
```
**ListProviderResults** lists all results queried from this specific provider only.
It sorts the results and returns at most {count} results.

<details>
  <summary>Example</summary>
  <pre>sources := torrodle.ListProviderResults(torrodle.LeetxProvider, "the great gatsby", 50, torrodle.CategoryMovie, torrodle.SortBySeeders)</pre>
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
  <p><b>You can pass in a slice of strings which are the names of the providers.</b></p>
  <pre>sources := torrodle.ListResults([]string{"1337x", "RARBG"}, "the great gatsby", 50, torrodle.CategoryMovie, torrodle.SortBySeeders)</pre>
  <p><b>You can also directly import <code>torrodle/models</code> package and pass in a slice of the provider interfaces.</b></p>
  <pre>sources := torrodle.ListResults([]models.ProviderInterface{torrodle.LeetxProvider, torrodle.RarbgProvider}, "the great gatsby", 50, torrodle.CategoryMovie, torrodle.SortBySeeders)</pre>
</details>

### Models

#### Source

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

#### Provider

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

## Available Providers

### 1337x *

**`torrodle/providers/leetx`**

* **Site:** https://1337x.to/

* **Categories:** Movie, TV, Anime, Porn (All)
 
### RARBG *

**`torrodle/providers/rarbg`**
 
* **Site:** http://rarbg.to/
 
* **Categories:** Movie, TV, Porn

### The Pirate Bay *

**`torrodle/providers/thepiratebay`**

* **Site:** https://thepiratebay.org/

* **Categories:** Movie, TV, Porn
 
### YIFY *

**`torrodle/providers/yify`**
 
* **Site:** https://yts.am/
 
* **Categories:** Movie
 
### Torrentz2

**`torrodle/providers/torrentz`**
 
* **Site:** https://torrentz2.eu/
 
* **Categories:** Movie, TV, Anime, Porn (All)
 
### LimeTorrents

**`torrodle/providers/limetorrents`**
 
* **Site:** https://www.limetorrents.info/
 
* **Categories:** Movie, TV, Anime
 
 (`*` recommended provider)
 
 More providers comming soon !
 
 ### OpenSubtitles
 
 The only provider for providing movies / tv series subtitles.
 
 API client powered by [oz/osdb](https://github.com/oz/osdb).
 
 #### Available languages:
 
 1. English `eng`
 2. Chinese (simplified) `chi`
 3. Chinese (traditional) `zht`
 4. Arabic `ara`
 5. Hindi `hin`
 6. Dutch `dut`
 7. French `fre`
 8. Russian `rus`
 9. Portuguese `por`

## Credit

This project is inspired by [@Fabio Spampinato](https://github.com/fabiospampinato)'s [cliflix](https://github.com/fabiospampinato/cliflix).

Torrent streaming technique refined from [@Sioro Neoku](https://github.com/Sioro-Neoku)'s [go-peerflix](https://github.com/Sioro-Neoku/go-peerflix).

<br>
<p align="center">Made with ❤️︎ by <a href="https://github.com/a1phat0ny">a1phat0ny</a><br>under <a href="./LICENSE.txt">MIT license</a></p>
