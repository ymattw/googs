package main

import (
	"fmt"

	"github.com/ymattw/googs"
)

const (
	// Full-width characters for stones and grid
	GridChar   = "〸"
	BlackStone = "⚫"
	WhiteStone = "⚪"

	// 24-bit color codes: https://en.wikipedia.org/wiki/ANSI_escape_code#24-bit
	GridFG      = "\033[38;2;31;31;31m"    // Grid fg: #1f1f1f (grey)
	BoardBG     = "\033[48;2;124;76;56m"   // Board bg: #7c4c38 (reddish-brown)
	LastBlackBG = "\033[48;2;230;230;230m" // Last move bg: #66ccff (grey)
	LastWhiteBG = "\033[48;2;204;0;0m"     // Last move bg: #cc0000 (red)
	Reset       = "\033[0m"
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
}

func newCell(g *googs.GameState, row, col int) Cell {
	return Cell{
		Stone:      Stone(g.Board[row][col]),
		IsLastMove: g.LastMove.X == col && g.LastMove.Y == row,
	}
}

func (c Cell) content() string {
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
