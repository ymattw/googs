package googs

type User struct {
	ID           int     `json:"id"`
	Username     string  `json:"username"`
	Country      string  `json:"country,omitempty"`
	Professional bool    `json:"professional,omitempty"`
	Ranking      float32 `json:"ranking,omitempty"`
	// Ratings      map[string]Glicko2 `json:"ratings,omitempty"`
	UIClass string `json:"ui_class,omitempty"`
}

type Glicko2 struct {
	Deviation   float32 `json:"deviation"`
	GamesPlayed int     `json:"games_played,omitempty"`
	Rating      float32 `json:"rating"`
	Volatility  float32 `json:"volatility"`
}

type Me struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	About    string  `json:"about"`
	Ranking  float32 `json:"ranking,omitempty"`
	// Ratings  map[string]Glicko2 `json:"ratings,omitempty"`
}
