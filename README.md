# Go-Socketify
A simple WebSocket framework for Go

## Install
```go get -u github.com/aliforever/go-socketify```

## Usage
A simple app that PONG when PING
```go
options := socketify.ServerOptions().SetAddress(":8080").SetEndpoint("/ws").IgnoreCheckOrigin()
server := socketify.Server(options)
go server.Listen()

for connection := range server.Connections() {
    connection.HandleUpdate("PING", func(message json.RawMessage) {
        connection.WriteUpdate("PONG", nil)
    })
    go connection.ProcessUpdates()
}
```
Run the application and send below JSON to "ws://127.0.0.1:8080/ws":
```json
{
  "type": "PING"
}
```
You'll receive:
```json
{
  "type": "PONG"
}
```

## Conventions
Events are ought to be sent/received with following JSON format:
```json
{
  "type": "UpdateType",
  "data": {}
}
```
Type is going to be your update type and data is going to be anything.

## Storage
You can retrieve clients within other clients by using Socketify's client storage. 

You can enable the storage by setting option `EnableStorage()`:

```go
options := socketify.Options().
	SetAddress(":8080").
	SetEndpoint("/ws").
	IgnoreCheckOrigin().
	EnableStorage() // <-- This LINE
```
Each client has a unique ID set by [shortid](github.com/teris-io/shortid) package, you can recall using `client.ID()`.

Clients are stored in a map with their unique ID and you can retrieve them by calling:
```go
client.Server().Storage().GetClientByID(UniqueID)
```

## Handlers
You can specify a handler for an `updateType` to each client by using:
```go
client.HandleUpdate("UpdateType", func(message json.RawMessage) {
	// Process message
})
```
This way Socketify will call your registered handler if it receives any updates with `UpdateType` specified.

Or you can just listen on updates on your own:
```go
go client.ProcessUpdates()
go func(c *socketify.Client) {
    for update := range c.Updates() {
        fmt.Println(update)
    }
}(client)
```
Note: You should always call `go client.ProcessUpdates()` to let a Socketify client receive updates. 

## Docs:
Checkout Docs [Here](https://pkg.go.dev/github.com/aliforever/go-socketify)