package d1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const Endpoint = "api.cloudflare.com"

var accountIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)
var databaseIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}(-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12}$`)

type Client struct {
	baseURL      string
	http         *http.Client
	pollInterval time.Duration
}

func NewClient() *Client {
	return &Client{
		baseURL:      "https://" + Endpoint + "/client/v4",
		http:         &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
		pollInterval: time.Second,
	}
}

type Credentials struct {
	AccountID  string
	DatabaseID string
	APIToken   string
}

func (c *Client) databaseURL(creds Credentials) (string, error) {
	switch {
	case !accountIDPattern.MatchString(creds.AccountID):
		return "", errors.New("invalid Cloudflare account ID")
	case !databaseIDPattern.MatchString(creds.DatabaseID):
		return "", errors.New("invalid D1 database ID")
	case strings.TrimSpace(creds.APIToken) == "":
		return "", errors.New("Cloudflare API token is required")
	}
	return c.baseURL + "/accounts/" + creds.AccountID + "/d1/database/" + creds.DatabaseID, nil
}

func (c *Client) DatabaseSize(ctx context.Context, creds Credentials) (int64, error) {
	endpoint, err := c.databaseURL(creds)
	if err != nil {
		return 0, err
	}
	var database struct {
		FileSize int64 `json:"file_size"`
	}
	err = c.request(ctx, http.MethodGet, endpoint, creds.APIToken, nil, &database)
	return database.FileSize, err
}

func (c *Client) Export(ctx context.Context, creds Credentials, dst io.Writer) error {
	endpoint, err := c.databaseURL(creds)
	if err != nil {
		return err
	}
	request := struct {
		OutputFormat    string `json:"output_format"`
		CurrentBookmark string `json:"current_bookmark,omitempty"`
	}{OutputFormat: "polling"}
	for {
		var result struct {
			Bookmark string `json:"at_bookmark"`
			Status   string `json:"status"`
			Success  bool   `json:"success"`
			File     struct {
				SignedURL string `json:"signed_url"`
			} `json:"result"`
		}
		body, err := json.Marshal(request)
		if err != nil {
			return err
		}
		if err := c.request(ctx, http.MethodPost, endpoint+"/export", creds.APIToken, body, &result); err != nil {
			return err
		}
		if !result.Success || result.Status == "error" {
			return errors.New("Cloudflare D1 export failed")
		}
		if result.Status == "complete" {
			return c.download(ctx, result.File.SignedURL, dst)
		}
		if result.Bookmark == "" {
			return errors.New("Cloudflare D1 export returned no polling bookmark")
		}
		request.CurrentBookmark = result.Bookmark
		timer := time.NewTimer(c.pollInterval)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

func (c *Client) request(ctx context.Context, method, endpoint, token string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("Cloudflare D1 request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Cloudflare D1 API returned HTTP %d", resp.StatusCode)
	}
	var envelope struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&envelope); err != nil {
		return errors.New("invalid response from Cloudflare D1 API")
	}
	if !envelope.Success {
		return errors.New("Cloudflare D1 API reported a failure")
	}
	if len(envelope.Result) == 0 || bytes.Equal(envelope.Result, []byte("null")) || json.Unmarshal(envelope.Result, out) != nil {
		return errors.New("invalid result from Cloudflare D1 API")
	}
	return nil
}

func (c *Client) download(ctx context.Context, signedURL string, dst io.Writer) error {
	link, err := url.Parse(signedURL)
	if err != nil || link.Scheme != "https" || link.Hostname() == "" || link.User != nil {
		return errors.New("Cloudflare D1 export returned an invalid download URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link.String(), nil)
	if err != nil {
		return errors.New("could not create D1 export download request")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("could not download D1 export")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("D1 export download returned HTTP %d", resp.StatusCode)
	}
	n, err := io.Copy(dst, resp.Body)
	if err != nil {
		return fmt.Errorf("could not save D1 export: %w", err)
	}
	if n == 0 {
		return errors.New("D1 export download was empty")
	}
	return nil
}
