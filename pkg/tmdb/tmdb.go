package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	IncludeAdult string
	Language     string
	BaseURL      string
	httpClient   *http.Client
}

func New(authKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if httpClient.Transport == nil {
		httpClient.Transport = http.DefaultTransport
	}
	httpClient.Transport = auth{
		authKey: authKey,
		next:    httpClient.Transport,
	}
	return &Client{
		IncludeAdult: "false",
		Language:     "en-US",
		BaseURL:      "https://api.themoviedb.org",
		httpClient:   httpClient,
	}
}

func (c Client) baseForm() url.Values {
	form := make(url.Values)
	form.Add("include_adult", c.IncludeAdult)
	form.Add("language", c.Language)
	return form
}

var _ http.RoundTripper = auth{}

type auth struct {
	authKey string
	next    http.RoundTripper
}

func (a auth) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Bearer "+a.authKey)
	return a.next.RoundTrip(r)
}

func call[T any](ctx context.Context, c Client, url string, values url.Values) (T, error) {
	form := c.baseForm()
	for key, v := range values {
		for _, value := range v {
			form.Add(key, value)
		}
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url+"?"+form.Encode(), nil)
	req.Header.Add("accept", "application/json")

	var result T
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return result, err
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return result, errors.New(resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("decode: %w", err)
	}

	return result, nil
}
