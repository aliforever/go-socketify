package socketify_test

import (
	"fmt"
	"github.com/aliforever/go-socketify"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClient(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful",
			args: args{
				address: "ws://127.0.0.1:8080/",
			},
			wantErr: false,
		},
	}

	go runServer()

	// time.Sleep(time.Second * 5)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan bool)

			got, _ := socketify.NewClient(tt.args.address)

			got.SetRawHandler(func(message []byte) {
				fmt.Println("received", string(message), "and closing...")
				got.Close(200, "close")
			})

			var receivedErr error

			got.SetOnClose(func(err error) {
				fmt.Println("received error", err)
				receivedErr = err
				ch <- true
			})

			<-ch
			assert.ErrorContains(t, receivedErr, "close")
		})
	}
}

func TestNewClientWrite(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful",
			args: args{
				address: "ws://127.0.0.1:8080/",
			},
			wantErr: false,
		},
	}

	go runServer()

	// time.Sleep(time.Second * 5)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := socketify.NewClient(tt.args.address)
			if err != nil {
				panic(err)
			}
			var wait = make(chan bool)

			var receivedErr error

			got.SetRawHandler(func(message []byte) {
				if string(message) == "Hello" {
					fmt.Println("received", string(message), "and closing...")
					got.Close(200, "bye")
				} else {
					fmt.Println("received", string(message))
				}
			})

			got.SetOnClose(func(err error) {
				receivedErr = err
				wait <- true
			})

			go got.WriteText("Hello")
			fmt.Println("waiting")
			<-wait
			assert.ErrorContains(t, receivedErr, "bye")
		})
	}
}

func runServer() {
	s := socketify.NewServer(socketify.ServerOptions().SetAddress(":8080").SetEndpoint("/"))

	go func() {
		err := s.Listen()
		if err != nil {
			panic(err)
		}
	}()

	for upgradeRequest := range s.UpgradeRequests() {
		client, err := upgradeRequest.Upgrade()
		if err != nil {
			panic(err)
		}
		client.HandleRawUpdate(func(message []byte) {
			fmt.Println(string(message))
			if string(message) == "Hello" {
				client.WriteText("Hello")
			}
		})
		go func(connection *socketify.Connection) {
			connection.WriteUpdate("client_id", connection.ID())
			err := connection.ProcessUpdates()
			if err != nil {
				fmt.Println(err)
				return
			}

		}(client)
	}
}
