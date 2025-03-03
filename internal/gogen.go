package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

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
}

// GogenList is returned by the /v1/list and /v1/search APIs for Gogen
type GogenList struct {
	Gogen       string
	Description string
}

// List calls /v1/list
func List() []GogenList {
	return listsearch(fmt.Sprintf("%s/v1/list", getAPIURL()))
}

// Search calls /v1/search
func Search(q string) []GogenList {
	return listsearch(fmt.Sprintf("%s/v1/search?q=%s", getAPIURL(), url.QueryEscape(q)))
}

func listsearch(url string) (ret []GogenList) {
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp.StatusCode != 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			log.Fatalf("Non 200 response code searching for Gogen: %s", string(body))
		} else {
			log.Fatalf("Error retrieving list of Gogens: %s", err)
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading body from response: %s", err)
	}
	var list map[string]interface{}
	err = json.Unmarshal(body, &list)
	// log.Debugf("List body: %s", string(body))
	// log.Debugf("list: %s", fmt.Sprintf("%# v", pretty.Formatter(list)))
	items := list["Items"].([]interface{})
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
	return ret
}

// Get calls /v1/get
func Get(q string) (g GogenInfo, err error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/v1/get/%s", getAPIURL(), q)
	log.Debugf("Calling %s", url)
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			if resp.StatusCode == 404 {
				return g, fmt.Errorf("Could not find Gogen: %s\n", q)
			}
			if resp.StatusCode != 200 {
				body, _ := ioutil.ReadAll(resp.Body)
				return g, fmt.Errorf("Non 200 response code retrieving Gogen: %s", string(body))
			}
		} else {
			return g, fmt.Errorf("Error retrieving Gogen %s: %s", q, err)
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return g, fmt.Errorf("Error reading body from response: %s", err)
	}
	// log.Debugf("Body: %s", body)
	var gogen map[string]interface{}
	err = json.Unmarshal(body, &gogen)
	if err != nil {
		return g, fmt.Errorf("Error unmarshaling body: %s", err)
	}
	tmp, err := json.Marshal(gogen["Item"])
	if err != nil {
		return g, fmt.Errorf("Error converting Item to JSON: %s", err)
	}
	log.Debugf("tmp: %s", string(tmp))
	err = json.Unmarshal(tmp, &g)
	if err != nil {
		return g, fmt.Errorf("Error unmarshaling item: %s", err)
	}
	log.Debugf("Gogen: %# v", pretty.Formatter(g))
	return g, nil
}

// Upsert calls /v1/upsert
func Upsert(g GogenInfo) {
	gh := NewGitHub(true)
	client := &http.Client{}

	b, err := json.Marshal(g)
	if err != nil {
		log.Fatalf("Error marshaling Gogen %#v: %s", g, err)
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/upsert", getAPIURL()), bytes.NewReader(b))
	req.Header.Add("Authorization", "token "+gh.token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp.StatusCode != 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			log.Fatalf("Non 200 response code Upserting: %s", string(body))
		} else {
			log.Fatalf("Error POSTing to upsert: %s", err)
		}
	}
	log.Debugf("Upserted: %# v", pretty.Formatter(g))
}

// getAPIURL returns the API URL from environment variable or default value
func getAPIURL() string {
	if url := os.Getenv("GOGEN_APIURL"); url != "" {
		return url
	}
	return "https://api.gogen.io"
}
