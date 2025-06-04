package main

import (
	"bufio"
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

	chGame := make(chan *googs.Game, 10)
	if err := client.GameConnect(gameID, func(g *googs.Game) {
		chGame <- g
	}); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer client.GameDisconnect(gameID)
	fmt.Printf("Connected to game %s\n%s\n", game.URL(), game.Overview())

	if !game.IsMyGame(client.UserID) {
		fmt.Printf("Not your game, watching only\n")
	}

	chGameMove := make(chan *googs.GameMove, 10)
	client.OnMove(gameID, func(m *googs.GameMove) {
		chGameMove <- m
	})

	// NOTE: `gameState` is updated on every move, `game` is only updated
	// on game result change.
	var gameState *googs.GameState
	numMoves := -1

	for {
		gameState, err = client.GameState(gameID)
		if err != nil {
			fmt.Printf("failed to get GameState: %v\n", err)
			time.Sleep(time.Second * 2)
			continue
		}
		if numMoves != gameState.MoveNumber {
			numMoves = gameState.MoveNumber
			fmt.Printf("\n")
			drawBoard(gameState)
			fmt.Printf("\n%s\n", game.State(gameState))
		}
		if gameState.Phase == "finished" {
			fmt.Printf("\n%s\n", game.Result(gameState))
			break
		}

		currentPlayer := game.CurrentPlayer(gameState)
		if currentPlayer.ID == client.UserID {
			for {
				if err := playMove(client, gameID); err != nil {
					fmt.Printf("Failed to submit move: %v\n", err)
				}
				break
			}
			select {
			case <-chGameMove:
			case game = <-chGame:
			case <-time.After(500 * time.Millisecond):
				fmt.Printf("Looks like last move wasn't submitted\n")
			}
		} else { // blocking
			select {
			case <-chGameMove:
			case game = <-chGame:
			}
		}
	}
}

func playMove(client *googs.Client, gameID int64) error {
	fmt.Printf(`Your turn. Enter a coordinate in "A1" format, "pass" or "resign":` + "\n> ")
	reader := bufio.NewReader(os.Stdin)
	cmd, _ := reader.ReadString('\n')
	cmd = strings.TrimSpace(strings.ToUpper(cmd))

	switch cmd {
	case "PASS":
		return client.PassTurn(gameID)
	case "RESIGN":
		return client.GameResign(gameID)
	default:
		x, y, err := a1ToOrigin(19, cmd)
		if err != nil {
			return err
		}
		return client.GameMove(gameID, x, y)
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
