package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ymattw/googs"
)

const (
	// Full-width characters for stones and grid, best choices by far.
	GridChar   = "〸"
	HoshiChar  = "＊"
	BlackStone = "⚫"
	WhiteStone = "⚪"

	// 24-bit color codes: https://en.wikipedia.org/wiki/ANSI_escape_code#24-bit
	GridFG      = "\033[38;2;31;31;31m"    // Grid fg: #1f1f1f (grey)
	BoardBG     = "\033[48;2;124;76;56m"   // Board bg: #7c4c38 (reddish-brown)
	LastBlackBG = "\033[48;2;230;230;230m" // Last move bg: #66ccff (grey)
	LastWhiteBG = "\033[48;2;204;0;0m"     // Last move bg: #cc0000 (red)
	Reset       = "\033[0m"
)

var (
	hoshiPoints = map[int][]googs.OriginCoordinate{
		19: {
			{X: 3, Y: 3}, {X: 3, Y: 9}, {X: 3, Y: 15},
			{X: 9, Y: 3}, {X: 9, Y: 9}, {X: 9, Y: 15},
			{X: 15, Y: 3}, {X: 15, Y: 9}, {X: 15, Y: 15},
		},
		13: {
			{X: 3, Y: 3}, {X: 3, Y: 9},
			{X: 6, Y: 6},
			{X: 9, Y: 3}, {X: 9, Y: 9},
		},
		9: {
			{X: 2, Y: 2}, {X: 2, Y: 6},
			{X: 4, Y: 4},
			{X: 6, Y: 2}, {X: 6, Y: 6},
		},
	}
)

type Stone int

const (
	Empty Stone = iota
	Black
	White
)

type Cell struct {
	Stone      Stone
	IsLastMove bool
	IsHoshi    bool
}

func newCell(g *googs.GameState, row, col int) Cell {
	isHoshi := false
	hPoints := hoshiPoints[g.BoardSize()]
	for _, h := range hPoints {
		if h.X == col && h.Y == row {
			isHoshi = true
		}
	}
	return Cell{
		Stone:      Stone(g.Board[row][col]),
		IsLastMove: g.LastMove.X == col && g.LastMove.Y == row,
		IsHoshi:    isHoshi,
	}
}

func (c Cell) content() string {
	if c.Stone == Empty && c.IsHoshi {
		return HoshiChar
	}
	return map[Stone]string{
		Empty: GridChar,
		Black: BlackStone,
		White: WhiteStone,
	}[c.Stone]
}

func (c Cell) StyledContent() string {
	fg := GridFG
	bg := BoardBG
	if c.IsLastMove && c.Stone == Black {
		bg = LastBlackBG
	} else if c.IsLastMove && c.Stone == White {
		bg = LastWhiteBG
	}
	return fmt.Sprintf("%s%s%s%s", fg, bg, c.content(), Reset)
}

func colLabel(col int) rune {
	letter := 'Ａ' + rune(col) // Full-width Latin capital A
	if col >= 8 {
		letter += 1
	}
	return letter
}

// Board layout:
//
//	  ＡＢＣＤＥＦＧＨＪ
//	9 〸〸〸〸〸〸〸〸〸 9
//	8 〸〸〸〸〸〸〸〸〸 8
//	7 〸〸⚪〸〸⚫＊〸〸 7
//	6 〸〸〸〸〸〸〸〸〸 6
//	5 〸〸〸＊〸〸〸〸〸 5
//	4 〸〸〸〸〸〸⚪〸〸 4
//	3 〸〸⚫〸〸⚫⚪〸〸 3
//	2 〸〸〸〸〸〸〸〸〸 2
//	1 〸〸〸〸〸〸〸〸〸 1
//	  ＡＢＣＤＥＦＧＨＪ
func drawBoard(g *googs.GameState) {
	size := g.BoardSize()

	// Top coordinate labels (A, B, C, ... skipping I)
	fmt.Printf("%3s", " ") // 3-char offset for row numbers on the left
	for c := 0; c < size; c++ {
		fmt.Printf("%c", colLabel(c))
	}
	fmt.Println()

	for row := 0; row < size; row++ {
		// Left side coordinate label (19, 18, .., 1)
		fmt.Printf("%2d ", size-row)

		for col := 0; col < size; col++ {
			cell := newCell(g, row, col)
			fmt.Printf("%s", cell.StyledContent())
		}
		// Right side coordinate label (19, 18, .., 1)
		fmt.Printf(" %-2d\n", size-row)
	}

	// Bottom coordinate labels (A, B, C, ... skipping I)
	fmt.Printf("%3s", " ") // 3-char offset for row numbers on the left
	for c := 0; c < size; c++ {
		fmt.Printf("%c", colLabel(c))
	}
	fmt.Println()
}

// Private use for testing board drawing.
var board9 string = `
{
  "board": [
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 2, 0, 0, 1, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 2, 0, 0 ],
    [ 0, 0, 1, 0, 0, 1, 2, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0 ]
  ],
  "last_move": { "x": 6, "y": 6 }
}
`

var board13 string = `
{
  "board": [
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 2, 0, 0, 0, 0, 0, 2, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 1, 0, 0, 0, 0, 2, 0, 1, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 2, 1, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ]
  ],
  "last_move": { "x": 10, "y": 7 }
}
`
var board19 string = `
{
  "board": [
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 0, 0 ],
    [ 0, 0, 0, 1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 2, 1, 2, 1, 0, 0 ],
    [ 0, 0, 1, 0, 2, 0, 0, 0, 0, 2, 0, 0, 0, 0, 1, 1, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 1, 2, 1, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0 ],
    [ 0, 2, 1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 0, 0 ],
    [ 0, 2, 1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0 ],
    [ 0, 2, 1, 2, 0, 1, 0, 1, 0, 0, 0, 0, 0, 2, 0, 2, 0, 0, 0 ],
    [ 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ],
    [ 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ]
  ],
  "last_move": { "x": 14, "y": 1 }
}
`

func board() {
	var gameState googs.GameState

	for _, b := range []string{board9, board13, board19} {
		if err := json.Unmarshal([]byte(b), &gameState); err != nil {
			log.Fatal(err)
		}
		drawBoard(&gameState)
	}
}
