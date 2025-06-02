package main

import (
	"fmt"
	"os"

	"github.com/ymattw/googs"
)

func watch(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: watch <gameID>\n")
		os.Exit(1)
	}
	gameID, err := parseGameID(args[0])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	client := loadClient()
	client.NotificationConnect()

	if err := client.ConnectGame(gameID, func(g *googs.GameData) {
		// fmt.Printf("ConnectGame got response:\n%s\n", formatObject(g))
		fmt.Printf("Connected to game %s\n", g)
	}); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	client.OnMove(gameID, func(m *googs.GameMove) {
		// fmt.Printf("OnMove got response:\n%s\n", formatObject(m))
		g, err := client.GameState(gameID)
		if err != nil {
			fmt.Printf("failed to get GameState: %v\n", err)
			return
		}
		drawBoard(g)
	})
	client.OnClock(gameID, func(c *googs.Clock) {
		fmt.Printf("OnClock got response:\n%s\n", formatObject(c))
	})

	// Keep the main goroutine alive to process events
	select {}
}
