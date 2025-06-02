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
	ColStart   = 'Ａ' // Full-width Latin capital A

	// 256-color codes https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
	GridColor               = 238 // Grid foreground: light grey
	BoardColor              = 243 // Board background: dark grey
	BlackLastMoveBackground = 1
	WhiteLastMoveBackground = 1

	// ANSI color escape codes
	Reset   = "\033[0m"
	Reverse = "\033[7m"
)

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

func drawBoard(g *googs.GameState) {
	size := g.BoardSize()

	// Helper function to get the column letter, skipping 'I'
	getColumnLetter := func(col int) rune {
		letter := ColStart + rune(col)
		if col >= 8 {
			letter += 1
		}
		return letter
	}

	// Top coordinate labels (A, B, C, ... skipping I)
	fmt.Printf("%3s", " ") // 3-char offset for row numbers on the left
	for c := 0; c < size; c++ {
		fmt.Printf("%c", getColumnLetter(c))
	}
	fmt.Println()

	for row := 0; row < size; row++ {
		// Left side coordinate label (19, 18, .., 1)
		fmt.Printf("%2d ", size-row)

		for col := 0; col < size; col++ {
			var cellContent string
			isLastMove := g.LastMove.X == col && g.LastMove.Y == row

			switch g.Board[row][col] {
			case 0:
				cellContent = GridChar
			case 1: // Black stone
				cellContent = BlackStone
				if isLastMove {
					cellContent = bgColor(BlackLastMoveBackground) + BlackStone
				}
			case 2: // White stone
				cellContent = WhiteStone
				if isLastMove {
					cellContent = bgColor(WhiteLastMoveBackground) + WhiteStone
				}
			}
			fmt.Printf("%s%s%s%s", fgColor(GridColor), bgColor(BoardColor), cellContent, Reset)
		}
		// Right side coordinate label (19, 18, .., 1)
		fmt.Printf(" %-2d\n", size-row)
	}

	// Bottom coordinate labels (A, B, C, ... skipping I)
	fmt.Printf("%3s", " ") // 3-char offset for row numbers on the left
	for c := 0; c < size; c++ {
		fmt.Printf("%c", getColumnLetter(c))
	}
	fmt.Printf("\n\n%s\n\n", g)
}
