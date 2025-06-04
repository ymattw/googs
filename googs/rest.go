package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
)

func rest(args ...string) {
	if len(args) != 1 {
		log.Fatal("Syntax: rest <api>")
	}
	api := args[0]

	client := loadClient()
	res := make(map[string]any)
	err := client.Get(api, nil, &res)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", formatObject(res))
}

func formatObject(obj any) string {
	var out bytes.Buffer
	data, _ := json.Marshal(obj)
	if json.Indent(&out, []byte(data), "", "  ") != nil {
		return ""
	}
	return out.String()
}
