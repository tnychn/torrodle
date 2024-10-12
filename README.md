<h1 align="center">torrodle</h1>

<p align="center"><strong><i>A mighty all-in-one magnet scraper & streamer.</i></strong></p>

<p align="center"><img src="demo.gif" width=70%></p>

<p align="center">
    <a href="https://github.com/tnychn/torrodle/releases"><img alt="github releases" src="https://img.shields.io/github/v/release/tnychn/torrodle"></a>
    <a href="https://github.com/tnychn/torrodle/releases"><img alt="release date" src="https://img.shields.io/github/release-date/tnychn/torrodle"></a>
    <a href="https://github.com/tnychn/torrodle/releases"><img alt="downloads" src="https://img.shields.io/github/downloads/tnychn/torrodle/total"></a>
    <a href="./LICENSE.txt"><img alt="license" src="https://img.shields.io/github/license/tnychn/torrodle.svg"></a>
</p>

<div align="center">
    <strong>
    <a href="#features">Features</a> |
    <a href="#installation">Install</a> |
    <a href="PROVIDERS.md">Providers</a> |
    <a href="CLI.md">CLI</a> |
    <a href="LIB.md">Library</a>
    </strong>
</div>

---

**torrodle** is a command-line program which searches and gathers magnet links of movies, tv shows, animes and porn videos from a variety of [providers](PROVIDERS.md).
It then streams the torrent via HTTP (along with its subtitles) and plays it with a user preferred video player (such as *vlc* and *mpv*).

> If you don't know what BitTorrent is, you shouldn't be using **torrodle**.
> There are some copyrighted content which might be illegal downloading them in your country.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
  - [Pre Built Binaries](#pre-built-binaries)
  - [Go Get](#go-get)
  - [Build From Source](#build-from-source)
    - [Dependencies](#dependencies)
- [Usage](#usage)
- [Contributing](#contributing)
- [Credit](#credit)

## Features

* 🔥 Blazing fast
* 🚸 User-friendly
* 🤖 Built-in torrent streaming client via HTTP (refined from [`go-peerflix`](https://github.com/Sioro-Neoku/go-peerflix))
* 🔰 Watch the video while it is being downloaded
* 🔎 Query multiple providers in a single search
* 🚀 Sorted results from 7 different providers at once 
* 📄 Along with subtitles fetching for the video (using [`osdb`](https://github.com/Sioro-Neoku/go-peerflix))

## Installation

### Pre Built Binaries

Download the latest stable release of the binary at [releases](https://github.com/tnychn/torrodle/releases).

### Go Get

Make sure you have **Go 1.12+** installed on your machine.

`$ go get github.com/tnychn/torrodle/cmd/...`

### Build From Source

Make sure you have **Go 1.12+** installed on your machine.

```shell script
$ git clone github.com/tnychn/torrodle.git
$ cd torrodle
$ go build cmd/torrodle/main.go
```

## Usage

For command-line usage, see [`CLI.md`](README-CLI.md).

For library usage, see [`LIB.md`](README-LIB.md).

## Credit

This project is inspired by [@Fabio Spampinato](https://github.com/fabiospampinato)'s [cliflix](https://github.com/fabiospampinato/cliflix).

Torrent streaming technique adapted from [@Sioro Neoku](https://github.com/Sioro-Neoku)'s [go-peerflix](https://github.com/Sioro-Neoku/go-peerflix).

---

<div align="center">
    <sub><strong>~ crafted with ♥︎ by tnychn ~</strong></sub>
    <br>
    <sub><strong>MIT © 2019 Tony Chan</strong></sub>
</div>
