// Package blob wraps the Vercel Blob REST API for storing WASM binaries.
// Docs: https://vercel.com/docs/storage/vercel-blob/using-blob-sdk#upload-a-blob
package blob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiBase = "https://blob.vercel-storage.com"

// Client uploads files to Vercel Blob storage.
type Client struct {
	token string
	http  *http.Client
}

// New returns a Blob Client using the given BLOB_READ_WRITE_TOKEN.
func New(token string) *Client {
	return &Client{
		token: token,
		http:  &http.Client{Timeout: 120 * time.Second},
	}
}

// UploadResult is the JSON response from a successful Blob PUT.
type UploadResult struct {
	URL         string `json:"url"`
	DownloadURL string `json:"downloadUrl"`
	Pathname    string `json:"pathname"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

// Upload stores data at the given path inside the Blob store and returns the
// result. The path becomes the public URL suffix, e.g. "skills/foo/1.0.0/skill.wasm".
func (c *Client) Upload(ctx context.Context, path string, data []byte) (*UploadResult, error) {
	endpoint := fmt.Sprintf("%s/%s?token=%s", apiBase, url.PathEscape(path), url.QueryEscape(c.token))

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/wasm")
	req.Header.Set("X-Api-Version", "7")
	req.ContentLength = int64(len(data))

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vercel blob upload: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vercel blob returned HTTP %d: %s", resp.StatusCode, body)
	}

	var result UploadResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding blob response: %w", err)
	}
	return &result, nil
}

// UploadStream is the streaming variant for large WASM binaries without
// buffering the entire body in memory first.
func (c *Client) UploadStream(ctx context.Context, path string, r io.Reader, size int64) (*UploadResult, error) {
	endpoint := fmt.Sprintf("%s/%s?token=%s", apiBase, url.PathEscape(path), url.QueryEscape(c.token))

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/wasm")
	req.Header.Set("X-Api-Version", "7")
	req.ContentLength = size

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vercel blob stream upload: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vercel blob returned HTTP %d: %s", resp.StatusCode, body)
	}

	var result UploadResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding blob response: %w", err)
	}
	return &result, nil
}
