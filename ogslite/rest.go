package main

import (
	"fmt"
	"os"
)

func rest(args ...string) {
	if len(args) != 1 {
		fmt.Printf("Syntax: rest <api>\n")
		os.Exit(1)
	}
	api := args[0]

	client := loadClient()
	res := make(map[string]any)
	err := client.Get(api, nil, &res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", formatObject(res))
}
