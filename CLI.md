# Torrodle CLI Usage

[⬅️ Back to Main](./README.md)

The command-line usage of **torrodle**.

## Index

1. [Search for magnets](#search-for-magnets)
2. [Stream from your own magnet](#stream-from-your-own-magnet)
3. [Configurations](#configurations)

---

**Recommended video player:** [mpv](https://mpv.io)

> **NOTE:** For automatically launch of video players, only **mpv** and **vlc** are supported.
> If you want to use other video players, you can choose `None` in the video player options prompt.
> Then open up your video player and play from the stream url (default: http://localhost:8080).

## Search for magnets

`$ torrodle`

That's it!
This command will launch a *wizard* that will help you search for magnet links.

## Stream from your own magnet

`$ torrodle "your magnet uri"`

Then choose your preferred video player and enjoy!

## Configurations

**Path to the config file:** `~/.torrodle.json`

* **`DataDir`** (`$TMPDIR/torrodle/`) -- Directory where the directories of download files (and subtitles) will be stored.
* **`ResultsLimit`** (`100`) -- Maximum count of results will be fetched from provider(s).
* **`TorrentPort`** (`9999`) -- Listen port for the torrent client.
* **`HostPort`** (`8080`) -- Listen port for HTTP localhost video streaming (`http://localhost:<port>`).
* **`Debug`** (`false`) -- Detailed debug messages will be printed to output if `true`.
