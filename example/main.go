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
	"net/url"
	"os"
	"strconv"
	"strings"

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
	go run . me
	go run . overview
	go run . getraw /players/1
	go run . getraw /me/games                     # all my games
	go run . getraw /me/games ended__isnull=true  # my active games
	go run . realtime                             # try realtime APIs
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
	case "me":
		me()
	case "overview":
		overview()
	case "getraw":
		getraw(args...)
	case "realtime":
		realtime(args...)
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

func me() {
	client := loadClient()
	me, err := client.AboutMe()
	fmt.Printf("%#v %v\n", me, err)
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
	if len(args) < 1 {
		fmt.Printf("Syntax: getraw <api> [param=value ...]\n")
		os.Exit(1)
	}
	api := args[0]
	values, err := pairsToURLValues(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	client := loadClient()
	body, err := client.GetRaw(api, values)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	formatted, _ := formatJSON(body)
	fmt.Printf("%s\n", string(formatted))
}

// pairsToURLValues a list of "key=value" strings into url.Values.
func pairsToURLValues(pairs []string) (url.Values, error) {
	values := make(url.Values)
	for _, pair := range pairs {
		// Use SplitN to handle cases where value might contain '='
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid parameter format: %s. Expected 'key=value'", pair)
		}
		key, value := parts[0], parts[1]
		values.Add(key, value)
	}
	return values, nil
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

func realtime(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: realtime <gameID>\n")
		os.Exit(1)
	}
	gameID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Printf("Invalid gameID %s\n", args[0])
		os.Exit(1)
	}

	client := loadClient()

	err = client.NotificationConnect()
	fmt.Printf("NotificationConnect got err: %v\n", err)

	client.GameConnect(gameID, func(g *googs.Game) {
		fmt.Printf("GameConnect got response:\n%s\n", formatObject(g))
	})

	client.OnMove(gameID, func(m *googs.GameMove) {
		fmt.Printf("OnMove got response:\n%s\n", formatObject(m))
	})

	// Keep the main goroutine alive to process events
	select {}
}
