package main

import (
	"flag"
	"log"

	"github.com/ymattw/googs"
)

var (
	clientID     = flag.String("c", "", "client ID")
	clientSecret = flag.String("s", "", "client secret")
	username     = flag.String("u", "", "username")
	password     = flag.String("p", "", "password")
)

func login() {
	if *clientID == "" || *username == "" || *password == "" {
		// Empty clientSecret is fine if the client type is public
		log.Fatal("Syntax: -c clientID [-s clientSecret] -u username -p password login")
	}

	client := googs.NewClient(*clientID, *clientSecret)
	if err := client.Login(*username, *password); err != nil {
		log.Fatal(err)
	}
	client.Save(*secretFile)
	log.Printf("Credentials wrote to %s", *secretFile)
}
