// Package googs implements REST and Realtime APIs of OGS (online-go.com).
//
// The package allows users to authenticate, connect to games, receive game
// events, and submit moves as a player or watch as an observer.
//
// Example usage:
//
// 1. Login once and persist credentials
//
//	client := googs.NewClient(clientID, clientSecret)
//	err := client.Login(username, password)
//	// if err != nil { ... }
//
//	client.Save(secretFile)
//
//	// Use REST API
//	overview, err := client.Overview())
//	// if err != nil { ... }
//	fmt.Printf("Total %d active games\n", len(overview.ActiveGames))
//
//	// Use Realtime API
//	client.GameConnect(12345)

//	client.OnGameData(12345, func(g *googs.Game) {
//		fmt.Printf("Received game data %s\n", g)
//	})
//
// 2. Load a client from a credential file
//
//	client, err := googs.LoadClient(secretFile)
//	// if err != nil { ... }
//
//	// Websocket is connected, ready to use the APIs
//
// See real examples in demo/ which is a working minimal OGS client program
// that you can use to watch and play games on OGS.
package googs
