package googs

import (
	"encoding/json"
	"os"
)

func (c *Client) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadClient(path string) (*Client, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Client
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
