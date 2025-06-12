# Go OGS

## Summary

`googs` is a Go package implements REST and Realtime APIs of [OGS
(online-go.com)](https://online-go.com).

## Status

The package allows users to authenticate, connect to games, receive game
 events, and submit moves as a player or watch as an observer.

## Usage

First [request an OGS application](https://online-go.com/oauth2/applications/),
with `Authorization grant type` set to *Resource owner password-based*, keep
note of the client ID and client secret. Note using empty client secret is fine
if the `Client type` is *Public*.

### Login once and persist credentials

```go
client := googs.NewClient(clientID, clientSecret)
err := client.Login(username, password)
// if err != nil { ... }

client.Save(secretFile)

// Use REST API
overview, err := client.Overview())
// if err != nil { ... }
fmt.Printf("Total %d active games\n", len(overview.ActiveGames))

// Use Realtime API
client.GameConnect(12345, func(g *googs.Game) {
	fmt.Printf("Received game data %s\n", g)
})
```

### Load a client from a credential file

```go
client, err := googs.LoadClient(secretFile)
// if err != nil { ... }

// Websocket is connected, ready to use the APIs
```

### Demo

See example usages in `demo/` which is a **working** minimal OGS client program
that you can use to watch and play games on OGS.

Seeing is believing!

![Alt text](https://github.com/ymattw/googs/blob/main/demo/demo.png?raw=true)
