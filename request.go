package googs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	baseURL = "https://online-go.com"
)

func (c *Client) Get(api string) ([]byte, error) {
	return ogsGet("/api/v1/"+api, c.AccessToken)
}

func (c *Client) GetUnmarshaled(api string, ref any) error {
	body, err := c.Get(api)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &ref); err != nil {
		return err
	}
	return nil
}

func ogsGet(uri string, accessToken string) ([]byte, error) {
	req, err := http.NewRequest("GET", baseURL+uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server responded %q for %q", resp.Status, uri)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response of %q: %v", uri, err)
	}
	return body, nil
}

func ogsPost(uri string, data url.Values) ([]byte, error) {
	resp, err := http.PostForm(baseURL+uri, data)
	if err != nil {
		return nil, fmt.Errorf("failed to post %q: %v", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server responded %q for %q", resp.Status, uri)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response of %q: %v", uri, err)
	}
	return body, nil
}
