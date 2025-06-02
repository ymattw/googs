// Ref: https://apidocs.online-go.com/
package googs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
)

const (
	restBaseURL = "https://online-go.com"
)

func (c *Client) AboutMe() (*User, error) {
	res := User{}
	if err := c.Get("/api/v1//me", nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Overview returns active games, open challenges etc.
// NOTE: /me/games?ended__isnull=true can also return my active games.
func (c *Client) Overview() (*Overview, error) {
	res := Overview{}
	if err := c.Get("/api/v1//ui/overview", nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) Game(gameID int64) (*Game, error) {
	res := Game{}
	if err := c.Get(fmt.Sprintf("/api/v1/games/%d", gameID), nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) GameState(gameID int64) (*GameState, error) {
	res := GameState{}
	if err := c.Get(fmt.Sprintf("/api/v1/games/%d/state", gameID), nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) Get(uri string, params url.Values, ptr any) error {
	if reflect.ValueOf(ptr).Kind() != reflect.Ptr {
		return fmt.Errorf("ptr argument must be a pointer, got %T", ptr)
	}

	body, err := ogsGet(uri, c.AccessToken, params)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, ptr); err != nil {
		return err
	}
	return nil
}

func ogsGet(uri string, accessToken string, params url.Values) ([]byte, error) {
	url := restBaseURL + uri
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = params.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s -> %s", url, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s -> %w", url, err)
	}
	return body, nil
}

func ogsPost(uri string, data url.Values) ([]byte, error) {
	resp, err := http.PostForm(restBaseURL+uri, data)
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
