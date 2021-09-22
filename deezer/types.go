package deezer

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Track is the entry of track searches
type Track struct {
	// The title (name) of the song
	Title string
	// The artist name
	Artist string
	// The link to the song
	Link string
}

func (t Track) String() string {
	return t.Artist + " - " + t.Title
}

type trackSearchResponse struct {
	Data []trackInfoResponse `json:"data"`
}

type trackInfoResponse struct {
	Title  string `json:"title"`
	Link   string `json:"link"`
	Artist struct {
		Name string `json:"name"`
	} `json:"artist"`
}

// Track converts trackInfoResponse to Track
func (t trackInfoResponse) Track() Track {
	return Track{
		Title:  t.Title,
		Link:   t.Link,
		Artist: t.Artist.Name,
	}
}

// TempDir is a simple structure which can hold the path to a temporary directory
type TempDir struct {
	// Address of the directory
	Address string
}

// Delete deletes the temporary directory
func (d TempDir) Delete() {
	if d.Address != "" {
		_ = os.RemoveAll(d.Address)
	}
}

// GetMusics gets the downloaded music filenames from temp dir
// If there is an error, returns nil
func (d TempDir) GetMusics() []string {
	result := make([]string, 0)
	err := filepath.WalkDir(d.Address, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".mp3") {
			result = append(result, path)
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return result
}
