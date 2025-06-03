// Refs:
// - https://github.com/online-go/goban/blob/main/src/engine/protocol/ClientToServer.ts
// - https://github.com/online-go/goban/blob/main/src/engine/protocol/ServerToClient.ts
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

// GameConnect starts watching gamedata and emit the connect message.
// NOTE: To debug server reponse, start with a `map[string]any` callback
// parameter to ensure that the response can always be decoded successfully.
func (c *Client) GameConnect(gameID int64, fn func(*GameData)) error {
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

func (c *Client) GameDisconnect(gameID int64) error {
	return c.socket.Emit("game/disconnect", map[string]any{
		"game_id": gameID,
	})
}

// GameMove submits a move (GameConnect must be called already).
func (c *Client) GameMove(gameID int64, x, y int) error {
	return c.socket.Emit("game/move", map[string]any{
		"game_id":   gameID,
		"player_id": c.UserID,
		"move":      fmt.Sprintf("%c%c", rune('a'+x), rune('a'+y)), // SGF
	})
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

func (c *Client) OnMove(gameID int64, fn func(*GameMove)) error {
	callback := func(_ *socketio.Channel, m *GameMove) { fn(m) }
	return c.socket.On(fmt.Sprintf("game/%d/move", gameID), callback)
}
