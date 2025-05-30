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
	clientID = flag.String("c", "", "client ID")
	username = flag.String("u", "", "username")
	password = flag.String("p", "", "password")
	file     = flag.String("f", "client.json", "file to write client info to and load from")
)

func usage() {
	fmt.Printf("Usage: %s <login|me> [args ...]\n", os.Args[0])
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
	case "get":
		get(args...)
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
	me, err := client.Me()
	fmt.Printf("%v %v\n", me, err)
}

func get(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: get <api>\n")
		os.Exit(1)
	}
	client := loadClient()
	body, err := client.Get(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
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

	refreshed, err := client.MaybeRefresh()
	if err != nil {
		fmt.Printf("Refresh failed: %v, need to relogin\n", err)
		os.Exit(1)
	}

	if refreshed {
		client.Save(*file)
		fmt.Printf("Credentials refreshed and wrote to %s\n", *file)
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
