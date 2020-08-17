package bugsnagapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"runtime"
	"time"
)

const (
	baseURL = "https://api.bugsnag.com"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	token   string
	baseURL string
	client  HTTPClient
}

func newDefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          10,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
}

func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: baseURL,
		client:  newDefaultHTTPClient()}
}

func (c *Client) patch(path string, payload interface{}, headers *map[string]string) (*http.Response, error) {
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		return c.do("PUT", path, bytes.NewBuffer(data), headers)
	}
	return c.do("PUT", path, nil, headers)
}

func (c *Client) post(path string, payload interface{}, headers *map[string]string) (*http.Response, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return c.do("POST", path, bytes.NewBuffer(data), headers)
}

func (c *Client) get(path string) (*http.Response, error) {
	return c.do("GET", path, nil, nil)
}

func (c *Client) do(method, path string, body io.Reader, headers *map[string]string) (*http.Response, error) {
	endpoint := c.baseURL + path
	req, _ := http.NewRequest(method, endpoint, body)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/json; version=2")

	if headers != nil {
		for k, v := range *headers {
			req.Header.Set(k, v)
		}
	}

	return c.client.Do(req)
}
