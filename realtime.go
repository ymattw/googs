// Refs:
// https://forums.online-go.com/t/ogs-api-notes/17136
// https://ogs.readme.io/docs/real-time-api
package googs

import (
	"fmt"

	// NOTE: Verified not working client packages:
	// socketio "github.com/maldikhan/go.socket.io/engine.io/v4/client"
	// socketio "github.com/googollee/go-socket.io" // v1.8.0-rc.1
	socketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

const (
	realtimeURL = "wss://online-go.com/socket.io/?transport=websocket&EIO=3"
)

func (c *Client) connect() error {
	conn, err := socketio.Dial(realtimeURL, transport.GetDefaultWebsocketTransport())
	if err != nil {
		return err
	}
	c.socket = conn

	// Authenticate with the chat_auth
	if err := c.socket.Emit("authenticate", map[string]any{
		"auth":      c.ChatAuth,
		"player_id": c.UserID,
		"username":  c.Username,
	}); err != nil {
		return err
	}
	return err
}

func (c *Client) Disconnect() {
	if c.socket != nil {
		c.socket.Close()
	}
}

// XXX: This does not seem to have an effect
func (c *Client) NotificationConnect() error {
	return c.socket.Emit("notification/connect", map[string]any{
		"auth":      c.NotificationAuth,
		"player_id": c.UserID,
		"username":  c.Username,
	})
}

// NOTE: To debug server reponse, start with a `map[string]any` callback
// parameter to ensure that the response can always be decoded successfully.
func (c *Client) ConnectGame(gameID int64, fn func(*GameData)) error {
	if fn != nil {
		callback := func(_ *socketio.Channel, g *GameData) { fn(g) }
		err := c.socket.On(fmt.Sprintf("game/%d/gamedata", gameID), callback)
		if err != nil {
			return err
		}
	}
	return c.socket.Emit("game/connect", map[string]any{
		"game_id":   gameID,
		"player_id": c.UserID,
		"chat":      false,
	})
}

func (c *Client) DisconnectGame(gameID int64) error {
	return c.socket.Emit("game/disconnect", map[string]any{
		"game_id": gameID,
	})
}

func (c *Client) OnMove(gameID int64, fn func(*GameMove)) error {
	callback := func(_ *socketio.Channel, m *GameMove) { fn(m) }
	return c.socket.On(fmt.Sprintf("game/%d/move", gameID), callback)
}

// PlayMove submits a move. ConnectGame must be called already..
func (c *Client) PlayMove(gameID int64, x, y int) error {
	return c.socket.Emit("game/move", map[string]any{
		"game_id":   gameID,
		"player_id": c.UserID,
		"move":      fmt.Sprintf("%c%c", rune('a'+x), rune('a'+y)), // SGF
	})
}

func (c *Client) Resign(gameID int64) error {
	return c.socket.Emit("game/resign", map[string]any{
		"auth":    c.ChatAuth,
		"game_id": gameID,
	})
}

func (c *Client) OnClock(gameID int64, fn func(*Clock)) error {
	callback := func(_ *socketio.Channel, clock *Clock) { fn(clock) }
	return c.socket.On(fmt.Sprintf("game/%d/clock", gameID), callback)
}
