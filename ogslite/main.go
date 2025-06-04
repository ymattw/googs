// Package main offers a minimal OGS client mainly to showcase usage of the
// googs package, as well as serving as a debug tool for developing googs.
//
// Requires an OGS application (https://online-go.com/oauth2/applications/,
// choose Client Type: "Public" so that we do not need a client sect, choose
// grant type "Resource owner password-based").
package main

import (
	"flag"
	"log"

	"github.com/ymattw/googs"
)

var (
	secretFile = flag.String("f", "secret.json", "file to write client info to and load from")
)

const usageText = `Usage:

	read -s PASS                    # avoid log password into shell history
	go run . -c clientID -u username -p "$PASS" login
	cat secret.json                 # secrets are stored after login once

	go run . overview               # show my active games
	go run . connect 123            # connect to a game to watch or play
	go run . rest /api/v1/players/1 # debug rest API (shows user profile)
`

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatal(usageText)
	}

	cmd := flag.Args()[0]
	args := flag.Args()[1:]

	switch cmd {
	case "login":
		login()
	case "overview":
		overview()
	case "connect":
		connect(args...)
	case "rest":
		rest(args...)
	default:
		log.Fatal(usageText)
	}
}

func loadClient() *googs.Client {
	client, err := googs.LoadClient(*secretFile)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
