package googs

import (
	"encoding/json"
	"fmt"
	"time"

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

// GameConnect connects to a game, client should call On... functions to start
// watching events.
func (c *Client) GameConnect(gameID int64) error {
	return c.socket.Emit("game/connect", map[string]any{
		"game_id":   gameID,
		"player_id": c.UserID,
		"chat":      false,
	})
}

// GameDisconnect disconnects a game.
func (c *Client) GameDisconnect(gameID int64) error {
	return c.socket.Emit("game/disconnect", map[string]any{
		"game_id": gameID,
	})
}

// OnGameData starts watching gamedata events.
func (c *Client) OnGameData(gameID int64, fn func(*Game)) error {
	// The first paramter is actually of type `*socketio.Channel` (unused)
	callback := func(_ any, g *Game) { fn(g) }
	return c.socket.On(fmt.Sprintf("game/%d/gamedata", gameID), callback)
}

// OnGamePhase starts watching game phase changes.
func (c *Client) OnGamePhase(gameID int64, fn func(GamePhase)) error {
	callback := func(_ any, p GamePhase) { fn(p) }
	return c.socket.On(fmt.Sprintf("game/%d/phase", gameID), callback)
}

// OnGameRemovedStones starts watching game removed stones changes.
func (c *Client) OnGameRemovedStones(gameID int64, fn func(*RemovedStones)) error {
	callback := func(_ any, r *RemovedStones) { fn(r) }
	return c.socket.On(fmt.Sprintf("game/%d/removed_stones", gameID), callback)
}

// OnGameRemovedStones starts watching game removed stones acceptance.
func (c *Client) OnGameRemovedStonesAccepted(gameID int64, fn func(*RemovedStonesAccepted)) error {
	callback := func(_ any, r *RemovedStonesAccepted) { fn(r) }
	return c.socket.On(fmt.Sprintf("game/%d/removed_stones_accepted", gameID), callback)
}

// OnClock starts watching clock events.
func (c *Client) OnClock(gameID int64, fn func(*Clock)) error {
	callback := func(_ any, clock *Clock) { fn(clock) }
	return c.socket.On(fmt.Sprintf("game/%d/clock", gameID), callback)
}

// OnMove starts watching game move events.
func (c *Client) OnMove(gameID int64, fn func(*GameMove)) error {
	callback := func(_ any, m *GameMove) { fn(m) }
	return c.socket.On(fmt.Sprintf("game/%d/move", gameID), callback)
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

func (c *Client) GameRemovedStonesAccept(gameID int64, g *GameState) error {
	return c.socket.Emit("game/removed_stones/accept", map[string]any{
		"game_id": gameID,
		"stones":  g.RemovalString(),
	})
}

func (c *Client) GameListQuery(list GameListType, from, limit int, where *GameListWhere, timeout time.Duration) (*GameListResponse, error) {
	data := map[string]any{
		"list":    list,
		"sort_by": "rank",
		"from":    from,
		"limit":   limit,
		"where":   where,
	}
	res, err := c.socket.Ack("gamelist/query", data, timeout)
	if err != nil {
		return nil, err
	}

	resp := GameListResponse{}
	if err := json.Unmarshal([]byte(res), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
