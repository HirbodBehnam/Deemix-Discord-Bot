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
	// The link to the song
	Link string
}

type trackSearchResponse struct {
	Data []struct {
		Title string `json:"title"`
		Link  string `json:"link"`
	} `json:"data"`
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
