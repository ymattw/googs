package main

import (
	"fmt"
	"os"
	"strconv"
)

func board(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: board <gameID>\n")
		os.Exit(1)
	}
	gameID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Printf("Invalid gameID %s\n", args[0])
		os.Exit(1)
	}

	client := loadClient()
	b, err := client.GameState(gameID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", formatObject(b))
}
