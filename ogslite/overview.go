package main

import (
	"fmt"
	"os"
)

func overview() {
	client := loadClient()
	v, err := client.Overview()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total %d active games\n", len(v.ActiveGames))
	for _, a := range v.ActiveGames {
		prefix := " "
		if a.Clock.CurrentPlayerID == client.UserID {
			prefix = "*" // my turn
		}
		fmt.Printf("%s %s\n", prefix, a.Overview())
	}
}
