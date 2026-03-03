package internal

import "fmt"

// HTTPError represents an HTTP response with a non-2xx status code.
type HTTPError struct {
	StatusCode int
	URL        string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d from %s: %s", e.StatusCode, e.URL, e.Body)
}

// IsNotFound returns true if the HTTP status code is 404.
func (e *HTTPError) IsNotFound() bool {
	return e.StatusCode == 404
}
