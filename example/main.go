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
	"strconv"
	"strings"
	"time"

	"github.com/ymattw/googs"
)

var (
	clientID = flag.String("c", "", "client ID")
	username = flag.String("u", "", "username")
	password = flag.String("p", "", "password")
	file     = flag.String("f", "secret.json", "file to write client info to and load from")
)

func usage() {
	fmt.Printf(`Usage:

	read -s PASS  # Avoid password being logged in shell history
	go run . -c clientID -u username -p "$PASS" login
	cat secret.json
	go run . overview
	go run . getraw /players/1
	go run . getraw /me/games       # all my games
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
	case "getraw":
		getraw(args...)
	case "watch":
		watch(args...)
	case "playmove":
		playMove(args...)
	default:
		usage()
	}
}

func login() {
	if *clientID == "" || *username == "" || *password == "" {
		fmt.Printf("Syntax: -c clientID -u username -p password login\n")
		os.Exit(1)
	}

	client := googs.NewClient(*clientID, "")
	if err := client.Login(*username, *password); err != nil {
		fmt.Printf("Failed to login: %v\n", err)
		os.Exit(1)
	}
	client.Save(*file)
	fmt.Printf("Credentials wrote to %s\n", *file)
}

func overview() {
	client := loadClient()
	v, err := client.Overview()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total %d active games\n", len(v.ActiveGames))
	for i, a := range v.ActiveGames {
		whosTurn := "opponent's turn"
		if a.Game.Clock.CurrentPlayerID == client.UserID {
			whosTurn = "my turn"
		}
		fmt.Printf("%d %s %s (B) vs %s (W), %d moves, %s\n", i+1, a.Name, a.Game.Players.Black.Username, a.Game.Players.White.Username, len(a.Game.Moves), whosTurn)
	}
}

func getraw(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: getraw <api>\n")
		os.Exit(1)
	}
	api := args[0]

	client := loadClient()
	body, err := client.GetRaw(api, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	formatted, _ := formatJSON(body)
	fmt.Printf("%s\n", string(formatted))
}

func loadClient() *googs.Client {
	client, err := googs.LoadClient(*file)
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

func watch(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: watch <gameID>\n")
		os.Exit(1)
	}
	gameID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Printf("Invalid gameID %s\n", args[0])
		os.Exit(1)
	}

	client := loadClient()
	client.NotificationConnect()
	client.ConnectGame(gameID, func(g *googs.Game) {
		fmt.Printf("ConnectGame got response:\n%s\n", formatObject(g))
	})
	client.OnMove(gameID, func(m *googs.GameMove) {
		fmt.Printf("OnMove got response:\n%s\n", formatObject(m))
	})

	// Keep the main goroutine alive to process events
	select {}
}

func playMove(args ...string) {
	if len(args) != 2 {
		fmt.Printf("Syntax: move <gameID> <coord> (\"A1\" format)\n")
		os.Exit(1)
	}
	gameID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Printf("Invalid gameID %s\n", args[0])
		os.Exit(1)
	}
	coord := args[1]

	client := loadClient()
	client.ConnectGame(gameID, nil)

	x, y, err := A1ToOrigin(19, coord)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	if confirm(fmt.Sprintf("Play at %s ([%d, %d]) on https://online-go.com/game/%d, proceed?", coord, x, y, gameID)) {
		client.PlayMove(gameID, x, y)
		// simple workaround to make sure the move is played
		time.Sleep(time.Second * 2)
	}
}

func A1ToOrigin(size int, coord string) (int, int, error) {
	if len(coord) < 2 {
		return -1, -1, fmt.Errorf("invalid coordinate string %q", coord)
	}

	col := rune(strings.ToUpper(coord)[0])
	row := coord[1:]

	var x int
	if col >= 'A' && col <= 'H' {
		x = int(col - 'A')
	} else if col >= 'J' && col <= 'T' { // Account for skipped 'I'
		x = int(col - 'A' - 1)
	} else {
		return -1, -1, fmt.Errorf("invalid column letter '%c' in coordinate %q: must be A-H or J-T (or a-h or j-t)", col, coord)
	}

	rowNum, err := strconv.Atoi(row)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid row number format in coordinate %q: %w", coord, err)
	}
	y := size - rowNum

	if x < 0 || x >= size || y < 0 || y >= size {
		return 0, 0, fmt.Errorf("converted coordinates [%d, %d] from %q are out of board bounds (0-%d)", x, y, coord, size-1)
	}

	return x, y, nil

}

func confirm(prompt string) bool {
	var answer string
	for {
		fmt.Printf("%s (yes/no) ", prompt)
		fmt.Scanln(&answer)
		switch answer {
		case "yes":
			return true
		case "no":
			return false
		}
	}
}
