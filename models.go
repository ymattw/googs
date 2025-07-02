package googs

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type PlayerColor int

const (
	PlayerUnknown PlayerColor = iota
	PlayerBlack
	PlayerWhite
)

func (p PlayerColor) String() string {
	return [...]string{"Unknown", "Black", "White"}[p]
}

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
// represented in both seconds or milliseconds.
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return fmt.Errorf("Timestamp.UnmarshalJSON: expected a numeric Unix timestamp, but got %q: %w", string(b), err)
	}
	if ts > 1_000_000_000_000 { //  Assume milliseconds
		t.Time = time.UnixMilli(ts)
	} else {
		t.Time = time.Unix(ts, 0)
	}
	return nil
}

type GamePhase string

const (
	PlayPhase         GamePhase = "play"
	StoneRemovalPhase GamePhase = "stone removal"
	FinishedPhase     GamePhase = "finished"
)

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
	OpponentPlaysFirstAfterResume bool   `json:"opponent_plays_first_after_resume"`
	Outcome                       string // Only when Phase is "finished"
	Phase                         GamePhase
	PlayerPool                    map[string]Player `json:"player_pool"` // Keys are player IDs (string)
	Players                       Players
	Private                       bool
	Ranked                        bool
	Removed                       string
	Rengo                         bool
	Rules                         string
	Score                         Score       // Only available when Phase is "finished"
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
	WinnerID                      int64 `json:"winner"` // Only when Phase is "finished"
}

type Score struct {
	Black PlayerScore
	White PlayerScore
}

type PlayerScore struct {
	Handicap         int
	Komi             float32
	Prisoners        int
	ScoringPositions string `json:"scoring_positions"`
	Stones           int
	Territory        float32
	Total            float32
}

// Equivalent to Python `return x if b else y`
func cond[T any](b bool, x, y T) T {
	if b {
		return x
	}
	return y
}

func (g *Game) String() string {
	whoseTurn := cond(g.Clock.CurrentPlayerID == g.Players.Black.ID, "Black", "White")
	return fmt.Sprintf("%d %q %s vs %s, %d moves, %s to play",
		g.GameID,
		g.GameName,
		g.BlackPlayerTitle(),
		g.WhitePlayerTitle(),
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
	return cond(g.Players.Black.ID == myUserID, g.Players.White, g.Players.Black)
}

func (g *Game) PlayerByID(userID int64) Player {
	return g.PlayerPool[fmt.Sprintf("%d", userID)]
}

func (g *Game) BlackPlayer() Player {
	return g.Players.Black
}

func (g *Game) WhitePlayer() Player {
	return g.Players.White
}

func (g *Game) BlackPlayerTitle() string {
	return "(B) " + g.Players.Black.String()
}

func (g *Game) WhitePlayerTitle() string {
	return "(W) " + g.Players.White.String()
}

func (g *Game) Result() string {
	if g.Phase != FinishedPhase {
		return ""
	}
	winner := cond(g.WinnerID == g.BlackPlayerID, g.BlackPlayerTitle(), g.WhitePlayerTitle())
	return fmt.Sprintf("%s won by %s", winner, g.Outcome)
}

func (g *Game) Status(state *GameState, myUserID int64) string {
	if state == nil {
		return g.String() + " (unknown board state)"
	}
	if state.MoveNumber == 0 {
		return fmt.Sprintf("Game ready, %s to start", g.BlackPlayerTitle())
	}
	if state.Phase == FinishedPhase {
		return "Game has finished, " + g.Result()
	}

	var whoPlayed, turn string
	if g.IsMyGame(myUserID) {
		turn = cond(state.PlayerToMove == myUserID, "your", "opponent's")
		whoPlayed = cond(state.PlayerToMove == myUserID, "Opponent", "You")
	} else {
		turn = cond(state.PlayerToMove == g.BlackPlayerID, "Black's", "White's")
		whoPlayed = cond(state.PlayerToMove == g.BlackPlayerID, "White", "Black")
	}

	if state.LastMove.IsPass() {
		return fmt.Sprintf("%d moves. %s passed, %s turn", state.MoveNumber, whoPlayed, turn)
	}

	a1, _ := state.LastMove.ToA1Coordinate(g.BoardSize())
	return fmt.Sprintf("%d moves. %s played %s, %s turn", state.MoveNumber, whoPlayed, a1, turn)
}

func (g *Game) WhoseTurn(state *GameState) PlayerColor {
	return cond(state.PlayerToMove == g.BlackPlayer().ID, PlayerBlack, PlayerWhite)
}

// Player contains basic user information as part of Game.
type Player struct {
	ID           int64
	Username     string
	Professional bool
	Rank         float32

	// Accepted removals, see RemovedStones for explanation. Make it
	// a pointer and nil means "not accepted yet".
	AcceptedStones *string `json:"accepted_stones"`
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
	PausedSince     Timestamp `json:"paused_since"`
	Title           string
	WhitePlayerID   int64      `json:"white_player_id"`
	WhiteTime       PlayerTime `json:"white_time"`
	StartMode       bool
	Now             Timestamp // Only for OnClock
}

type ComputedClock struct {
	System         ClockSystem
	MainTime       float64
	PeriodsLeft    int
	PeriodTimeLeft float64 // Byoyomi only
	MovesLeft      int     // Canadian only
	BlockTimeLeft  float64 // Canadian only
	SuddenDeath    bool
	TimedOut       bool
}

// ComputeClock returns a computed clock struct of the given players.
func (c *Clock) ComputeClock(tc *TimeControl, player PlayerColor) *ComputedClock {
	var t PlayerTime
	var isTurn bool

	switch player {
	case PlayerBlack:
		t = c.BlackTime
		isTurn = c.CurrentPlayerID == c.BlackPlayerID
	case PlayerWhite:
		t = c.WhiteTime
		isTurn = c.CurrentPlayerID == c.WhitePlayerID
	default:
		return nil
	}

	// Pause clock if not turn or game has not started yet
	elapsed := cond(isTurn && !c.StartMode, time.Since(c.LastMove.Time).Seconds(), 0)

	switch tc.System {

	case ClockAbsolute, ClockFischer:
		mainTime := cond(isTurn, math.Max(0, t.ThinkingTime-elapsed), t.ThinkingTime)
		return &ComputedClock{
			System:      tc.System,
			MainTime:    mainTime,
			SuddenDeath: mainTime < 10,
			TimedOut:    mainTime < 1e-7,
		}

	case ClockByoyomi:
		var periodsLeft int
		var mainTime, periodTimeLeft, overTime float64
		if isTurn {
			if t.ThinkingTime > 1e-7 {
				mainTime = t.ThinkingTime - elapsed
				if mainTime < 1e-7 {
					overTime = -mainTime
					mainTime = 0
				}
			} else {
				mainTime = 0
				overTime = elapsed
			}
			periodsLeft = t.Periods
			periodTimeLeft = t.PeriodTime
			if overTime > 1e-7 {
				periodsUsed := math.Floor(overTime / tc.PeriodTime)
				periodsLeft -= int(periodsUsed)
				periodsLeft = cond(periodsLeft > 0, periodsLeft, 0)
				periodTimeLeft = tc.PeriodTime - (overTime - periodsUsed*tc.PeriodTime)
				periodTimeLeft = cond(periodTimeLeft > 1e-7, periodTimeLeft, 0)
			}
		} else {
			periodsLeft = t.Periods
			periodTimeLeft = tc.PeriodTime
			mainTime = t.ThinkingTime
		}
		return &ComputedClock{
			System:         tc.System,
			MainTime:       mainTime,
			PeriodsLeft:    periodsLeft,
			PeriodTimeLeft: periodTimeLeft,
			SuddenDeath:    periodsLeft <= 1,
			TimedOut:       mainTime < 1e-7 && periodsLeft < 0,
		}

	case ClockCanadian:
		var movesLeft int
		var mainTime, blockTimeLeft, overTime float64
		if isTurn {
			if t.ThinkingTime > 1e-7 {
				mainTime = t.ThinkingTime - elapsed
				if mainTime < 1e-7 {
					overTime = -mainTime
					mainTime = 0
				}
			} else {
				mainTime = 0
				overTime = elapsed
			}
			movesLeft = t.MovesLeft
			blockTimeLeft = t.BlockTime
			if overTime > 1e-7 {
				blockTimeLeft -= overTime
				blockTimeLeft = cond(blockTimeLeft > 1e-7, blockTimeLeft, 0)
			}
		} else {
			mainTime = t.ThinkingTime
			movesLeft = t.MovesLeft
			blockTimeLeft = t.BlockTime
		}
		return &ComputedClock{
			System:        tc.System,
			MainTime:      mainTime,
			MovesLeft:     movesLeft,
			BlockTimeLeft: blockTimeLeft,
			SuddenDeath:   mainTime < 1e-7 && (blockTimeLeft < 10 || movesLeft < 2),
			TimedOut:      mainTime < 1e-7 && blockTimeLeft < 1e-7,
		}

	case ClockSimple:
		mainTime := cond(isTurn, math.Max(0, tc.PerMove-elapsed), tc.PerMove)
		return &ComputedClock{
			System:      tc.System,
			MainTime:    mainTime,
			SuddenDeath: mainTime < 10,
			TimedOut:    mainTime < 1e-7,
		}
	}
	return nil
}

func (c ComputedClock) String() string {
	if c.TimedOut {
		return "Timeout"
	}

	switch c.System {
	case ClockAbsolute, ClockFischer, ClockSimple:
		return fmt.Sprintf("%s%s", prettyTime(c.MainTime), cond(c.SuddenDeath, " (SD)", ""))
	case ClockByoyomi:
		if c.SuddenDeath {
			return fmt.Sprintf("%s (SD)", prettyTime(c.PeriodTimeLeft))
		}
		if c.MainTime > 0 {
			return fmt.Sprintf("%s +%s (%d)", prettyTime(c.MainTime), prettyTime(c.PeriodTimeLeft), c.PeriodsLeft)
		}
		return fmt.Sprintf("%s (%d)", prettyTime(c.PeriodTimeLeft), c.PeriodsLeft)
	case ClockCanadian:
		if c.SuddenDeath {
			return fmt.Sprintf("%s/%d (SD)", prettyTime(c.BlockTimeLeft), c.MovesLeft)
		}
		if c.MainTime > 0 {
			return fmt.Sprintf("%s +%s/%d", prettyTime(c.MainTime), prettyTime(c.BlockTimeLeft), c.MovesLeft)
		}
		return fmt.Sprintf("%s/%d", prettyTime(c.BlockTimeLeft), c.MovesLeft)
	}
	return "??:??"
}

func prettyTime(seconds float64) string {
	days := math.Floor(seconds / 86400)
	seconds -= days * 86400
	hours := math.Floor(seconds / 3600)
	seconds -= hours * 3600
	minutes := math.Floor(seconds / 60)
	seconds -= minutes * 60

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%.0fd%.0fh", days, hours)
		}
		// "1d" is confusing, use "24h" instead
		return fmt.Sprintf("%.0fh", days*24)
	}
	if hours > 0 {
		return fmt.Sprintf("%.0fh%s", hours, cond(minutes > 0, fmt.Sprintf("%.0fm", minutes), ""))
	}
	if minutes > 0 {
		return fmt.Sprintf("%.0f:%02.0f", minutes, seconds)
	}
	return fmt.Sprintf("%.0fs", seconds)
}

type PlayerTime struct {
	// Non Rengo games
	PeriodTime     float64 `json:"period_time"`
	PeriodTimeLeft float64 `json:"period_time_left"` // Byoyomi only
	Periods        int
	ThinkingTime   float64 `json:"thinking_time"`
	MovesLeft      int     `json:"moves_left"` // Canadian only
	BlockTime      float64 `json:"block_time"` // Canadian only

	// Only for Rengo games
	Value Timestamp
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

type ClockSystem string

const (
	ClockAbsolute ClockSystem = "absolute"
	ClockByoyomi  ClockSystem = "byoyomi"
	ClockCanadian ClockSystem = "canadian"
	ClockFischer  ClockSystem = "fischer"
	ClockSimple   ClockSystem = "simple"
)

type TimeControl struct {
	System          ClockSystem
	Speed           string
	PauseOnWeekends bool `json:"pause_on_weekends"`

	// Absolute
	TotalTime float64 `json:"total_time"`

	// Byoyomi
	MainTime   float64 `json:"main_time"`   // Also for Canadian
	PeriodTime float64 `json:"period_time"` // Also for Canadian
	Periods    int
	PeriodsMax int `json:"periods_max"`
	PeriodsMin int `json:"periods_min"`

	// Canadian
	StonesPerPeriod int `json:"stones_per_period"`

	// Fischer
	InitialTime   float64 `json:"initial_time"`
	TimeIncrement float64 `json:"time_increment"`
	MaxTime       float64 `json:"max_time"`

	// Simple
	PerMove float64 `json:"per_move"`
}

func (t TimeControl) String() string {
	switch t.System {
	case ClockAbsolute:
		return fmt.Sprintf("%s %s", t.System, prettyTime(t.TotalTime))
	case ClockByoyomi:
		return fmt.Sprintf("%s %s+%sx%d", t.System, prettyTime(t.MainTime), prettyTime(t.PeriodTime), t.Periods)
	case ClockCanadian:
		return fmt.Sprintf("%s %s+%s/%d moves", t.System, prettyTime(t.MainTime), prettyTime(t.PeriodTime), t.StonesPerPeriod)
	case ClockFischer:
		return fmt.Sprintf("%s %s+%s/ max %s", t.System, prettyTime(t.InitialTime), prettyTime(t.TimeIncrement), prettyTime(t.MaxTime))
	case ClockSimple:
		return fmt.Sprintf("%s %s/move", t.System, prettyTime(t.PerMove))
	}
	return string(t.System)
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
	// Phase has value "play", "stone removal", "finished" etc.
	Phase GamePhase

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

func (g *GameState) RemovalString() string {
	var pairs []string
	for y, row := range g.Removal {
		for x, val := range row {
			if val == 1 {
				move := fmt.Sprintf("%c%c", rune('a'+x), rune('a'+y)) // SGF
				pairs = append(pairs, move)
			}
		}
	}
	return strings.Join(pairs, "")
}

// RemovedStones is the response of Realtime API "game/:id/removed_stones".
type RemovedStones struct {
	// Result removal string is a sequence of SGF coordinates, e.g.
	// "edhdid" is equivalent to origin coordinates (3,4) (3,7) (3,8).
	AllRemoved string `json:"all_removed"`

	// Removal changes
	Removed bool
	Stones  string
}

// RemovedStonesAccepted is the response of Realtime API "game/:id/removed_stones_accepted".
type RemovedStonesAccepted struct {
	PlayerID int64 `json:"player_id"`

	// Result removal string is a sequence of SGF coordinates, e.g.
	// "edhdid" is equivalent to origin coordinates (3,4) (3,7) (3,8).
	Stones  string
	Players Players

	// This will change to "finished" when both sides accepted
	Phase GamePhase
	Score Score

	// Only available when Phase is "finished"
	EndTime  Timestamp `json:"end_time"`
	Outcome  string
	WinnerID int64 `json:"winner"`
}

func (r *RemovedStonesAccepted) Result() string {
	if r.Phase != FinishedPhase {
		return ""
	}
	winner := cond(r.WinnerID == r.Players.Black.ID, "(B) "+r.Players.Black.String(), "(W) "+r.Players.White.String())
	return fmt.Sprintf("%s won by %s", winner, r.Outcome)
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

type GameListWhere struct {
	HideRanked     bool    `json:"hide_ranked"`
	HideUnranked   bool    `json:"hide_unranked"`
	RengoOnly      bool    `json:"rengo_only"`
	Hide19x19      bool    `json:"hide_19x19"`
	Hide9x9        bool    `json:"hide_9x9"`
	Hide13x13      bool    `json:"hide_13x13"`
	HideOther      bool    `json:"hide_other"`
	HideTournament bool    `json:"hide_tournament"`
	HideLadder     bool    `json:"hide_ladder"`
	HideOpen       bool    `json:"hide_open"`
	HideHandicap   bool    `json:"hide_handicap"`
	HideEven       bool    `json:"hide_even"`
	HideBotGames   bool    `json:"hide_bot_games"`
	HideBeginning  bool    `json:"hide_beginning"`
	HideMiddle     bool    `json:"hide_middle"`
	HideEnd        bool    `json:"hide_end"`
	PlayerIDs      []int64 `json:"players"`
	TournamentID   int64   `json:"tournament_id"`
	LadderID       int64   `json:"ladder_id"`
	MalkOnly       bool    `json:"malk_only"`
}

type GameListEntry struct {
	ID               int64
	GroupIDs         []int64         `json:"group_ids"`
	GroupIDsMap      map[string]bool `json:"group_ids_map"`
	KidsGoGame       bool            `json:"kidsgo_game"`
	Phase            GamePhase
	Name             string
	PlayerToMove     int64 `json:"player_to_move"`
	Width            int
	Height           int
	MoveNumber       int `json:"move_number"`
	Paused           int // XXX: server response is a number 0/1
	Private          bool
	Black            Player
	White            Player
	Rengo            bool
	DroppedPlayerID  int64     `json:"dropped_player"`
	RengoCasualMode  bool      `json:"rengo_casual_mode"`
	SecondsPerMove   int64     `json:"time_per_move"`
	ClockExpiration  Timestamp `json:"clock_expiration"`
	BotGame          bool      `json:"bot_game"`
	Ranked           bool
	Handicap         int
	TournamentID     int64 `json:"tournament_id"`
	LadderID         int64 `json:"ladder_id"`
	Komi             float32
	InBeginning      bool `json:"in_beginning"`
	InMiddle         bool `json:"in_middle"`
	InEnd            bool `json:"in_end"`
	MalkovichPresent bool `json:"malkovich_present"`
}

type GameListType string

const (
	LiveGameList           GameListType = "live"
	CorrespondenceGameList GameListType = "corr"
	KidsGoGameList         GameListType = "kidsgo"
)

type GameListResponse struct {
	List    GameListType
	SortBy  string `json:"by"`
	Size    int
	Where   GameListWhere
	From    int
	Limit   int
	Results []GameListEntry
}

type GameChat struct {
	Channel string
	Line    GameChatLine
}

type GameChatLine struct {
	ChatID       string `json:"chat_id"`
	Body         string
	Date         Timestamp
	MoveNumber   int `json:"move_number"`
	Channel      string
	PlayerID     int64 `json:"player_id"`
	Username     string
	Professional int // XXX: server response is a number 0/1
	Ranking      float32
}
