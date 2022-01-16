package main

import (
	"encoding/json"
	"fmt"
	"github.com/aliforever/go-socketify"
)

func main() {
	server := socketify.New(socketify.Options().SetAddress(":8080").SetEndpoint("/ws").IgnoreCheckOrigin())
	go server.Listen()

	for client := range server.Clients() {
		client.HandleUpdate("PING", func(message json.RawMessage) {
			client.WriteUpdate("PONG", nil)
		})
		go client.ProcessUpdates()
		go func(c *socketify.Client) {
			for update := range c.Updates() {
				fmt.Println(update)
			}
		}(client)
	}
}
