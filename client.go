package googs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	socketio "github.com/graarh/golang-socketio"
)

// Token represents an OAuth-compatible token structure.
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"-"` // Ignore, always "Bearer"
	ExpiresIn    int64     `json:"expires_in,omitempty"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

// Auth holds authentication credentials for OGS Realtime APIs.
type Auth struct {
	ChatAuth         string `json:"chat_auth"`
	NotificationAuth string `json:"notification_auth"`
	UserJWT          string `json:"user_jwt"`
}

// Client represents an authenticated client with credentials and tokens.
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

// NewClient creates a Client instance with the given client ID and secret,
// Login() should be called for authentication.
func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// Login authenticates the Client with the given username and password, also
// establishes websocket connection to OGS. The Client instance is ready to use
// right after.
func (c *Client) Login(username, password string) error {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("username", username)
	data.Set("password", password)
	if err := c.authenticate(data); err != nil {
		return err
	}

	if err := c.Identify(); err != nil {
		return err
	}

	if err := c.connect(); err != nil {
		return err
	}
	return nil
}

// Save stores authenticated Client credentials into a file in JSON format.
// This is recommended practice right after logged in via Login() once.
func (c *Client) Save(secretFile string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(secretFile, data, 0600)
}

// Load stores Client credentials from a JSON file previously written via
// Save(),  also establishes websocket connection to OGS so the Client is ready
// to use right after.
func LoadClient(secretFile string) (*Client, error) {
	data, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, err
	}
	var c Client
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	// OGS access token is valid for 30 days, refresh if it's expiring in
	// 7 days.
	refreshed, err := c.MaybeRefresh(time.Hour * 24 * 7)
	if err != nil {
		return nil, err
	}
	if refreshed {
		if err := c.Save(secretFile); err != nil {
			return nil, err
		}
	}

	if err := c.Identify(); err != nil {
		return nil, err
	}

	if err := c.connect(); err != nil {
		return nil, err
	}
	return &c, nil
}

// Identify verifies Client access token and populate Username & UserID fields.
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
	if err := c.Get("/api/v1/ui/config/", nil, &c.Auth); err != nil {
		return fmt.Errorf("failed to request auth config: %w", err)
	}

	return nil
}

// MaybeRefresh validates the expiry of Client credentials and refresh
// credentials on demand, a true value is returned when refresh happened
// successfully. Save() is expected to persist the new credentials.
func (c *Client) MaybeRefresh(deadline time.Duration) (bool, error) {
	expiring := time.Now().Add(deadline).After(c.ExpiresAt)
	if expiring || c.Identify() != nil {
		// TODO: Only fresh on 401 error
		err := c.refreshToken()
		return err == nil, err
	}
	return false, nil
}
