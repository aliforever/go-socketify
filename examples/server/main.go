package main

import (
	"encoding/json"
	"github.com/aliforever/go-socketify"
	"strings"
)

func main() {
	server := socketify.New(socketify.Options().SetAddress(":8080").SetEndpoint("/ws").IgnoreCheckOrigin())
	go server.Listen()

	for client := range server.Clients() {
		client.HandleRawUpdate(func(message json.RawMessage) {
			if strings.Contains(string(message), "close") {
				client.Close()
			}
		})
		go client.ProcessUpdates()
	}
}
