package googs

import (
	"fmt"

	socketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

const (
	// NOTE: So far only found github.com/graarh/golang-socketio works with the
	// `EIO=3` version. Verified that below socket.io packages do NOT work:
	//
	// - "github.com/maldikhan/go.socket.io/engine.io/v4/client"
	// - "github.com/googollee/go-socket.io" // v1.8.0-rc.1
	realtimeURL = "wss://online-go.com/socket.io/?transport=websocket&EIO=3"
)

// This is automatically called when Client is authenticated.
func (c *Client) connect() error {
	conn, err := socketio.Dial(realtimeURL, transport.GetDefaultWebsocketTransport())
	if err != nil {
		return err
	}
	c.socket = conn

	// Authenticate with user_jwt. The `chat/connect`, `incident/connect`,
	// and `notification/connect` messages have been removed and are an
	// implicitly called by the `authenticate` message.
	if err := c.socket.Emit("authenticate", map[string]any{
		"jwt": c.UserJWT,
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

// GameConnect starts watching gamedata event and emit the connect message.
func (c *Client) GameConnect(gameID int64, fn func(*Game)) error {
	if fn != nil {
		callback := func(_ *socketio.Channel, g *Game) { fn(g) }
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

func (c *Client) GameDisconnect(gameID int64) error {
	return c.socket.Emit("game/disconnect", map[string]any{
		"game_id": gameID,
	})
}

// GameMove submits a move (GameConnect must be called first).
func (c *Client) GameMove(gameID int64, x, y int) error {
	return c.socket.Emit("game/move", map[string]any{
		"game_id":   gameID,
		"player_id": c.UserID,
		"move":      fmt.Sprintf("%c%c", rune('a'+x), rune('a'+y)), // SGF
	})
}

func (c *Client) PassTurn(gameID int64) error {
	return c.GameMove(gameID, -1, -1)
}

func (c *Client) GameResign(gameID int64) error {
	return c.socket.Emit("game/resign", map[string]any{
		"game_id": gameID,
	})
}

func (c *Client) OnClock(gameID int64, fn func(*Clock)) error {
	callback := func(_ *socketio.Channel, clock *Clock) { fn(clock) }
	return c.socket.On(fmt.Sprintf("game/%d/clock", gameID), callback)
}

// OnMove starts watching game move event.
func (c *Client) OnMove(gameID int64, fn func(*GameMove)) error {
	callback := func(_ *socketio.Channel, m *GameMove) { fn(m) }
	return c.socket.On(fmt.Sprintf("game/%d/move", gameID), callback)
}
