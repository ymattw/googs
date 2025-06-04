package main

import (
	"fmt"
	"log"
)

func overview() {
	client := loadClient()
	v, err := client.Overview()
	if err != nil {
		log.Fatal(err)
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
