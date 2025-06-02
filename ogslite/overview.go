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
		whosTurn := "opponent's turn"
		if a.GameData.Clock.CurrentPlayerID == client.UserID {
			whosTurn = "my turn"
		}
		fmt.Printf("%-10d %-10s %s (B) vs %s (W), %d moves, %s\n",
			a.ID,
			a.Name,
			a.GameData.Players.Black.Username,
			a.GameData.Players.White.Username,
			len(a.GameData.Moves),
			whosTurn)
	}
}
