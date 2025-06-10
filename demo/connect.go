package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ymattw/googs"
)

func connect(args ...string) {
	if len(args) != 1 {
		log.Fatal("Syntax: connect <gameID>")
	}
	gameID, err := parseGameID(args[0])
	if err != nil {
		log.Fatal(err)
	}

	client := loadClient()

	// Fetch current game information once
	game, err := client.Game(gameID)
	if err != nil {
		log.Fatalf("Failed to get game information %v", err)
	}

	// Buffered channels for game events and game moves
	chGame := make(chan *googs.Game, 10)
	chGameMove := make(chan *googs.GameMove, 10)
	defer close(chGame)
	defer close(chGameMove)

	if err := client.GameConnect(gameID, func(g *googs.Game) {
		// log.Printf("Sending game data %s", g.Overview())
		chGame <- g
	}); err != nil {
		log.Fatal(err)
	}
	defer client.GameDisconnect(gameID)
	log.Printf("Connected to game %s", game)

	if !game.IsMyGame(client.UserID) {
		log.Printf("Not your game, watching only")
	}

	client.OnMove(gameID, func(m *googs.GameMove) {
		// log.Printf("Sending submitted move %v", m)
		chGameMove <- m
	})

	// NOTE: `gameState` is updated on every move, `game` is only updated
	// on game result change.
	var gameState *googs.GameState
	numMoves := -1

	for {
		gameState, err = client.GameState(gameID)
		if err != nil {
			log.Printf("failed to get GameState: %v", err)
			time.Sleep(time.Second * 2)
			continue
		}
		if numMoves != gameState.MoveNumber {
			numMoves = gameState.MoveNumber
			drawBoard(gameState)
			log.Printf("%s", game.Status(gameState, client.UserID))
		}
		if gameState.Phase == "finished" {
			log.Printf("%s", game.Result(gameState))
			break
		}

		if gameState.IsMyTurn(client.UserID) {
			for {
				if err := playMove(client, gameID, game.BoardSize()); err != nil {
					log.Printf("Failed to submit move: %v", err)
				}
				break
			}
			select {
			case <-chGameMove:
			case game = <-chGame:
			case <-time.After(500 * time.Millisecond):
				log.Printf("Last move wasn't submitted, illegal move? Try again")
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
	log.Printf(`Your turn. Enter a coordinate in "A1" format, "pass" or "resign"`)
	fmt.Print("> ")
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

// Can also take a URL like https://online-go.com/game/123
func parseGameID(s string) (int64, error) {
	parts := strings.Split("/"+s, "/")
	last := parts[len(parts)-1]
	gameID, err := strconv.ParseInt(last, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to extract gameID from %q: %w", s, err)
	}
	return gameID, nil
}
