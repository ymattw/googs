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
type OGSRating struct {
	Version int

	// The `json:"-"` tag prevents the map itself from being marshaled back
	// into a "Ratings" field if you were to re-marshal this struct. We'll
	// populate this map manually during unmarshaling.
	Ratings map[string]Glicko2 `json:"-"`
}

func (r *OGSRating) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	r.Ratings = make(map[string]Glicko2)
	for key, value := range raw {
		if key == "version" {
			if err := json.Unmarshal(value, &r.Version); err != nil {
				return err
			}
		} else {
			var g Glicko2
			if err := json.Unmarshal(value, &g); err != nil {
				return err
			}
			r.Ratings[key] = g
		}
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
	AgaHandicapScoring            bool    `json:"aga_handicap_scoring"`
	AllowSelfCapture              bool    `json:"allow_self_capture"`
	AllowSuperko                  bool    `json:"allow_superko"`
	AutomaticStoneRemoval         bool    `json:"automatic_stone_removal"`
	BlackPlayerID                 int64   `json:"black_player_id"`
	Clock                         Clock   `json:"clock"`
	GameID                        int64   `json:"game_id"`
	GameName                      string  `json:"game_name"`
	GroupIDs                      []any   `json:"group_ids"` // Can be []int or []string, depending on content
	Handicap                      int     `json:"handicap"`
	HandicapRankDifference        float32 `json:"handicap_rank_difference"`
	Height                        int
	InitialPlayer                 string `json:"initial_player"`
	Komi                          float32
	Latencies                     map[string]int64 // playerID => latencies
	Moves                         []Move
	OpponentPlaysFirstAfterResume bool `json:"opponent_plays_first_after_resume"`
	Phase                         string
	PlayerPool                    map[string]Player `json:"player_pool"` // Keys are player IDs (string)
	Players                       Players           `json:"players"`
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
	ActiveGames []Game `json:"active_games"`
}

// Move is an array of [x, y, TimeDelta] but in different types, so needs
// a customized decoder.
type Move struct {
	X         int
	Y         int
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
	GameData GameData `json:"json"` // Embedded
}

func (a *Game) BlacksTurn() bool {
	return a.GameData.Clock.CurrentPlayerID == a.GameData.Players.Black.ID
}

func (a *Game) WhitesTurn() bool {
	return a.GameData.Clock.CurrentPlayerID == a.GameData.Players.White.ID
}

type GameMove struct {
	GameID     int64 `json:"game_id"`
	Move       Move
	MoveNumber int `json:"move_number"`
}

type GameState struct {
	Board      [][]int
	MoveNumber int `json:"move_number"`
	LastMove   struct {
		X int
		Y int
	} `json:"last_move"`
}
