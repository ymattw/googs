package main

import (
	"flag"
	"log"

	"github.com/ymattw/googs"
)

var (
	clientID = flag.String("c", "", "client ID")
	username = flag.String("u", "", "username")
	password = flag.String("p", "", "password")
)

func login() {
	if *clientID == "" || *username == "" || *password == "" {
		log.Fatal("Syntax: -c clientID -u username -p password login")
	}

	client := googs.NewClient(*clientID, "")
	if err := client.Login(*username, *password); err != nil {
		log.Fatal(err)
	}
	client.Save(*secretFile)
	log.Printf("Credentials wrote to %s", *secretFile)
}
