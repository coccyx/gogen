package outputter

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"

	config "github.com/coccyx/gogen/internal"
)

type httpout struct {
	buf            *bytes.Buffer
	client         *http.Client
	resp           *http.Response
	initialized    bool
	closed         bool
	endpoint       string
	headers        map[string]string
	lastSampleName string
}

func (h *httpout) Send(item *config.OutQueueItem) error {
	if h.initialized == false {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		h.client = &http.Client{Transport: tr, Timeout: item.S.Output.Timeout}
		h.buf = bytes.NewBuffer([]byte{})
		h.initialized = true
	}
	_, err := io.Copy(h.buf, item.IO.R)
	if err != nil {
		return err
	}

	if h.buf.Len() > item.S.Output.BufferBytes {
		h.endpoint = item.S.Output.Endpoints[rand.Intn(len(item.S.Output.Endpoints))]
		h.headers = item.S.Output.Headers
		h.lastSampleName = item.S.Name
		return h.flush()
	}
	return nil
}

func (h *httpout) flush() error {
	req, err := http.NewRequest("POST", h.endpoint, h.buf)
	for k, v := range h.headers {
		req.Header.Add(k, v)
	}
	h.resp, err = h.client.Do(req)
	if err != nil && h.resp == nil {
		return fmt.Errorf("Error making request from sample '%s' to endpoint '%s': %s", h.lastSampleName, h.endpoint, err)
	}
	body, err := ioutil.ReadAll(h.resp.Body)
	if err != nil {
		return fmt.Errorf("Error making request from sample '%s' to endpoint '%s': %s", h.lastSampleName, h.endpoint, err)
	} else if h.resp.StatusCode < 200 || h.resp.StatusCode > 299 {
		return fmt.Errorf("Error making request from sample '%s' to endpoint '%s', status '%d': %s", h.lastSampleName, h.endpoint, h.resp.StatusCode, body)
	}
	h.resp.Body.Close()
	h.buf.Reset()
	return nil
}

func (h *httpout) Close() error {
	if !h.closed {
		h.flush()
		h.closed = true
	}
	return nil
}
