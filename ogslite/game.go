package main

import (
	"fmt"
	"os"
	"strconv"
)

func game(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: game <gameID>\n")
		os.Exit(1)
	}
	gameID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Printf("Invalid gameID %s\n", args[0])
		os.Exit(1)
	}

	client := loadClient()
	g, err := client.Game(gameID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", formatObject(g))
}
