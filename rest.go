// Ref: https://apidocs.online-go.com/
package googs

import (
	"encoding/json"
)

func (c *Client) Me() (*Me, error) {
	res := Me{}
	body, err := c.Get("/me", nil)
	if err != nil {
		return nil, err
	}

	// Remove malformed `"version": 5` field
	// TODO: report to OGS if this is a bug
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	if ratings, ok := data["ratings"].(map[string]any); ok {
		delete(ratings, "version")
	}
	cleanBody, _ := json.Marshal(data)

	if err := json.Unmarshal(cleanBody, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
