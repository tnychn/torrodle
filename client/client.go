/*
This package is the embeded version of 'github.com/Sioro-Neoku/go-peerflix/'.
We did some modifications on it in order to let it fit into 'Torrodle'
*/
package client

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"

	"github.com/tnychn/torrodle/models"
)

// Client manages the torrent downloading.
type Client struct {
	Client       *torrent.Client
	ClientConfig *torrent.ClientConfig
	Torrent      *torrent.Torrent
	Source       models.Source
	URL          string
	HostPort     int
}

// NewClient initializes a new torrent client.
func NewClient(dataDir string, torrentPort int, hostPort int) (Client, error) {
	var client Client

	// Initialize Config
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.DataDir = dataDir
	clientConfig.ListenPort = torrentPort
	clientConfig.NoUpload = true
	clientConfig.Seed = false
	clientConfig.Debug = false
	client.ClientConfig = clientConfig

	// Create Client
	c, err := torrent.NewClient(clientConfig)
	if err != nil {
		return client, err
	}
	client.Client = c
	client.HostPort = hostPort

	return client, err
}

// SetSource sets the source (magnet uri) which the client is based on.
// * must be called before `Client.Start()`
func (client *Client) SetSource(source models.Source) (*Client, error) {
	client.Source = source
	t, err := client.Client.AddMagnet(source.Magnet)
	if err == nil {
		t.SetDisplayName(source.Title)
		client.Torrent = t
	}
	return client, err
}

func (client *Client) getLargestFile() *torrent.File {
	var largestFile *torrent.File
	var lastFileSize int64
	for _, file := range client.Torrent.Files() {
		if file.Length() > lastFileSize {
			lastFileSize = file.Length()
			largestFile = file
		}
	}
	return largestFile
}

func (client *Client) download() {
	t := client.Torrent
	t.DownloadAll()
	// Set priorities of file (5% ahead)
	for {
		largestFile := client.getLargestFile()
		firstPieceIndex := largestFile.Offset() * int64(t.NumPieces()) / t.Length()
		endPieceIndex := (largestFile.Offset() + largestFile.Length()) * int64(t.NumPieces()) / t.Length()
		for idx := firstPieceIndex; idx <= endPieceIndex*10/100; idx++ {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
		}
	}
}

// Start starts the client by getting the torrent information and allocating the priorities of each piece.
func (client *Client) Start() {
	<-client.Torrent.GotInfo() // blocks until it got the info
	go client.download()       // download file
}

func (client *Client) streamHandler(w http.ResponseWriter, r *http.Request) {
	file := client.getLargestFile()
	entry, err := NewFileReader(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+file.DisplayPath()+"\"")
	http.ServeContent(w, r, file.DisplayPath(), time.Now(), entry)
}

// Serve serves the torrent via HTTP localhost:{port}.
func (client *Client) Serve() {
	p := strconv.Itoa(client.HostPort)
	client.URL = "http://localhost:" + p
	go func() {
		http.HandleFunc("/", client.streamHandler)
		logrus.Fatalln(http.ListenAndServe(":"+p, nil))
	}()
}

// PrintProgress prints out the current download progress of the client for the CLI.
func (client *Client) PrintProgress() {
	t := client.Torrent
	if t.Info() == nil {
		return
	}
	total := t.Length()
	currentProgress := t.BytesCompleted()
	complete := humanize.Bytes(uint64(currentProgress))
	size := humanize.Bytes(uint64(total))
	percentage := float64(currentProgress) / float64(total) * 100
	output := bufio.NewWriter(os.Stdout)
	fmt.Fprintf(output, "Progress: %s / %s  %.2f%%\r", complete, size, percentage)
	// TODO: print download speed
	output.Flush()
}

// Close cleans up the connections of the client.
func (client *Client) Close() {
	client.Torrent.Drop()
	client.Client.Close()
}

// SeekableContent describes an io.ReadSeeker that can be closed as well.
type SeekableContent interface {
	io.ReadSeeker
	io.Closer
}

// FileEntry helps reading a torrent file.
type FileEntry struct {
	*torrent.File
	torrent.Reader
}

// Seek seeks to the correct file position, paying attention to the offset.
func (f *FileEntry) Seek(offset int64, whence int) (int64, error) {
	return f.Reader.Seek(offset+f.File.Offset(), whence)
}

// NewFileReader sets up a torrent file for streaming reading.
func NewFileReader(f *torrent.File) (SeekableContent, error) {
	t := f.Torrent()
	reader := t.NewReader()

	// We read ahead 1% of the file continuously.
	reader.SetReadahead(f.Length() / 100)
	reader.SetResponsive()
	_, err := reader.Seek(f.Offset(), io.SeekStart)

	return &FileEntry{File: f, Reader: reader}, err
}
