// Package main offers a minimal OGS client mainly to showcase usage of the
// googs package, as well as serving as a debug tool for developing googs.
//
// Requires an OGS application (https://online-go.com/oauth2/applications/,
// choose Client Type: "Public" so that we do not need a client sect, choose
// grant type "Resource owner password-based").
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/ymattw/googs"
)

var (
	secretFile = flag.String("f", "secret.json", "file to write client info to and load from")
)

const usageText = `Usage:

	read -s PASS                    # avoid log password into shell history
	go run . -c clientID -u username -p "$PASS" login
	cat secret.json			# secrets are stored after login once

	go run . overview		# show my active games
	go run . board 123              # show game state and print board
	go run . move 123 Q16           # submit a move at coordinate in "A1" format
	go run . watch 123              # watch a game
	go run . play 123               # connect to a game and play in GNU Go style
	go run . rest /api/v1/players/1 # debug rest API (shows user profile)
	go run . rest /api/v1/games/123 # show game information
`

func usage() {
	fmt.Printf(usageText + "\n")
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
	case "board":
		board(args...)
	case "move":
		move(args...)
	case "watch":
		watch(args...)
	case "rest":
		rest(args...)
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

func formatObject(obj any) string {
	var out bytes.Buffer
	data, _ := json.Marshal(obj)
	if json.Indent(&out, []byte(data), "", "  ") != nil {
		return ""
	}
	return out.String()
}

// Can also take a URL like https://online-go.com/game/123
func parseGameID(s string) (int64, error) {
	u, err := url.Parse(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %q: %w", s, err)
	}

	part := path.Base(u.Path)
	gameID, err := strconv.ParseInt(part, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to extract gameID from %q: %w", s, err)
	}

	return gameID, nil
}
