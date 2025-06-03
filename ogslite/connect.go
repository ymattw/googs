package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ymattw/googs"
)

func connect(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: play <gameID>\n")
		os.Exit(1)
	}
	gameID, err := parseGameID(args[0])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	client := loadClient()

	// Fetch current game information once
	game, err := client.Game(gameID)
	if err != nil {
		fmt.Printf("failed to get game information %v\n", err)
		os.Exit(1)
	}
	// TODO: research how is the Game struct different for Rengo games
	if game.Rengo {
		fmt.Printf("Rengo game is not supported yet\n")
		os.Exit(1)
	}

	chGame := make(chan *googs.Game)
	if err := client.GameConnect(gameID, func(g *googs.Game) {
		chGame <- g
	}); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer client.GameDisconnect(gameID)
	fmt.Printf("Connected to game %s\n%s\n", game.URL(), game)

	isMyGame := game.IsMyGame(client.UserID)
	if !isMyGame {
		fmt.Printf("Not your game, watching only\n")
	}

	chGameMove := make(chan *googs.GameMove)
	client.OnMove(gameID, func(m *googs.GameMove) {
		chGameMove <- m
	})

	// Dynamic updated information
	var gameMove *googs.GameMove
	var gameState *googs.GameState

	for {
		gameState, err = client.GameState(gameID)
		if err != nil {
			fmt.Printf("failed to get GameState: %v\n", err)
			time.Sleep(time.Second * 2)
			continue
		}
		drawBoard(gameState)

		if gameState.Phase == "finished" {
			fmt.Printf("game is finished: %s win by %s\n", game.PlayerByID(game.Winner).Username, gameState.Outcome)
			break
		}

		currentPlayer := game.PlayerByID(gameState.PlayerToMove)
		if currentPlayer.ID == client.UserID {
			fmt.Printf("It's your turn\n")
			// TODO: play a move, resign, pass
		} else {
			s := game.BlackPlayer()
			if currentPlayer.ID == game.WhitePlayerID {
				s = game.WhitePlayer()
			}
			fmt.Printf("Waiting for %s to move\n", s)
		}

		select {
		case gameMove = <-chGameMove:
			fmt.Printf("received game move %d: %v\n", gameMove.MoveNumber, gameMove.Move.OriginCoordinate)

		case game = <-chGame:
			fmt.Printf("received new game data: %s\n", game)
		}
	}
}

func playMove() {
	// if coord == "resign" {
	// 	if confirm(fmt.Sprintf("Resign https://online-go.com/game/%d, are you sure?", gameID)) {
	// 		client.GameResign(gameID)
	// 		waitSignal(moved, 5)
	// 	}
	// 	return
	// }
	//
	// x, y, err := a1ToOrigin(19, coord)
	// if err != nil {
	// 	fmt.Printf("%v\n", err)
	// 	os.Exit(1)
	// }
	// if confirm(fmt.Sprintf("Play at %s ([%d, %d]) on https://online-go.com/game/%d, proceed?", coord, x, y, gameID)) {
	// 	client.GameMove(gameID, x, y)
	// 	waitSignal(moved, 5)
	// }
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
