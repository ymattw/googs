package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func playMove(args ...string) {
	if len(args) != 2 {
		fmt.Printf("Syntax: move <gameID> <coord> (\"A1\" format)\n")
		os.Exit(1)
	}
	gameID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Printf("Invalid gameID %s\n", args[0])
		os.Exit(1)
	}
	coord := args[1]

	client := loadClient()
	client.ConnectGame(gameID, nil)

	x, y, err := A1ToOrigin(19, coord)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	if confirm(fmt.Sprintf("Play at %s ([%d, %d]) on https://online-go.com/game/%d, proceed?", coord, x, y, gameID)) {
		client.PlayMove(gameID, x, y)
		// simple workaround to make sure the move is played
		time.Sleep(time.Second * 2)
	}
}

func A1ToOrigin(size int, coord string) (int, int, error) {
	if len(coord) < 2 {
		return -1, -1, fmt.Errorf("invalid coordinate string %q", coord)
	}

	col := rune(strings.ToUpper(coord)[0])
	row := coord[1:]

	var x int
	if col >= 'A' && col <= 'H' {
		x = int(col - 'A')
	} else if col >= 'J' && col <= 'T' { // Account for skipped 'I'
		x = int(col - 'A' - 1)
	} else {
		return -1, -1, fmt.Errorf("invalid column letter '%c' in coordinate %q: must be A-H or J-T (or a-h or j-t)", col, coord)
	}

	rowNum, err := strconv.Atoi(row)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid row number format in coordinate %q: %w", coord, err)
	}
	y := size - rowNum

	if x < 0 || x >= size || y < 0 || y >= size {
		return 0, 0, fmt.Errorf("converted coordinates [%d, %d] from %q are out of board bounds (0-%d)", x, y, coord, size-1)
	}

	return x, y, nil

}

func confirm(prompt string) bool {
	var answer string
	for {
		fmt.Printf("%s (yes/no) ", prompt)
		fmt.Scanln(&answer)
		switch answer {
		case "yes":
			return true
		case "no":
			return false
		}
	}
}
