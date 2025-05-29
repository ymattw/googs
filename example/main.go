package main

import (
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

func main() {
	flag.Parse()

	client, err := googs.LoadClient(*file)
	if err != nil {
		fmt.Printf("failed to load client from file: %v\n", err)
		os.Exit(1)
	}

	refreshed, err := client.MaybeRefresh()
	if err == nil {
		if refreshed {
			client.Save(*file)
		}
	} else {
		fmt.Printf("Need to relogin\n")
		if *clientID == "" || *username == "" || *password == "" {
			fmt.Printf("New login requires clientID (-c), username (-u) and password (-p)\n")
			os.Exit(1)
		}

		client = googs.NewClient(*clientID, "")

		err = client.Login(*username, *password)
		if err != nil {
			fmt.Printf("Failed to login: %v\n", err)
			os.Exit(1)
		}
		client.Save(*file)
	}

	me, err := client.Me()
	fmt.Printf("%v %v\n", me, err)
}
