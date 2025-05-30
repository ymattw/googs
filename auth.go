package googs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"-"` // Ignore, always "Bearer"
	ExpiresIn    int64     `json:"expires_in,omitempty"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

type Client struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	Token               // Embedded
}

func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

func (c *Client) Login(username, password string) error {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("username", username)
	data.Set("password", password)

	c.requestToken(data)
	return nil
}

func (c *Client) refreshToken() error {
	if c.RefreshToken == "" {
		return fmt.Errorf("Client does not have a RefreshToken, login needed")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.RefreshToken)
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)

	c.requestToken(data)
	return nil
}

func (c *Client) requestToken(data url.Values) error {
	body, err := ogsPost("/oauth2/token/", data)
	if err != nil {
		return fmt.Errorf("failed to request token: %w", err)
	}
	if err := json.Unmarshal(body, &c.Token); err != nil {
		return err
	}

	c.ExpiresAt = time.Now().Add(time.Duration(c.ExpiresIn) * time.Second)
	c.ExpiresIn = 0 // Unset to omit when persisting to file
	return nil
}

func (c *Client) MaybeRefresh() (bool, error) {
	if c.AccessToken != "" && time.Now().Before(c.ExpiresAt) {
		return false, nil
	}
	if err := c.refreshToken(); err != nil {
		return false, err
	}
	return true, nil
}
