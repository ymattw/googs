// Ref: https://apidocs.online-go.com/
package googs

import (
	"net/url"
)

func (c *Client) AboutMe() (*Me, error) {
	res := Me{}
	if err := c.Get("/me", nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) Overview() (*Overview, error) {
	res := Overview{}
	if err := c.Get("/ui/overview", nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) MyActiveGames() (*MyGames, error) {
	values := url.Values{}
	// Ref: https://github.com/search?q=repo%3Aonline-go%2Fonline-go.com+ended__isnull&type=code
	values.Set("ended__isnull", "true")
	res := MyGames{}
	if err := c.Get("/megames", values, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
