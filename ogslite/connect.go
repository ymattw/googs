package main

import (
	"bufio"
	"fmt"
	"os"
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

	// Buffered channels for game events and game moves
	chGame := make(chan *googs.Game, 10)
	chGameMove := make(chan *googs.GameMove, 10)
	defer close(chGame)
	defer close(chGameMove)

	if err := client.GameConnect(gameID, func(g *googs.Game) {
		// fmt.Printf("Sending game data %s\n", g.Overview())
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

	client.OnMove(gameID, func(m *googs.GameMove) {
		// fmt.Printf("Sending submitted move %v\n", m)
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
				if err := playMove(client, gameID, game.BoardSize()); err != nil {
					fmt.Printf("Failed to submit move: %v\n", err)
				}
				break
			}
			select {
			case <-chGameMove:
			case game = <-chGame:
			case <-time.After(500 * time.Millisecond):
				fmt.Printf("Last move wasn't submitted, illegal move? Try again\n")
			}
		} else { // blocking
			select {
			case <-chGameMove:
			case game = <-chGame:
			}
		}
	}
}

func playMove(client *googs.Client, gameID int64, boardSize int) error {
	fmt.Printf(`Your turn. Enter a coordinate in "A1" format, "pass" or "resign":` + "\n> ")
	reader := bufio.NewReader(os.Stdin)
	op, _ := reader.ReadString('\n')
	op = strings.TrimSpace(strings.ToUpper(op))

	switch op {
	case "PASS":
		return client.PassTurn(gameID)
	case "RESIGN":
		return client.GameResign(gameID)
	default:
		a1, err := googs.NewA1Coordinate(op)
		if err != nil {
			return err
		}
		coord, err := a1.ToOriginCoordinate(boardSize)
		if err != nil {
			return err
		}
		return client.GameMove(gameID, coord.X, coord.Y)
	}
}
