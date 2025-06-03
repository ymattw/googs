package main

import (
	"fmt"
	"os"

	"github.com/ymattw/googs"
)

const (
	// Full-width characters for stones and grid
	GridChar   = "〸"
	BlackStone = "⚫"
	WhiteStone = "⚪"

	// 256-color codes https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
	GridColor               = 238 // Grid foreground: light grey
	BoardColor              = 243 // Board background: dark grey
	BlackLastMoveBackground = 1
	WhiteLastMoveBackground = 1
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
	fg := fgColor(GridColor)
	bg := bgColor(BoardColor)

	switch c.Stone {
	case Black:
		if c.IsLastMove {
			bg = bgColor(BlackLastMoveBackground)
		}
	case White:
		if c.IsLastMove {
			bg = bgColor(WhiteLastMoveBackground)
		}
	}
	return fmt.Sprintf("%s%s%s%s", fg, bg, c.content(), "\033[0m")
}

func fgColor(colorCode int) string {
	return fmt.Sprintf("\033[38;5;%dm", colorCode)
}

func bgColor(colorCode int) string {
	return fmt.Sprintf("\033[48;5;%dm", colorCode)
}

func board(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: board <gameID>\n")
		os.Exit(1)
	}
	gameID, err := parseGameID(args[0])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	client := loadClient()
	g, err := client.GameState(gameID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	drawBoard(g)
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
	fmt.Printf("\n\n%s\n\n", g)
}
