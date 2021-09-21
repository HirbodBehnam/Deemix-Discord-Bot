package deezer

import (
	"Deemix-Discord-Bot/util"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"
)

// httpClient is the client to do the requests with it
var httpClient = &http.Client{Timeout: 5 * time.Second}

// trackSearchEndpoint is where we should send our search requests for tracks
const trackSearchEndpoint = "https://api.deezer.com/search"

// maxSearchEntries is the maximum number of searches in response
const maxSearchEntries = 5

// SearchTrack searches the deezer for a track by keyword
func SearchTrack(keyword string) ([]Track, error) {
	// Build the request and do it
	req, err := http.NewRequest("GET", trackSearchEndpoint, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("q", keyword)
	req.URL.RawQuery = q.Encode()
	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	var respRaw trackSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&respRaw)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	// Convert the raw response to SearchResult array
	result := make([]Track, 0, maxSearchEntries)
	for i, entry := range respRaw.Data {
		if i >= maxSearchEntries { // limit entries of result
			break
		}
		result = append(result, Track{
			Title: entry.Title,
			Link:  entry.Link,
		})
	}
	return result, nil
}

// KeywordToLink at firsts checks if the text is a link or not
// If it's a link, it will return the text itself
// Otherwise it searches deezer for the text and returns the first result's Track
func KeywordToLink(text string) (track Track, ok bool) {
	// If the text is url just return it
	if util.IsUrl(text) {
		return Track{
			Title: "your requested song",
			Link:  text,
		}, true
	}
	// Otherwise, search deezer
	tracks, _ := SearchTrack(text)
	if len(tracks) == 0 {
		return Track{}, false
	}
	return tracks[0], true
}

// Download tries to download a spotify/deezer track from deezer
// We return a pointer to ensure that user don't recklessly call TempDir.Delete on result
func Download(u string, cancelChannel <-chan struct{}) (*TempDir, error) {
	// Create a temp dir
	dirName, err := ioutil.TempDir("", "deemix*")
	if err != nil {
		return nil, err
	}
	result := &TempDir{Address: dirName}
	// Download the file
	cmd := exec.Command("deemix", "-p", dirName, "-b", "128", u)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Start()
	if err != nil {
		log.Printf("Error on excuting deemix: %s\n", stderr.String())
		result.Delete()
		return nil, err
	}
	// Wait either for the deemix to finish or kill it
	doneChannel := make(chan struct{}, 1)
	go func() {
		select {
		case <-doneChannel:
		case <-cancelChannel:
			_ = cmd.Process.Kill()
		}
	}()
	err = cmd.Wait()
	doneChannel <- struct{}{} // Don't wait for cancel anymore
	if err != nil {
		log.Printf("Error on excuting deemix: %s\n", stderr.String())
		result.Delete()
		return nil, err
	}
	// Return the directory
	return result, nil
}
