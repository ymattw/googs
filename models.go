package googs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// User contains full profile of a user
type User struct {
	ID           int64
	Username     string
	Country      string
	Professional bool
	About        string
	Ranking      float64
	Ratings      OGSRating
	IsBot        bool   `json:"is_bot"`
	IsFriend     bool   `json:"is_friend"`
	UIClass      string `json:"ui_class"`
}

type Glicko2 struct {
	Deviation   float64
	GamesPlayed int64 `json:"games_played"`
	Rating      float64
	Volatility  float64
}

// OGSRating contains a `"version": 5` field besides the string keyed ratings,
// so needs a customized decoder.
type OGSRating map[string]Glicko2

func (r *OGSRating) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	delete(raw, "version")

	*r = make(map[string]Glicko2)
	for key, value := range raw {
		g := Glicko2{}
		if err := json.Unmarshal(value, &g); err != nil {
			return err
		}
		(*r)[key] = g
	}
	return nil
}

type Timestamp struct {
	time.Time
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return fmt.Errorf("Timestamp.UnmarshalJSON: expected a numeric Unix timestamp, but got %q: %w", string(b), err)
	}
	if ts > 1_000_000_000_000 { //  Assume miliseconds
		t.Time = time.UnixMilli(ts)
	} else {
		t.Time = time.Unix(ts, 0)
	}
	return nil
}

type GameData struct {
	AgaHandicapScoring            bool  `json:"aga_handicap_scoring"`
	AllowSelfCapture              bool  `json:"allow_self_capture"`
	AllowSuperko                  bool  `json:"allow_superko"`
	AutomaticStoneRemoval         bool  `json:"automatic_stone_removal"`
	BlackPlayerID                 int64 `json:"black_player_id"`
	Clock                         Clock
	GameID                        int64  `json:"game_id"`
	GameName                      string `json:"game_name"`
	GroupIDs                      []any  `json:"group_ids"` // Can be []int or []string, depending on content
	Handicap                      int
	HandicapRankDifference        float32 `json:"handicap_rank_difference"`
	Height                        int
	InitialPlayer                 string `json:"initial_player"`
	Komi                          float32
	Latencies                     map[string]int64 // playerID => latencies
	Moves                         []Move
	OpponentPlaysFirstAfterResume bool `json:"opponent_plays_first_after_resume"`
	Phase                         string
	PlayerPool                    map[string]Player `json:"player_pool"` // Keys are player IDs (string)
	Players                       Players
	Private                       bool
	Ranked                        bool
	Rengo                         bool
	Rules                         string
	ScoreHandicap                 bool        `json:"score_handicap"`
	ScorePasses                   bool        `json:"score_passes"`
	ScorePrisoners                bool        `json:"score_prisoners"`
	ScoreStones                   bool        `json:"score_stones"`
	ScoreTerritory                bool        `json:"score_territory"`
	ScoreTerritoryInSeki          bool        `json:"score_territory_in_seki"`
	StartTime                     Timestamp   `json:"start_time"`
	StateVersion                  int         `json:"state_version"`
	StrictSekiMode                bool        `json:"strict_seki_mode"`
	SuperkoAlgorithm              string      `json:"superko_algorithm"`
	TimeControl                   TimeControl `json:"time_control"`
	WhiteMustPassLast             bool        `json:"white_must_pass_last"`
	WhitePlayerID                 int64       `json:"white_player_id"`
	Width                         int
}

func (g *GameData) String() string {
	whosTurn := "black"
	if g.WhitesTurn() {
		whosTurn = "white"
	}
	return fmt.Sprintf("%-10d %-10q %s (B) vs %s (W), %d moves, %s to play",
		g.GameID,
		g.GameName,
		g.Players.Black.Username,
		g.Players.White.Username,
		len(g.Moves),
		whosTurn)
}

func (g *GameData) BlacksTurn() bool {
	return g.Clock.CurrentPlayerID == g.Players.Black.ID
}

func (g *GameData) WhitesTurn() bool {
	return g.Clock.CurrentPlayerID == g.Players.White.ID
}

// Player ontains basic user information as part of GameData
type Player struct {
	ID           int64
	Username     string
	Professional bool
	Rank         float64
}

// Clock struct
type Clock struct {
	BlackPlayerID   int64      `json:"black_player_id"`
	BlackTime       PlayerTime `json:"black_time"`
	CurrentPlayerID int64      `json:"current_player"`
	Expiration      Timestamp
	GameID          int64     `json:"game_id"`
	LastMove        Timestamp `json:"last_move"`
	Title           string
	WhitePlayerID   int64      `json:"white_player_id"`
	WhiteTime       PlayerTime `json:"white_time"`
	Now             Timestamp  // Only for OnClock
}

// PlayerTime struct
type PlayerTime struct {
	PeriodTime     int64   `json:"period_time"`
	PeriodTimeLeft float64 `json:"period_time_left"`
	Periods        int
	ThinkingTime   int `json:"thinking_time"`
}

// Players struct
type Players struct {
	Black Player
	White Player
}

// TimeControl struct
type TimeControl struct {
	MainTime        int   `json:"main_time"`
	PauseOnWeekends bool  `json:"pause_on_weekends"`
	PeriodTime      int64 `json:"period_time"`
	Periods         int
	PeriodsMax      int `json:"periods_max"`
	PeriodsMin      int `json:"periods_min"`
	Speed           string
	System          string
	TimeControl     string `json:"time_control"`
}

type Overview struct {
	ActiveGames []GameOverview `json:"active_games"`
}

// Move is a list of [x, y, TimeDelta] but in different types, so needs
// a customized decoder.
type Move struct {
	OriginCoordinate
	TimeDelta float64
}

func (m *Move) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) < 3 {
		return fmt.Errorf("expected at least 3 elements in move array, got %d", len(raw))
	}

	var x int
	if err := json.Unmarshal(raw[0], &x); err != nil {
		return fmt.Errorf("error unmarshaling move.X: %w", err)
	}
	var y int
	if err := json.Unmarshal(raw[1], &y); err != nil {
		return fmt.Errorf("error unmarshaling move.Y: %w", err)
	}

	var timeDelta float64
	if err := json.Unmarshal(raw[2], &timeDelta); err != nil {
		return fmt.Errorf("error unmarshaling move.TimeDelta: %w", err)
	}

	m.X = x
	m.Y = y
	m.TimeDelta = timeDelta
	return nil
}

type Game struct {
	ID       int64
	Name     string
	GameData GameData // Embedded
}

// GameOverview is almost identical to Game but with a different json tag,
// which is for decoding the /api/v1/overview reponse.
type GameOverview struct {
	ID       int64
	Name     string
	GameData GameData `json:"json"` // Embedded
}

func (a *Game) BlacksTurn() bool {
	return a.GameData.BlacksTurn()
}

func (a *Game) WhitesTurn() bool {
	return a.GameData.WhitesTurn()
}

type GameMove struct {
	GameID     int64 `json:"game_id"`
	Move       Move
	MoveNumber int `json:"move_number"`
}

type GameState struct {
	Board      [][]int
	MoveNumber int              `json:"move_number"`
	LastMove   OriginCoordinate `json:"last_move"`
}

func (g *GameState) BoardSize() int {
	return len(g.Board) // GameState() already validated
}

func (g *GameState) String() string {
	if g.LastMove.IsPass() {
		return fmt.Sprintf("%d moves, last move was a pass", g.MoveNumber)
	}

	whoPlayed := ""
	switch g.Board[g.LastMove.Y][g.LastMove.X] {
	case 1:
		whoPlayed = "black"
	case 2:
		whoPlayed = "white"
	}
	a1, _ := g.LastMove.ToA1Coordinate(g.BoardSize())
	return fmt.Sprintf("%d moves, last move: %s %s", g.MoveNumber, whoPlayed, a1)
}

type OriginCoordinate struct {
	X int
	Y int
}

func (c OriginCoordinate) String() string {
	return fmt.Sprintf("[%d,%d]", c.X, c.Y)
}

func (c OriginCoordinate) IsPass() bool {
	return c.X == -1 || c.Y == -1
}

func (c OriginCoordinate) ToA1Coordinate(boardSize int) (A1Coordinate, error) {
	res := A1Coordinate{}
	if c.X < 0 || c.X >= boardSize || c.Y < 0 || c.Y >= boardSize {
		return res, fmt.Errorf("OriginCoordinate %s is out of board bounds [0-%d]", c, boardSize-1)
	}

	res.Col = 'A' + rune(c.X)
	if c.X >= 8 { // Skip 'I'
		res.Col += 1
	}
	res.Row = boardSize - c.Y // Reverse counting
	return res, nil
}

type A1Coordinate struct {
	Col rune // 'A', 'B', ... (skip 'I')
	Row int  // 1, 2, ...
}

func (c A1Coordinate) String() string {
	return fmt.Sprintf("%c%d", c.Col, c.Row)
}

func (c A1Coordinate) ToOriginCoordinate(boardSize int) (OriginCoordinate, error) {
	res := OriginCoordinate{}
	col := c.Col
	if col >= 'a' && col <= 'z' {
		col -= 'a' - 'A' // to upper case
	}

	var x int
	if col >= 'A' && col <= 'H' {
		x = int(col - 'A')
	} else if col >= 'J' && col <= 'T' { // Account for skipped 'I'
		x = int(col - 'A' - 1)
	} else {
		return res, fmt.Errorf("invalid column letter '%c' in A1Coordinate %q: must be A-H or J-T (or a-h or j-t)", col, c)
	}

	y := boardSize - c.Row
	if x < 0 || x >= boardSize || y < 0 || y >= boardSize {
		return res, fmt.Errorf("converted OriginCoordinate %s from %q are out of board bounds [0-%d]", res, c, boardSize-1)
	}
	return res, nil
}
