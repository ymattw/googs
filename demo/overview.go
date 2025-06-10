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
	for _, g := range v.ActiveGames {
		prefix := " "
		whoseTurn := "Opponent's turn"
		if g.IsMyTurn(client.UserID) {
			prefix = "*" // my turn
			whoseTurn = "Your turn"
		}
		fmt.Printf("%s %d %-10q %s vs %s, %d moves, %s\n",
			prefix,
			g.GameID,
			g.GameName,
			g.BlackPlayerTitle(),
			g.WhitePlayerTitle(),
			len(g.Moves),
			whoseTurn)
	}
}
