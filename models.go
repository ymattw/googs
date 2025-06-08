package googs

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// User contains full profile of a user
type User struct {
	ID           int64
	Username     string
	Country      string
	Professional bool
	About        string
	Ranking      float32
	Ratings      OGSRating
	IsBot        bool   `json:"is_bot"`
	IsFriend     bool   `json:"is_friend"`
	UIClass      string `json:"ui_class"`
}

// Glicko2 contains Glicko2 ratings of a user.
type Glicko2 struct {
	Deviation   float32
	GamesPlayed int64 `json:"games_played"`
	Rating      float32
	Volatility  float32
}

// OGSRating is a map of Glicko2 ratings with keys like "overall", "19x19" etc.
type OGSRating map[string]Glicko2

// UnmarshalJSON is a customized JSON decoder for properly handling the
// `"version": 5` field in the JSON returned by OGS server.
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

// Timestamp is a customized Time struct.
type Timestamp struct {
	time.Time
}

// UnmarshalJSON is a customized JSON decoder for properly handling timestamps
// represented in both seconds or miliseconds.
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

type Game struct {
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
	WinnerID                      int64 `json:"winner"` // Only when Phase == "finished"
}

func (g *Game) Overview() string {
	whoseTurn := "Black"
	if g.Clock.CurrentPlayerID == g.Players.White.ID {
		whoseTurn = "White"
	}
	return fmt.Sprintf("%d %-10q %s vs %s, %d moves, %s to play",
		g.GameID,
		g.GameName,
		g.BlackPlayerSummary(),
		g.WhitePlayerSummary(),
		len(g.Moves),
		whoseTurn)
}

func (g *Game) URL() string {
	return fmt.Sprintf("%s/game/%d", ogsBaseURL, g.GameID)
}

func (g *Game) BoardSize() int {
	return g.Height // client.Game() validates
}

func (g *Game) IsMyGame(myUserID int64) bool {
	return g.PlayerPool[fmt.Sprintf("%d", myUserID)].ID == myUserID
}

func (g *Game) IsMyTurn(myUserID int64) bool {
	return g.Clock.CurrentPlayerID == myUserID
}

func (g *Game) Opponent(myUserID int64) Player {
	if g.Players.Black.ID == myUserID {
		return g.Players.White
	}
	return g.Players.Black
}

func (g *Game) PlayerByID(userID int64) Player {
	return g.PlayerPool[fmt.Sprintf("%d", userID)]
}

func (g *Game) BlackPlayerSummary() string {
	return "(B) " + g.Players.Black.String()
}

func (g *Game) WhitePlayerSummary() string {
	return "(W) " + g.Players.White.String()
}

func (g *Game) Result(state *GameState) string {
	if g.Phase != "finished" {
		return ""
	}
	winner := g.BlackPlayerSummary()
	if g.WinnerID == g.WhitePlayerID {
		winner = g.WhitePlayerSummary()
	}
	return fmt.Sprintf("%s won by %s", winner, state.Outcome)
}

func (g *Game) Status(state *GameState) string {
	if state.MoveNumber == 0 {
		return fmt.Sprintf("Game ready, %s to start", g.BlackPlayerSummary())
	}
	if state.Phase == "finished" {
		return "Game has finished, " + g.Result(state)
	}

	whoPlayed := "Black"
	turn := "White"
	if state.PlayerToMove == g.BlackPlayerID {
		whoPlayed = "White"
		turn = "Black"
	}
	if state.LastMove.IsPass() {
		return fmt.Sprintf("%d moves, %s passed, %s's turn", state.MoveNumber, whoPlayed, turn)
	}

	a1, _ := state.LastMove.ToA1Coordinate(g.BoardSize())
	return fmt.Sprintf("%d moves, %s played %s, %s's turn", state.MoveNumber, whoPlayed, a1, turn)
}

// Player ontains basic user information as part of Game.
type Player struct {
	ID           int64
	Username     string
	Professional bool
	Rank         float32
}

func (p Player) String() string {
	return p.Username + "[" + p.Ranking() + "]"
}

// Ranking returns the player's OGS ranking as a string in notation like "1p",
// "2d", "3k" etc.
func (p *Player) Ranking() string {
	if p.Professional {
		return fmt.Sprintf("%.fp", p.Rank-36)
	}
	if p.Rank >= 1037 {
		return fmt.Sprintf("%.fp", p.Rank-1036)
	} else if p.Rank >= 30 {
		return fmt.Sprintf("%.fd", p.Rank-29)
	} else if p.Rank >= 1 {
		return fmt.Sprintf("%.fk", 30-math.Floor(float64(p.Rank)))
	}
	return "?"
}

type Clock struct {
	BlackPlayerID   int64      `json:"black_player_id"`
	BlackTime       PlayerTime `json:"black_time"`
	CurrentPlayerID int64      `json:"current_player"`
	Expiration      Timestamp
	GameID          int64     `json:"game_id"`
	LastMove        Timestamp `json:"last_move"`
	PausedSince     int64     `json:"paused_since"`
	Title           string
	WhitePlayerID   int64      `json:"white_player_id"`
	WhiteTime       PlayerTime `json:"white_time"`
	Now             Timestamp  // Only for OnClock
}

type PlayerTime struct {
	// Only for Rengo games
	Value Timestamp

	// Only for non Rengo games
	PeriodTime     int64   `json:"period_time"`
	PeriodTimeLeft float64 `json:"period_time_left"`
	Periods        int
	ThinkingTime   float64 `json:"thinking_time"`
}

// UnmarshalJSON is a customized JSON decoder for properly handling the
// different type of clock details in the Clock struct.
func (t *PlayerTime) UnmarshalJSON(data []byte) error {
	if json.Unmarshal(data, &t.Value) == nil {
		return nil
	}

	type alias PlayerTime // Avoid recursive decoding
	var pt alias
	if err := json.Unmarshal(data, &pt); err != nil {
		return err
	}
	*t = PlayerTime(pt)
	return nil
}

type Players struct {
	Black Player
	White Player
}

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

// Overview contains the overview as what users see after logged into OGS.
type Overview struct {
	ActiveGames []GameOverview `json:"active_games"`
}

// Move is a list of [x, y, TimeDelta] values.
type Move struct {
	OriginCoordinate
	TimeDelta float64
}

// UnmarshalJSON is a customized JSON decoder for properly handling the
// different types in the Move struct.
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

// GameOverview is almost identical to Game but decoded using a different json
// tag.
type GameOverview struct {
	Game `json:"json"` // Embedded
}

type GameMove struct {
	GameID     int64 `json:"game_id"`
	Move       Move
	MoveNumber int `json:"move_number"`
}

type GameState struct {
	// Phase has value "play", "finished" etc.
	Phase string

	// Number of moves already played.
	MoveNumber int `json:"move_number"`

	// Last move, coordinate [-1, -1] indicates a pass
	LastMove OriginCoordinate `json:"last_move"`

	// User ID of the player in turn.
	PlayerToMove int64 `json:"player_to_move"`

	// Game result, "Resignation", "2.5 points" etc.
	Outcome string

	// The 2-D array with value 0=Empty, 1=Black, 2=White
	Board   [][]int
	Removal [][]int
}

func (g *GameState) BoardSize() int {
	return len(g.Board) // client.GameState() validates
}

func (g *GameState) IsMyTurn(myUserID int64) bool {
	return g.PlayerToMove == myUserID
}

// OriginCoordinate is zero base coordinate.
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

func (c OriginCoordinate) ToA1Coordinate(boardSize int) (*A1Coordinate, error) {
	if c.X < 0 || c.X >= boardSize || c.Y < 0 || c.Y >= boardSize {
		return nil, fmt.Errorf("OriginCoordinate %s is out of board bounds [0-%d]", c, boardSize-1)
	}

	col := 'A' + rune(c.X)
	if c.X >= 8 { // Skip 'I'
		col += 1
	}
	row := boardSize - c.Y // Reverse counting
	return &A1Coordinate{Col: col, Row: row}, nil
}

// A1Coordinate is coordinate represented in format "A1", note letter 'I' is
// skipped.
type A1Coordinate struct {
	Col rune // 'A', 'B', ... (skip 'I')
	Row int  // 1, 2, ...
}

// A1Coordinate creates an instance from a coordinate string in format "A1".
func NewA1Coordinate(coord string) (*A1Coordinate, error) {
	if len(coord) < 2 {
		return nil, fmt.Errorf("invalid coordinate string %q", coord)
	}

	col := rune(strings.ToUpper(coord)[0])
	row := coord[1:]

	if col < 'A' || col > 'Z' || col == 'I' {
		return nil, fmt.Errorf("invalid column letter '%c' in coordinate %q: must be A-H or J-Z (or a-h or j-z)", col, coord)
	}
	rowNum, err := strconv.Atoi(row)
	if err != nil || rowNum <= 0 || rowNum > 25 {
		return nil, fmt.Errorf("invalid row number format in coordinate %q: %w", coord, err)
	}
	return &A1Coordinate{Col: col, Row: rowNum}, nil
}

func (c A1Coordinate) String() string {
	return fmt.Sprintf("%c%d", c.Col, c.Row)
}

func (c A1Coordinate) ToOriginCoordinate(boardSize int) (*OriginCoordinate, error) {
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
		return nil, fmt.Errorf("invalid column letter '%c' in A1Coordinate %q: must be A-H or J-T (or a-h or j-t)", col, c)
	}

	y := boardSize - c.Row
	if x < 0 || x >= boardSize || y < 0 || y >= boardSize {
		return nil, fmt.Errorf("coordinate %q is out of board bounds [0-%d]", c, boardSize-1)
	}
	return &OriginCoordinate{X: x, Y: y}, nil
}
