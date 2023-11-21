package shrink

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HttpClient http.Client
	Token      string
	Host       string
}

func NewClient(token string) *Client {
	return &Client{
		HttpClient: http.Client{
			Timeout: 3 * time.Second,
		},
		Token: token,
		Host:  "https://shrink.ch",
	}
}

// DoRequest makes a request to the Shrinkster API, caller is responsible to close response body
func (c *Client) DoRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.Host+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "Shrinkster CLI")

	if c.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	return c.HttpClient.Do(req)
}
