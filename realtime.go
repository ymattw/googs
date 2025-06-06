package googs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	socketio "github.com/maldikhan/go.socket.io/socket.io/v5/client"
)

const (
	realtimeURL = "wss://online-go.com/socket.io/?transport=websocket&EIO=3"
)

// This is automatically called when Client is authenticated.
func (c *Client) connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := socketio.NewClient(socketio.WithRawURL(realtimeURL))
	if err != nil {
		return err
	}

	if err := conn.Connect(ctx); err != nil {
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

func marshal(data map[string]any) []byte {
	res, _ := json.Marshal(data)
	return res
}

func (c *Client) Disconnect() {
	if c.socket != nil {
		c.socket.Close()
	}
}

// GameConnect starts watching gamedata event and emit the connect message.
func (c *Client) GameConnect(gameID int64, fn func(*Game)) error {
	if fn != nil {
		callback := func(data []byte) {
			var g Game
			json.Unmarshal(data, &g)
			fn(&g)
		}
		c.socket.On(fmt.Sprintf("game/%d/gamedata", gameID), callback)
		// if err != nil {
		// 	return err
		// }
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

// func (c *Client) OnClock(gameID int64, fn func(*Clock)) error {
// 	callback := func(_ *socketio.Channel, clock *Clock) { fn(clock) }
// 	return c.socket.On(fmt.Sprintf("game/%d/clock", gameID), callback)
// }

// OnMove starts watching game move event.
func (c *Client) OnMove(gameID int64, fn func(*GameMove)) error {
	callback := func(data []byte) {
		var m GameMove
		json.Unmarshal(data, &m)
		fn(&m)
	}
	c.socket.On(fmt.Sprintf("game/%d/move", gameID), callback)
	return nil
}
