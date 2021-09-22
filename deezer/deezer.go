package deezer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// httpClient is the client to do the requests with it
var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// trackPathRegex is used to extract the track ID from path of deezer
var trackPathRegex = regexp.MustCompile("/track/(\\d+)")

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
		result = append(result, entry.Track())
	}
	return result, nil
}

// GetTrack gets a single track's info by its track ID
func GetTrack(trackID int) (Track, error) {
	resp, err := httpClient.Get("https://api.deezer.com/track/" + strconv.Itoa(trackID))
	if err != nil {
		return Track{}, err
	}
	var result trackInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	_ = resp.Body.Close()
	return result.Track(), err
}

// KeywordToLink at firsts checks if the text is a link or not
// If it's a link, it will return the text itself
// Otherwise it searches deezer for the text and returns the first result's Track
func KeywordToLink(text string) (track Track, err error) {
	// If the text is url just return it
	u, err := url.Parse(text)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return trackFromUrl(u)
	}
	// Otherwise, search deezer
	tracks, _ := SearchTrack(text)
	if len(tracks) == 0 {
		return Track{}, errors.New("track not found")
	}
	return tracks[0], nil
}

// trackFromUrl tries to get a Track from url
func trackFromUrl(u *url.URL) (track Track, err error) {
	if u.Host == "deezer.page.link" {
		// This is a readwrite page. Just open it and follow the redirection
		resp, err := httpClient.Head(u.String())
		if err != nil {
			log.Println("cannot head the page with url", u.String(), ":", err)
			return Track{}, errors.New("cannot load page data")
		}
		_ = resp.Body.Close()
		u, err = url.Parse(resp.Header.Get("location"))
		if err != nil {
			return Track{}, errors.New("cannot parse the url after redirect")
		}
	}
	if u.Host != "www.deezer.com" {
		return Track{}, errors.New("invalid url")
	}
	// Extract the track ID
	matches := trackPathRegex.FindStringSubmatch(u.Path)
	// Check if the url is a track link
	if len(matches) != 2 {
		return Track{}, errors.New("invalid url")
	}
	trackID, err := strconv.Atoi(matches[1])
	if err != nil {
		return Track{}, errors.New("invalid url")
	}
	// Now get the track from track ID
	return GetTrack(trackID)
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
