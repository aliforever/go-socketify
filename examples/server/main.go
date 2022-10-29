package main

import (
	"fmt"
	"github.com/aliforever/go-socketify"
	"strings"
	"time"
)

func main() {
	server := socketify.NewServer(socketify.ServerOptions().SetAddress(":8080").SetEndpoint("/").IgnoreCheckOrigin())
	go server.Listen()

	for client := range server.Clients() {
		client.HandleRawUpdate(func(message []byte) {
			if strings.Contains(string(message), "close") {
				err := client.WriteUpdate("closing_connection", "closing")
				if err != nil {
					fmt.Println(err)
					return
				}

				err = client.WriteBinaryBytes([]byte("hello world, raw bytes"))
				if err != nil {
					fmt.Println(err)
					return
				}

				err = client.WriteRawUpdate(map[string]any{
					"action": "closing",
				})
				if err != nil {
					fmt.Println(err)
					return
				}

				err = client.WriteText("Hello World String")
				if err != nil {
					fmt.Println(err)
					return
				}

				time.Sleep(time.Second * 1)
				client.Close()

				err = client.WriteText("testing")
				if err != nil {
					fmt.Println(err, "error writing")
				}
			}
		})
		// client.SetKeepAliveDuration(time.Second * 5)
		go client.ProcessUpdates()
	}
}
