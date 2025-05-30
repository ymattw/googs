package googs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	socketio "github.com/maldikhan/go.socket.io/socket.io/v5/client"
)

type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"-"` // Ignore, always "Bearer"
	ExpiresIn    int64     `json:"expires_in,omitempty"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

type Auth struct {
	ChatAuth         string `json:"chat_auth"`
	NotificationAuth string `json:"notification_auth"`
}

type Client struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	Token               // Embedded
	Auth                // Embedded

	// Not to persist
	Username string `json:"-"`
	UserID   int64  `json:"-"`

	// Internal
	socket *socketio.Client
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
	return c.authenticate(data)
}

func (c *Client) Identify() error {
	me, err := c.AboutMe()
	if err != nil {
		return err
	}
	c.Username = me.Username
	c.UserID = me.ID
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
	if err := c.authenticate(data); err != nil {
		return err
	}
	if err := c.Identify(); err != nil {
		return err
	}
	return nil
}

func (c *Client) authenticate(data url.Values) error {
	// Request tokens
	body, err := ogsPost("/oauth2/token/", data)
	if err != nil {
		return fmt.Errorf("failed to request token: %w", err)
	}
	if err := json.Unmarshal(body, &c.Token); err != nil {
		return err
	}

	c.ExpiresAt = time.Now().Add(time.Duration(c.ExpiresIn) * time.Second)
	c.ExpiresIn = 0 // Unset to omit when persisting to file

	// Request auth config
	if err := c.Get("/ui/config/", nil, &c.Auth); err != nil {
		return fmt.Errorf("failed to request auth config: %w", err)
	}

	// Initialize socket.io client
	if err := c.initSocket(); err != nil {
		return err
	}
	return nil
}

func (c *Client) MaybeRefresh(deadline time.Duration) (bool, error) {
	expiring := time.Now().Add(deadline).After(c.ExpiresAt)
	if expiring || c.Identify() != nil {
		err := c.refreshToken()
		return err == nil, err
	}
	return false, nil
}
