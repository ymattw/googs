package googs

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPlayer_Ranking(t *testing.T) {
	tests := []struct {
		name   string
		player Player
		want   string
	}{
		{
			name:   "Professional rank 39",
			player: Player{ID: 1086650, Rank: 39, Professional: true},
			want:   "3p",
		},
		{
			name:   "Professional rank 44",
			player: Player{ID: 59468, Rank: 44, Professional: true},
			want:   "8p",
		},
		{
			name:   "Rank above or equal to 1037",
			player: Player{Rank: 1037.1},
			want:   "1p",
		},
		{
			name:   "Rank between 30 and 1037",
			player: Player{Rank: 30.0001},
			want:   "1d",
		},
		{
			name:   "Rank between 1 and 30",
			player: Player{Rank: 29.9999},
			want:   "1k",
		},
		{
			name:   "Rank less than 1",
			player: Player{Rank: 0.9999},
			want:   "?",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.player.Ranking()
			if got != tc.want {
				t.Errorf("%#v.Ranking() want %q, got %q", tc.player, tc.want, got)
			}
		})
	}
}

func TestTimestamp_UnmarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "valid unix timestamp (seconds)",
			input:   "1672531200", // 2023-01-01 00:00:00 UTC
			want:    time.Unix(1672531200, 0),
			wantErr: false,
		},
		{
			name:    "valid unix timestamp (milliseconds)",
			input:   "1672531200000", // 2023-01-01 00:00:00 UTC in ms
			want:    time.UnixMilli(1672531200000),
			wantErr: false,
		},
		{
			name:    "invalid timestamp (not a number)",
			input:   `"not a number"`,
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "invalid timestamp (empty string)",
			input:   `""`,
			want:    time.Time{},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got Timestamp
			err := json.Unmarshal([]byte(tc.input), &got)
			if (err != nil) != tc.wantErr {
				t.Errorf("Unmarshal(%q) want error %v, got %v", tc.input, tc.wantErr, err)
				return
			}
			if !tc.wantErr && !got.Equal(tc.want) {
				t.Errorf("Unmarshal(%q) want %v, got %v, %v", tc.input, tc.want, got, err)
			}
		})
	}
}

func TestOriginCoordinate_ToA1Coordinate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		coord     OriginCoordinate
		boardSize int
		want      *A1Coordinate
		wantErr   bool
	}{
		{
			name:      "valid coordinate",
			coord:     OriginCoordinate{X: 1, Y: 0},
			boardSize: 9,
			want:      &A1Coordinate{Col: 'B', Row: 9},
		},
		{
			name:      "valid coordinate (X > 8, skip 'I')",
			coord:     OriginCoordinate{X: 8, Y: 0},
			boardSize: 9,
			want:      &A1Coordinate{Col: 'J', Row: 9},
		},
		{
			name:      "valid coordinate (Y = 8)",
			coord:     OriginCoordinate{X: 0, Y: 8},
			boardSize: 9,
			want:      &A1Coordinate{Col: 'A', Row: 1},
		},
		{
			name:      "invalid coordinate (X out of bounds)",
			coord:     OriginCoordinate{X: 9, Y: 0},
			boardSize: 9,
			wantErr:   true,
		},
		{
			name:      "invalid coordinate (Y out of bounds)",
			coord:     OriginCoordinate{X: 0, Y: 9},
			boardSize: 9,
			wantErr:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.coord.ToA1Coordinate(tc.boardSize)
			if (err != nil) != tc.wantErr {
				t.Errorf("%+v.ToA1Coordinate(%d) want error %v, got %#v, %v", tc.coord, tc.boardSize, tc.wantErr, got, err)
				return
			}
			if !tc.wantErr {
				if got == nil || *got != *tc.want {
					t.Errorf("%+v.ToA1Coordinate(%d) want %#v, got %#v, %v", tc.coord, tc.boardSize, tc.want, got, err)
				}
			}
		})
	}
}

func TestNewA1Coordinate(t *testing.T) {
	for _, tc := range []struct {
		name    string
		coord   string
		want    *A1Coordinate
		wantErr bool
	}{
		{
			name:    "valid coordinate",
			coord:   "A1",
			want:    &A1Coordinate{Col: 'A', Row: 1},
			wantErr: false,
		},
		{
			name:    "valid coordinate (lowercase)",
			coord:   "j10",
			want:    &A1Coordinate{Col: 'J', Row: 10},
			wantErr: false,
		},
		{
			name:    "invalid column (I)",
			coord:   "I1",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid column (too high)",
			coord:   "[1", // Next to 'Z'
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid row (zero)",
			coord:   "A0",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid row (negative)",
			coord:   "A-1",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid row (too large)",
			coord:   "A26",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid input (short)",
			coord:   "A",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid input (empty)",
			coord:   "",
			want:    nil,
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewA1Coordinate(tc.coord)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewA1Coordinate(%q) want error %v, got %#v, %v", tc.coord, tc.wantErr, got, err)
				return
			}
			if !tc.wantErr {
				if got == nil || *got != *tc.want {
					t.Errorf("NewA1Coordinate(%q) want %#v, got %#v, %v", tc.coord, tc.want, got, err)
				}
			}
		})
	}
}

func TestA1Coordinate_ToOriginCoordinate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		coord     A1Coordinate
		boardSize int
		want      *OriginCoordinate
		wantErr   bool
	}{
		{
			name:      "valid coordinate (A1 on 9x9)",
			coord:     A1Coordinate{Col: 'A', Row: 1},
			boardSize: 9,
			want:      &OriginCoordinate{X: 0, Y: 8},
			wantErr:   false,
		},
		{
			name:      "valid coordinate (J9 on 9x9)",
			coord:     A1Coordinate{Col: 'J', Row: 9},
			boardSize: 9,
			want:      &OriginCoordinate{X: 8, Y: 0},
			wantErr:   false,
		},
		{
			name:      "valid coordinate (lowercase, J9 on 9x9)",
			coord:     A1Coordinate{Col: 'j', Row: 9},
			boardSize: 9,
			want:      &OriginCoordinate{X: 8, Y: 0},
			wantErr:   false,
		},
		{
			name:      "invalid coordinate (col too high)",
			coord:     A1Coordinate{Col: 'U', Row: 1},
			boardSize: 19,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid coordinate (row out of bounds, too high)",
			coord:     A1Coordinate{Col: 'A', Row: 10},
			boardSize: 9,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid coordinate (row zero)",
			coord:     A1Coordinate{Col: 'A', Row: 0},
			boardSize: 9,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid coordinate (col I)",
			coord:     A1Coordinate{Col: 'I', Row: 1},
			boardSize: 9,
			want:      nil,
			wantErr:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.coord.ToOriginCoordinate(tc.boardSize)
			if (err != nil) != tc.wantErr {
				t.Errorf("%#v.ToOriginCoordinate(%d) want error %v, got %#v, %v", tc.coord, tc.boardSize, tc.wantErr, got, err)
				return
			}
			if !tc.wantErr {
				if got == nil || *got != *tc.want {
					t.Errorf("%#v.ToOriginCoordinate(%d) want %#v, got %#v, %v", tc.coord, tc.boardSize, tc.want, got, err)
				}
			}
		})
	}
}
