package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ymattw/googs"
)

func watch(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: watch <gameID>\n")
		os.Exit(1)
	}
	gameID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Printf("Invalid gameID %s\n", args[0])
		os.Exit(1)
	}

	client := loadClient()
	client.NotificationConnect()
	client.ConnectGame(gameID, func(g *googs.GameData) {
		fmt.Printf("ConnectGame got response:\n%s\n", formatObject(g))
	})
	client.OnMove(gameID, func(m *googs.GameMove) {
		fmt.Printf("OnMove got response:\n%s\n", formatObject(m))
	})
	client.OnClock(gameID, func(c *googs.Clock) {
		fmt.Printf("OnClock got response:\n%s\n", formatObject(c))
	})

	// Keep the main goroutine alive to process events
	select {}
}
