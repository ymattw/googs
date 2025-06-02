/*
Package main offers an example usage of the googs package, and it's actually
a debug tool.

Requires an OGS application (https://online-go.com/oauth2/applications/, choose
Client Type: "Public", grant type "Resource owner password-based"). Keep note
of the client id, client secret is not needed.
*/
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ymattw/googs"
)

var (
	secretFile = flag.String("f", "secret.json", "file to write client info to and load from")
)

func usage() {
	fmt.Printf(`Usage:

	read -s PASS  # Avoid password being logged in shell history
	go run . -c clientID -u username -p "$PASS" login
	cat secret.json
	go run . overview
	go run . get /players/1
	go run . get /me/games       # all my games
	go run . game gameID            # show game information
	go run . state gameID           # show game state
	go run . watch gameID           # watch a game
	go run . playmove gameID coord  # submit a move (coord is "A1" format)
	` + "\n")
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}

	cmd := flag.Args()[0]
	args := flag.Args()[1:]

	switch cmd {
	case "login":
		login()
	case "overview":
		overview()
	case "get":
		get(args...)
	case "game":
		game(args...)
	case "state":
		state(args...)
	case "watch":
		watch(args...)
	case "playmove":
		playMove(args...)
	default:
		usage()
	}
}

func loadClient() *googs.Client {
	client, err := googs.LoadClient(*secretFile)
	if err != nil {
		fmt.Printf("failed to load client from file: %v\n", err)
		os.Exit(1)
	}
	return client
}

func formatJSON(body []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(body), "", "  ")
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func formatObject(obj any) string {
	var out bytes.Buffer
	data, _ := json.Marshal(obj)
	if json.Indent(&out, []byte(data), "", "  ") != nil {
		return ""
	}
	return out.String()
}
