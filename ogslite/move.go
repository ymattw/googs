package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ymattw/googs"
)

func move(args ...string) {
	if len(args) != 2 {
		fmt.Printf("Syntax: move <gameID> <A1Coord | resign>\n")
		os.Exit(1)
	}
	gameID, err := parseGameID(args[0])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	coord := args[1]

	client := loadClient()
	connected := make(chan struct{})
	if err := client.GameConnect(gameID, func(g *googs.GameData) {
		fmt.Printf("Connected to game %s\n", g)
		close(connected)
	}); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	waitSignal(connected, 5)
	defer client.GameDisconnect(gameID)

	moved := make(chan struct{})
	client.OnMove(gameID, func(m *googs.GameMove) {
		fmt.Printf("move played: %v\n", m)
		close(moved)
	})

	if coord == "resign" {
		if confirm(fmt.Sprintf("Resign https://online-go.com/game/%d, are you sure?", gameID)) {
			client.GameResign(gameID)
			waitSignal(moved, 5)
		}
		return
	}

	x, y, err := a1ToOrigin(19, coord)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	if confirm(fmt.Sprintf("Play at %s ([%d, %d]) on https://online-go.com/game/%d, proceed?", coord, x, y, gameID)) {
		client.GameMove(gameID, x, y)
		waitSignal(moved, 5)
	}
}

func a1ToOrigin(size int, coord string) (int, int, error) {
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
