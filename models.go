package googs

import (
	"encoding/json"
	"time"
)

type Me struct {
	ID       int64
	Username string
	About    string
	Ranking  float32 `json:",omitempty"`
	Ratings  OGSRating
}

type Player struct {
	ID           int
	Username     string
	Country      string  `json:",omitempty"`
	Professional bool    `json:",omitempty"`
	Ranking      float32 `json:",omitempty"`
	Ratings      OGSRating
	UIClass      string `json:"ui_class,omitempty"`
}

type Glicko2 struct {
	Deviation   float32
	GamesPlayed int64 `json:"games_played,omitempty"`
	Rating      float32
	Volatility  float32
}

// Rating is customized struct to fit the mixed format from OGS APIs
type OGSRating struct {
	Version int

	// The `json:"-"` tag prevents the map itself from being marshaled back
	// into a "Ratings" field if you were to re-marshal this struct. We'll
	// populate this map manually during unmarshaling.
	Ratings map[string]Glicko2 `json:"-"`
}

// UnmarshalJSON customizes how Rating is unmarshaled from JSON.
func (r *OGSRating) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	r.Ratings = make(map[string]Glicko2)
	for key, value := range m {
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
	Outcome   string `json:",omitempty"`
	Annulled  bool
	Started   time.Time
	BlackLost bool `json:"black_lost"`
	WhiteLost bool `json:"white_lost"`
	Players   map[string]Player
}
