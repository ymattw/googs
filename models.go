package googs

import (
	"encoding/json"
	"fmt"
	"time"
)

type Me struct {
	ID       int64
	Username string
	About    string
	Ranking  float32
	Ratings  OGSRating
}

type Player struct {
	ID           int64
	Username     string
	Country      string
	Professional bool
	Ranking      float32
	Ratings      OGSRating
	UIClass      string `json:"ui_class"`
}

type Glicko2 struct {
	Deviation   float32
	GamesPlayed int64 `json:"games_played"`
	Rating      float32
	Volatility  float32
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

type MyGames struct {
	Count   int
	Results []Game
}

type Game struct {
	ID        int64
	Name      string
	Creator   int64
	Width     int
	Height    int
	Rules     string
	ranked    bool
	Handicap  int
	Komi      string
	Outcome   string
	Annulled  bool
	Started   time.Time
	BlackLost bool `json:"black_lost"`
	WhiteLost bool `json:"white_lost"`
	Players   map[string]Player
}

type Overview struct {
	ActiveGames []ActiveGame `json:"active_games"`
}

// Only most important fields for now
type Clock struct {
	CurrentPlayerID int64 `json:"current_player"`
}

// Move is an array of [x, y, TimeDelta] but in different types, so needs
// a customized decoder.
type Move struct {
	X         int
	Y         int
	TimeDelta float32
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

	var timeDelta float32
	if err := json.Unmarshal(raw[2], &timeDelta); err != nil {
		return fmt.Errorf("error unmarshaling move.TimeDelta: %w", err)
	}

	m.X = x
	m.Y = y
	m.TimeDelta = timeDelta
	return nil
}

type ActiveGameJSON struct {
	Players   map[string]Player // keys are "black", "white"
	Width     int
	Height    int
	Rules     string
	ranked    bool
	Handicap  int
	Komi      float32
	Phase     string
	StartTime int64 `json:"start_time"`
	Moves     []Move

	Clock Clock
}

type ActiveGame struct {
	ID             int64
	Name           string
	ActiveGameJSON `json:"json"` // Embedded
}

func (g *ActiveGame) BlacksTurn() bool {
	return g.Clock.CurrentPlayerID == g.Players["black"].ID
}
