package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/coccyx/gogen/logger"
	"github.com/kr/pretty"
)

// GogenInfo represents a remote object from our service which stores shared Gogens
type GogenInfo struct {
	Gogen       string `json:"gogen"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	SampleEvent string `json:"sampleEvent"`
	GistID      string `json:"gistID"`
	Version     int    `json:"version"`
	Config      string `json:"config"`
}

// GogenList is returned by the /v1/list and /v1/search APIs for Gogen
type GogenList struct {
	Gogen       string
	Description string
}

// defaultAPIClient is the shared HTTP client for API calls with a reasonable timeout.
var defaultAPIClient = &http.Client{Timeout: 30 * time.Second}

// doHTTPRequest executes an HTTP request, reads the response body, and returns
// the body bytes. It properly closes resp.Body and returns an *HTTPError for
// non-2xx status codes.
func doHTTPRequest(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to %s failed: %w", req.URL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from %s: %w", req.URL, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			URL:        req.URL.String(),
			Body:       string(body),
		}
	}
	return body, nil
}

// doGet performs an HTTP GET request to the given URL using the default API client.
func doGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", url, err)
	}
	return doHTTPRequest(defaultAPIClient, req)
}

// doPost performs an HTTP POST request to the given URL using the default API client.
func doPost(url string, body io.Reader, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", url, err)
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return doHTTPRequest(defaultAPIClient, req)
}

// List calls /v1/list
func List() ([]GogenList, error) {
	return listsearch(fmt.Sprintf("%s/v1/list", getAPIURL()))
}

// Search calls /v1/search
func Search(q string) ([]GogenList, error) {
	return listsearch(fmt.Sprintf("%s/v1/search?q=%s", getAPIURL(), url.QueryEscape(q)))
}

func listsearch(url string) ([]GogenList, error) {
	body, err := doGet(url)
	if err != nil {
		return nil, fmt.Errorf("error retrieving list of Gogens: %w", err)
	}
	var list map[string]interface{}
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling list response: %w", err)
	}
	items := list["Items"].([]interface{})
	var ret []GogenList
	for _, item := range items {
		tempitem := item.(map[string]interface{})
		if _, ok := tempitem["gogen"]; !ok {
			continue
		}
		if _, ok := tempitem["description"]; !ok {
			continue
		}
		li := GogenList{Gogen: tempitem["gogen"].(string), Description: tempitem["description"].(string)}
		ret = append(ret, li)
	}
	log.Debugf("List: %# v", pretty.Formatter(ret))
	return ret, nil
}

// Get calls /v1/get
var Get = func(q string) (g GogenInfo, err error) {
	url := fmt.Sprintf("%s/v1/get/%s", getAPIURL(), q)
	log.Debugf("Calling %s", url)
	body, err := doGet(url)
	if err != nil {
		var httpErr *HTTPError
		if errors.As(err, &httpErr) && httpErr.IsNotFound() {
			return g, fmt.Errorf("could not find Gogen %s: %w", q, err)
		}
		return g, fmt.Errorf("error retrieving Gogen %s: %w", q, err)
	}
	var gogen map[string]interface{}
	err = json.Unmarshal(body, &gogen)
	if err != nil {
		return g, fmt.Errorf("error unmarshaling body: %w", err)
	}
	tmp, err := json.Marshal(gogen["Item"])
	if err != nil {
		return g, fmt.Errorf("error converting Item to JSON: %w", err)
	}
	err = json.Unmarshal(tmp, &g)
	if err != nil {
		return g, fmt.Errorf("error unmarshaling item: %w", err)
	}
	gCopy := g
	gCopy.Config = "redacted"
	log.Debugf("Gogen: %# v", pretty.Formatter(gCopy))
	return g, nil
}

// Upsert calls /v1/upsert
func Upsert(g GogenInfo) error {
	gh := NewGitHub(true)
	return upsert(g, gh)
}

func upsert(g GogenInfo, gh *GitHub) error {
	b, err := json.Marshal(g)
	if err != nil {
		return fmt.Errorf("error marshaling Gogen %#v: %w", g, err)
	}

	headers := map[string]string{
		"Authorization": "token " + gh.token,
	}
	_, err = doPost(fmt.Sprintf("%s/v1/upsert", getAPIURL()), bytes.NewReader(b), headers)
	if err != nil {
		return fmt.Errorf("error upserting Gogen: %w", err)
	}
	log.Debugf("Upserted: %# v", pretty.Formatter(g))
	return nil
}

// getAPIURL returns the API URL from environment variable or default value
func getAPIURL() string {
	if url := os.Getenv("GOGEN_APIURL"); url != "" {
		return url
	}
	return "https://api.gogen.io"
}
