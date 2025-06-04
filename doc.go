// Package googs implements REST and Realtime APIs of OGS (online-go.com).
//
// The package allows users to authenticate, connect to games, receive game
// events, and submit moves as a player or watch as an observer.
//
// Example usage:
//
// 1. First time login:
//
//	client := googs.NewClient(clientID, clientSecret)
//	err := client.Login(username, password)
//	...
//	// Recommended: persistent the credentials
//	client.Save(secretFile)
//	client.Overview()
//
// 2. Later:
//
//	client, err := googs.LoadClient(secretFile)
//	...
//	client.Overview()
//	client.GameConnect(12345, func(g *googs.Game) {
//		log.Printf("Sending game data %s", g.Overview())
//	})
//
// See real exmples in ./googs/ which hosts a working minimal OGS client
// program that you can use to watch and play games on OGS.
package googs
