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
			got := socketify.NewClient(tt.args.address)

			got.SetRawHandler(func(message []byte) {
				fmt.Println("received", string(message), "and closing...")
				got.Close()
			})

			err := got.Connect()
			assert.ErrorContains(t, err, "close")
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

	for client := range s.Clients() {
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
