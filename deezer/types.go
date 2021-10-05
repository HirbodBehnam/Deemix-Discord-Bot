package deezer

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// SearchedTrack is the result of a search
type SearchedTrack struct {
	// It contains the basic info of a Track
	Track
	// The album name
	Album string
	// The duration of music
	Duration time.Duration
}

func (t SearchedTrack) Append(builder *strings.Builder) {
	builder.WriteString("\nTitle: ")
	builder.WriteString(t.Title)
	builder.WriteString("\nAlbum: ")
	builder.WriteString(t.Album)
	builder.WriteString("\nArtist: ")
	builder.WriteString(t.Artist)
	builder.WriteString("\nDuration: ")
	builder.WriteString(t.Duration.String())
	builder.WriteString("\nLink:\n`")
	builder.WriteString(t.Link)
	builder.WriteString("`\n\n")
}

type trackSearchResponse struct {
	Data []trackInfoResponse `json:"data"`
}

type trackInfoResponse struct {
	Title    string `json:"title"`
	Link     string `json:"link"`
	Duration int    `json:"duration"`
	Artist   struct {
		Name string `json:"name"`
	} `json:"artist"`
	Album struct {
		Title string `json:"title"`
	} `json:"album"`
}

// Track converts trackInfoResponse to Track
func (t trackInfoResponse) Track() Track {
	return Track{
		Title:  t.Title,
		Link:   t.Link,
		Artist: t.Artist.Name,
	}
}

// SearchedTrack converts trackInfoResponse to SearchedTrack
func (t trackInfoResponse) SearchedTrack() SearchedTrack {
	return SearchedTrack{
		Track:    t.Track(),
		Album:    t.Album.Title,
		Duration: time.Second * time.Duration(t.Duration),
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
