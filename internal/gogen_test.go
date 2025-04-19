package internal

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockGogenServer creates a test server that handles /v1/list, /v1/search, /v1/get, and /v1/upsert endpoints
func mockGogenServer(t *testing.T) (*httptest.Server, []GogenList) {
	// Define the expected JSON response for list/search
	mockResponse := `{"Items": [{"gogen": "coccyx/test3"}, {"description": "Generate CPU Metrics", "gogen": "coccyx/cpu"}, {"description": "Generate Disk Usage Metrics", "gogen": "coccyx/df"}, {"description": "Example of a custom lua generator", "gogen": "coccyx/users"}, {"gogen": "coccyx/test"}, {"description": "Simple CSV Example", "gogen": "coccyx/csv"}, {"description": "Weblog data in Apache common format", "gogen": "coccyx/weblog"}, {"description": "Generate CPU Metrics", "gogen": "coccyx/nixOS"}, {"description": "Example business event log from a middleware system, in key=value format.", "gogen": "coccyx/businessevent"}, {"description": "Generate Bandwidth Usage Metrics", "gogen": "coccyx/bandwidth"}, {"gogen": "coccyx/test2"}, {"description": "Generate Iostat Usage Metrics", "gogen": "coccyx/iostat"}, {"description": "Generate Memory Usage Metrics", "gogen": "coccyx/vmstat"}, {"description": "Tutorial 3", "gogen": "coccyx/tutorial3"}]}`

	// Define the expected JSON response for get
	mockGetResponse := `{"Item": {"owner": "coccyx", "sampleEvent": "{\"_raw\":\"2017-01-31 14:38:00.440 transType=Change transID=1 transGUID=048a6346-6671-4703-89ec-3843afd44e85 userName=ivettaadelima city=\\\"HARTFORD\\\" state=CT zip=6101 value=1.206\",\"host\":\"server2.gogen.io\",\"index\":\"main\",\"source\":\"/var/log/translog\",\"sourcetype\":\"translog\"}\n", "gogen": "coccyx/tutorial3", "name": "tutorial3", "description": "Tutorial 3", "version": 3.0, "gistID": "0e3c9fda88915239b21d8a85a837750c", "config": "global:\n  rotInterval: 1\n  samplesDir:\n  - .\nsamples:\n- name: tutorial3\n  description: Tutorial 3\n  disabled: false\n  generator: sample\n  rater: eventrater\n  interval: 60\n  count: 2\n  earliest: now\n  latest: now\n  begin: 2012-02-09T08:00:00Z\n  end: 2012-02-09T08:03:00Z\n  tokens:\n  - name: ts\n    format: regex\n    token: (\\d{4}-\\d{2}-\\d{2}\\s+\\d{2}:\\d{2}:\\d{2},\\d{3})\n    type: gotimestamp\n    replacement: 2006-01-02 15:04:05.000\n    field: _raw\n  - name: host\n    format: template\n    token: $host$\n    type: choice\n    field: host\n    choice:\n    - server1.gogen.io\n    - server2.gogen.io\n  - name: transtype\n    format: regex\n    token: transType=(\\w+)\n    type: weightedChoice\n    field: _raw\n    weightedChoice:\n    - weight: 3\n      choice: New\n    - weight: 5\n      choice: Change\n    - weight: 1\n      choice: Delete\n  - name: integerid\n    format: template\n    token: $integerid$\n    type: script\n    field: _raw\n    script: |\n      state[\"id\"] = state[\"id\"] + 1 return state[\"id\"]\n    init:\n      id: \"0\"\n  - name: guid\n    format: template\n    token: $guid$\n    type: random\n    replacement: guid\n    field: _raw\n  - name: username\n    format: template\n    token: $username$\n    type: choice\n    sample: usernames.sample\n    field: _raw\n    choice:\n    - birodivulga162\n    - nildajcbonanno\n    - ivettaadelima\n    - pckomono\n    - Looreeto\n    - JooPedro1591\n    - claaarecurlingg\n    - acciokcavote\n    - JungD\n    - InaraAllves\n    - Haroldmcaol\n    - xNessaa\n    - stylesdofunk\n    - meltemmeltemm\n    - emapujig\n    - cellphones4deal\n    - amisisuvi\n    - MegSeecharran95\n    - MargueritaYociu\n    - MarcioBFasano\n  - name: markets-city\n    format: template\n    token: $city$\n    type: fieldChoice\n    group: 1\n    sample: markets.csv\n    field: _raw\n    srcField: city\n    fieldChoice:\n    - city: SPRINGFIELD\n      county: HAMPDEN\n      lat: \"42.106\"\n      long: \"-72.5977\"\n      state: MA\n      zip: \"1101\"\n    - city: WORCESTER\n      county: WORCESTER\n      lat: \"42.2621\"\n      long: \"-71.8034\"\n      state: MA\n      zip: \"1601\"\n    - city: WOBURN\n      county: MIDDLESEX\n      lat: \"42.482894\"\n      long: \"-71.157404\"\n      state: MA\n      zip: \"1801\"\n    - city: BOSTON\n      county: SUFFOLK\n      lat: \"42.345\"\n      long: \"-71.0876\"\n      state: MA\n      zip: \"2123\"\n    - city: MANCHESTER\n      county: HILLSBOROUGH\n      lat: \"42.992858\"\n      long: \"-71.463255\"\n      state: NH\n      zip: \"3101\"\n    - city: PORTLAND\n      county: CUMBERLAND\n      lat: \"43.660564\"\n      long: \"-70.258864\"\n      state: ME\n      zip: \"4101\"\n    - city: MONTPELIER\n      county: WASHINGTON\n      lat: \"44.2574\"\n      long: \"-72.5698\"\n      state: VT\n      zip: \"5601\"\n    - city: HARTFORD\n      county: HARTFORD\n      lat: \"41.7636\"\n      long: \"-72.6855\"\n      state: CT\n      zip: \"6101\"\n    - city: WEST HARTFORD\n      county: HARTFORD\n      lat: \"41.755553\"\n      long: \"-72.75322\"\n      state: CT\n      zip: \"6107\"\n  - name: markets-state\n    format: template\n    token: $state$\n    type: fieldChoice\n    group: 1\n    sample: markets.csv\n    field: _raw\n    srcField: state\n    fieldChoice:\n    - city: SPRINGFIELD\n      county: HAMPDEN\n      lat: \"42.106\"\n      long: \"-72.5977\"\n      state: MA\n      zip: \"1101\"\n    - city: WORCESTER\n      county: WORCESTER\n      lat: \"42.2621\"\n      long: \"-71.8034\"\n      state: MA\n      zip: \"1601\"\n    - city: WOBURN\n      county: MIDDLESEX\n      lat: \"42.482894\"\n      long: \"-71.157404\"\n      state: MA\n      zip: \"1801\"\n    - city: BOSTON\n      county: SUFFOLK\n      lat: \"42.345\"\n      long: \"-71.0876\"\n      state: MA\n      zip: \"2123\"\n    - city: MANCHESTER\n      county: HILLSBOROUGH\n      lat: \"42.992858\"\n      long: \"-71.463255\"\n      state: NH\n      zip: \"3101\"\n    - city: PORTLAND\n      county: CUMBERLAND\n      lat: \"43.660564\"\n      long: \"-70.258864\"\n      state: ME\n      zip: \"4101\"\n    - city: MONTPELIER\n      county: WASHINGTON\n      lat: \"44.2574\"\n      long: \"-72.5698\"\n      state: VT\n      zip: \"5601\"\n    - city: HARTFORD\n      county: HARTFORD\n      lat: \"41.7636\"\n      long: \"-72.6855\"\n      state: CT\n      zip: \"6101\"\n    - city: WEST HARTFORD\n      county: HARTFORD\n      lat: \"41.755553\"\n      long: \"-72.75322\"\n      state: CT\n      zip: \"6107\"\n  - name: markets-zip\n    format: template\n    token: $zip$\n    type: fieldChoice\n    group: 1\n    sample: markets.csv\n    field: _raw\n    srcField: zip\n    fieldChoice:\n    - city: SPRINGFIELD\n      county: HAMPDEN\n      lat: \"42.106\"\n      long: \"-72.5977\"\n      state: MA\n      zip: \"1101\"\n    - city: WORCESTER\n      county: WORCESTER\n      lat: \"42.2621\"\n      long: \"-71.8034\"\n      state: MA\n      zip: \"1601\"\n    - city: WOBURN\n      county: MIDDLESEX\n      lat: \"42.482894\"\n      long: \"-71.157404\"\n      state: MA\n      zip: \"1801\"\n    - city: BOSTON\n      county: SUFFOLK\n      lat: \"42.345\"\n      long: \"-71.0876\"\n      state: MA\n      zip: \"2123\"\n    - city: MANCHESTER\n      county: HILLSBOROUGH\n      lat: \"42.992858\"\n      long: \"-71.463255\"\n      state: NH\n      zip: \"3101\"\n    - city: PORTLAND\n      county: CUMBERLAND\n      lat: \"43.660564\"\n      long: \"-70.258864\"\n      state: ME\n      zip: \"4101\"\n    - city: MONTPELIER\n      county: WASHINGTON\n      lat: \"44.2574\"\n      long: \"-72.5698\"\n      state: VT\n      zip: \"5601\"\n    - city: HARTFORD\n      county: HARTFORD\n      lat: \"41.7636\"\n      long: \"-72.6855\"\n      state: CT\n      zip: \"6101\"\n    - city: WEST HARTFORD\n      county: HARTFORD\n      lat: \"41.755553\"\n      long: \"-72.75322\"\n      state: CT\n      zip: \"6107\"\n  - name: value\n    format: regex\n    token: value=(\\d+)\n    type: random\n    replacement: float\n    field: _raw\n    precision: 3\n    upper: 10\n  lines:\n  - _raw: 2012-09-14 16:30:20,072 transType=ReplaceMe transID=$integerid$ transGUID=$guid$\n      userName=$username$ city=\"$city$\" state=$state$ zip=$zip$ value=0\n    host: $host$\n    index: main\n    source: /var/log/translog\n    sourcetype: translog\n  field: _raw\n  singlepass: true\nmix: []\nraters:\n- name: eventrater\n  type: config\n  options:\n    MinuteOfHour:\n      0: 1\n      1: 0.5\n      2: 2\n"}, "ResponseMetadata": {"RequestId": "a4e26206-1048-43a3-b5df-3e0d3f6989fa", "HTTPStatusCode": 200, "HTTPHeaders": {"server": "Jetty(12.0.14)", "date": "Thu, 06 Mar 2025 21:18:19 GMT", "x-amzn-requestid": "a4e26206-1048-43a3-b5df-3e0d3f6989fa", "content-type": "application/x-amz-json-1.0", "x-amz-crc32": "337133023", "content-length": "516"}, "RetryAttempts": 0}}`

	// Define the expected JSON response for upsert
	mockUpsertResponse := `{"ResponseMetadata": {"RequestId": "99e49103-3b6e-4e8d-a01d-c8163e9f50c4", "HTTPStatusCode": 200, "HTTPHeaders": {"server": "Jetty(12.0.14)", "date": "Thu, 06 Mar 2025 21:27:11 GMT", "x-amzn-requestid": "99e49103-3b6e-4e8d-a01d-c8163e9f50c4", "content-type": "application/x-amz-json-1.0", "x-amz-crc32": "2745614147", "content-length": "2"}, "RetryAttempts": 0}}`

	// Parse the expected response to create our expected GogenList items
	var expectedResponse map[string]interface{}
	err := json.Unmarshal([]byte(mockResponse), &expectedResponse)
	assert.NoError(t, err, "Failed to parse expected JSON response")

	// Extract the expected items that should be in the result
	expectedItems := expectedResponse["Items"].([]interface{})

	// Create a list of expected GogenList items that should be returned
	// Only include items that have both gogen and description fields
	var expectedList []GogenList
	for _, item := range expectedItems {
		itemMap := item.(map[string]interface{})
		gogen, hasGogen := itemMap["gogen"].(string)
		description, hasDescription := itemMap["description"].(string)

		if hasGogen && hasDescription {
			expectedList = append(expectedList, GogenList{
				Gogen:       gogen,
				Description: description,
			})
		}
	}

	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the content type
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")

		// Handle different endpoints
		if r.URL.Path == "/v1/list" {
			// Return the full mock response for list endpoint
			w.Write([]byte(mockResponse))
			return
		} else if r.URL.Path == "/v1/search" {
			// Get the search query parameter
			query := r.URL.Query().Get("q")
			if query == "" {
				http.Error(w, "Missing query parameter", http.StatusBadRequest)
				return
			}

			// Filter the items based on the query
			var filteredItems []map[string]interface{}
			for _, item := range expectedItems {
				itemMap := item.(map[string]interface{})
				gogen, hasGogen := itemMap["gogen"].(string)

				// Include the item if its gogen field contains the query string
				if hasGogen && strings.Contains(strings.ToLower(gogen), strings.ToLower(query)) {
					filteredItems = append(filteredItems, itemMap)
				}
			}

			// Create a filtered response
			filteredResponse := map[string]interface{}{
				"Items": filteredItems,
			}

			// Convert to JSON and return
			filteredJSON, err := json.Marshal(filteredResponse)
			if err != nil {
				http.Error(w, "Error creating response", http.StatusInternalServerError)
				return
			}

			w.Write(filteredJSON)
			return
		} else if strings.HasPrefix(r.URL.Path, "/v1/get/") {
			// Extract the gogen identifier from the path
			gogenID := strings.TrimPrefix(r.URL.Path, "/v1/get/")

			// Check if the requested gogen is the one we have a mock response for
			if gogenID == "coccyx/tutorial3" {
				w.Write([]byte(mockGetResponse))
				return
			} else {
				// Return a 404 for any other gogen ID
				http.Error(w, "Gogen not found", http.StatusNotFound)
				return
			}
		} else if r.URL.Path == "/v1/upsert" {
			// Check if the request method is POST
			if r.Method != "POST" {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Check if the request has an Authorization header
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "token ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Read the request body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}

			// Parse the request body to verify it's valid JSON
			var gogen GogenInfo
			err = json.Unmarshal(body, &gogen)
			if err != nil {
				http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
				return
			}

			// Verify that the required fields are present
			if gogen.Gogen == "" || gogen.Owner == "" || gogen.Name == "" {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
				return
			}

			// Return the mock upsert response
			w.Write([]byte(mockUpsertResponse))
			return
		} else {
			// Handle unknown paths
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}))

	return mockServer, expectedList
}

func TestListWithMockServer(t *testing.T) {
	// Create the mock server
	mockServer, expectedList := mockGogenServer(t)
	defer mockServer.Close()

	// Set the GOGEN_APIURL environment variable to point to our mock server
	originalAPIURL := os.Getenv("GOGEN_APIURL")
	os.Setenv("GOGEN_APIURL", mockServer.URL)
	defer os.Setenv("GOGEN_APIURL", originalAPIURL) // Restore the original value when done

	// Call the List function
	result := List()

	// Verify the results
	assert.NotEmpty(t, result, "List result should not be empty")

	// Check that we have the correct number of items
	assert.Equal(t, len(expectedList), len(result), "Number of items should match")

	// Create maps for easier comparison
	resultMap := make(map[string]string)
	for _, item := range result {
		resultMap[item.Gogen] = item.Description
	}

	expectedMap := make(map[string]string)
	for _, item := range expectedList {
		expectedMap[item.Gogen] = item.Description
	}

	// Verify that the maps are identical
	assert.Equal(t, expectedMap, resultMap, "The result should exactly match the expected items")

	// Verify each individual item for more detailed error messages if there's a mismatch
	for _, expected := range expectedList {
		found := false
		for _, actual := range result {
			if expected.Gogen == actual.Gogen {
				found = true
				assert.Equal(t, expected.Description, actual.Description,
					"Description mismatch for %s. Expected: %s, Got: %s",
					expected.Gogen, expected.Description, actual.Description)
				break
			}
		}
		assert.True(t, found, "Expected item %s not found in result", expected.Gogen)
	}
}

func TestSearchWithMockServer(t *testing.T) {
	// Create the mock server
	mockServer, allExpectedItems := mockGogenServer(t)
	defer mockServer.Close()

	// Set the GOGEN_APIURL environment variable to point to our mock server
	originalAPIURL := os.Getenv("GOGEN_APIURL")
	os.Setenv("GOGEN_APIURL", mockServer.URL)
	defer os.Setenv("GOGEN_APIURL", originalAPIURL) // Restore the original value when done

	// Define search queries and expected results
	testCases := []struct {
		query    string
		expected []GogenList
	}{
		{
			query:    "cpu",
			expected: filterExpectedItems(allExpectedItems, "cpu"),
		},
		{
			query:    "weblog",
			expected: filterExpectedItems(allExpectedItems, "weblog"),
		},
		{
			query:    "coccyx",
			expected: allExpectedItems, // All items contain "coccyx"
		},
	}

	for _, tc := range testCases {
		t.Run("Search_"+tc.query, func(t *testing.T) {
			// Call the Search function
			result := Search(tc.query)

			// Verify the results
			assert.NotEmpty(t, result, "Search result should not be empty for query: %s", tc.query)

			// Check that we have the correct number of items
			assert.Equal(t, len(tc.expected), len(result),
				"Number of items should match for query: %s. Expected: %d, Got: %d",
				tc.query, len(tc.expected), len(result))

			// Create maps for easier comparison
			resultMap := make(map[string]string)
			for _, item := range result {
				resultMap[item.Gogen] = item.Description
			}

			expectedMap := make(map[string]string)
			for _, item := range tc.expected {
				expectedMap[item.Gogen] = item.Description
			}

			// Verify that the maps are identical
			assert.Equal(t, expectedMap, resultMap,
				"The result should exactly match the expected items for query: %s", tc.query)
		})
	}
}

func TestGetWithMockServer(t *testing.T) {
	// Create the mock server
	mockServer, _ := mockGogenServer(t)
	defer mockServer.Close()

	// Set the GOGEN_APIURL environment variable to point to our mock server
	originalAPIURL := os.Getenv("GOGEN_APIURL")
	os.Setenv("GOGEN_APIURL", mockServer.URL)
	defer os.Setenv("GOGEN_APIURL", originalAPIURL) // Restore the original value when done

	// Call the Get function with the ID that should return a valid response
	result, err := Get("coccyx/tutorial3")

	// Verify there was no error
	assert.NoError(t, err, "Get should not return an error for a valid Gogen ID")

	// Verify the result fields match what we expect
	assert.Equal(t, "coccyx/tutorial3", result.Gogen, "Gogen field should match")
	assert.Equal(t, "coccyx", result.Owner, "Owner field should match")
	assert.Equal(t, "tutorial3", result.Name, "Name field should match")
	assert.Equal(t, "Tutorial 3", result.Description, "Description field should match")
	assert.Equal(t, "0e3c9fda88915239b21d8a85a837750c", result.GistID, "GistID field should match")
	assert.Equal(t, 3, result.Version, "Version field should match")
	assert.True(t, len(result.Config) > 0, "Config field should not be empty")
	assert.True(t, len(result.SampleEvent) > 0, "SampleEvent field should not be empty")

	// Test with an invalid Gogen ID
	_, err = Get("coccyx/nonexistent")
	assert.Error(t, err, "Get should return an error for an invalid Gogen ID")
	assert.Contains(t, err.Error(), "Could not find Gogen", "Error message should indicate Gogen not found")
}

// Mock the GitHub struct for testing Upsert
type mockGitHub struct {
	token string
}

// Store the original NewGitHub function
var originalNewGitHub func(bool) *GitHub

func init() {
	// Save the original function
	originalNewGitHub = NewGitHub
}

func TestUpsertWithMockServer(t *testing.T) {
	// Create the mock server
	mockServer, _ := mockGogenServer(t)
	defer mockServer.Close()

	// Set the GOGEN_APIURL environment variable to point to our mock server
	originalAPIURL := os.Getenv("GOGEN_APIURL")
	os.Setenv("GOGEN_APIURL", mockServer.URL)
	defer os.Setenv("GOGEN_APIURL", originalAPIURL) // Restore the original value when done

	// Create a GogenInfo object to upsert
	gogen := GogenInfo{
		Gogen:       "coccyx/tutorial1",
		Owner:       "coccyx",
		Name:        "tutorial1",
		Description: "Tutorial 1",
		Notes:       "",
		SampleEvent: "{\"_raw\":\"Mar/06/25 13:24:28 line3\"}\n",
		GistID:      "9edab9605421c036ce2ebef5b4966b1d",
		Version:     1,
		Config:      "",
	}

	// Call the Upsert function
	upsert(gogen, &GitHub{token: "mock-github-token"})

	// Since Upsert doesn't return anything, we can only verify that it didn't panic
	// The mock server will return an error if the request is not formatted correctly
}

// Helper function to filter expected items based on a query string
func filterExpectedItems(items []GogenList, query string) []GogenList {
	var filtered []GogenList
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Gogen), strings.ToLower(query)) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
