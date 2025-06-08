# sockevent

## Description

A simple and powerful Go library for managing WebSocket connections, clients, and rooms.

## Features

- Easy WebSocket connection management
- Customizable event system
- Client management with data storage
- Room creation and management

## Installation

```bash
go get github.com/LordPax/sockevent
```

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/your-username/go-websocket-lib/sockevent"
)

func main() {
    ws := sockevent.GetWebsocket()

    // Handle incoming messages
    ws.On("message", func(client *sockevent.Client, message any) error {
        // Process the message
        return client.Emit("response", "Message received!")
    })

    // Set up HTTP handler for WebSocket
    http.HandleFunc("/ws", ws.WsHandler)

    // Start the server
    http.ListenAndServe(":8080", nil)
}
```

## Detailed Features

### Connection Management

```go
ws := sockevent.GetWebsocket()
ws.OnConnect(func(client *sockevent.Client, w http.ResponseWriter, r *http.Request) error {
    // Connection logic
    return nil
})
ws.OnDisconnect(func(client *sockevent.Client) error {
    // Disconnection logic
    return nil
})
```

### Event Handling

```go
ws.On("customEvent", func(client *sockevent.Client, data any) error {
    // Handle custom event
    return nil
})
```

### Client Manipulation

```go
// Send a message to a client
client.Emit("event", "Data")

// Store data for a client
client.Set("userID", 123)

// Retrieve data from a client
userID := client.Get("userID")
```

### Room Management

```go
// Create or get a room
room := ws.Room("room-name")

// Add a client to a room
room.AddClient(client)

// Send a message to all clients in a room
room.Emit("announcement", "Message for everyone in the room")
```
