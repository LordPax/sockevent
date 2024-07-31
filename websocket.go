package sockevent

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var WebsocketInstance *Websocket

type EventFunc func(*Client, any) error
type OnConnectFunc func(*Client, http.ResponseWriter, *http.Request) error
type OnDisconnectFunc func(*Client) error
type CbClient func(*Client) bool

// Websocket is the main struct for managing websocket
type Websocket struct {
	clients      map[string]*Client
	events       map[string]EventFunc
	rooms        map[string]*Room
	onConnect    OnConnectFunc
	onDisconnect OnDisconnectFunc
}

// NewWebsocket creates a new Websocket instance
func NewWebsocket() *Websocket {
	return &Websocket{
		clients: make(map[string]*Client),
		events:  make(map[string]EventFunc),
		rooms:   make(map[string]*Room),
		onConnect: func(client *Client, w http.ResponseWriter, r *http.Request) error {
			return nil
		},
		onDisconnect: func(client *Client) error {
			return nil
		},
	}
}

// GetWebsocket returns the singleton instance of Websocket
//
// ws := GetWebsocket()
func GetWebsocket() *Websocket {
	if WebsocketInstance == nil {
		WebsocketInstance = NewWebsocket()
	}
	return WebsocketInstance
}

// connect creates a new websocket connection
func (w *Websocket) connect(wr http.ResponseWriter, r *http.Request) (*Client, error) {
	conn, err := upgrader.Upgrade(wr, r, nil)
	if err != nil {
		return nil, err
	}

	client := NewClient(conn)
	w.AddClient(client)

	if err := w.onConnect(client, wr, r); err != nil {
		return nil, err
	}

	return client, nil
}

// listen listens for incoming messages from the client
func (w *Websocket) listen(client *Client) error {
	defer w.Close(client.ID)

	for {
		var msg Message

		if err := client.Conn.ReadJSON(&msg); err != nil {
			return err
		}

		if err := w.dispatchCommand(client, msg); err != nil {
			return err
		}
	}
}

// dispatchCommand dispatches a command to the appropriate callback function
func (w *Websocket) dispatchCommand(client *Client, msg Message) error {
	callback, ok := w.events[msg.Command]
	if !ok {
		msgErr := fmt.Sprintf("Command \"%s\" not found", msg.Command)
		return client.EmitError(msgErr)
	}

	if err := callback(client, msg.Message); err != nil {
		return client.EmitError(err.Error())
	}

	return nil
}

// WsHandler is the handler for the websocket connection
func (w *Websocket) WsHandler(wr http.ResponseWriter, r *http.Request) {
	client, err := w.connect(wr, r)
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		_, _ = wr.Write([]byte(err.Error()))
		return
	}

	if err := w.listen(client); err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		_, _ = wr.Write([]byte(err.Error()))
		return
	}
}

// OnConnect sets the callback function to be called when a client connects
//
// ws.OnConnect(function)
func (w *Websocket) OnConnect(callback OnConnectFunc) {
	w.onConnect = callback
}

// OnDisconnect sets the callback function to be called when a client disconnects
//
// ws.OnDisconnect(function)
func (w *Websocket) OnDisconnect(callback OnDisconnectFunc) {
	w.onDisconnect = callback
}

// On sets a callback function to be called when a specific command is received
//
// ws.On("event", function)
func (w *Websocket) On(command string, callback EventFunc) {
	w.events[command] = callback
}

// GetClients returns the list of connected clients
//
// ws.GetClients()
func (w *Websocket) GetClients() map[string]*Client {
	return w.clients
}

// GetClient returns a client by its ID
//
// ws.GetClient(client.ID)
func (w *Websocket) GetClient(id string) *Client {
	return w.clients[id]
}

// FindClient returns a client by a callback function
//
//	ws.FindClient(func(client *Client) bool {
//		if !c.Get("logged").(bool) {
//			return false
//		}
//		return c.Get("user").(models.User).ID == 2
//	})
func (w *Websocket) FindClient(cb CbClient) *Client {
	for _, client := range w.clients {
		if cb(client) {
			return client
		}
	}

	return nil
}

// FilterClient returns a list of clients by a callback function
//
//	ws.FilterClient(func(client *Client) bool {
//		return c.Get("logged").(bool)
//	})
func (w *Websocket) FilterClient(cb CbClient) []*Client {
	var clients []*Client

	for _, client := range w.clients {
		if cb(client) {
			clients = append(clients, client)
		}
	}

	return clients
}

// AddClient adds a new client to the list of connected clients
//
// ws.AddClient(client)
func (w *Websocket) AddClient(client *Client) {
	w.clients[client.ID] = client
}

// RemoveClient removes a client from the list of connected clients
//
// ws.RemoveClient(client.ID)
func (w *Websocket) RemoveClient(id string) error {
	client := w.GetClient(id)
	if client == nil {
		return fmt.Errorf("Client not found\n")
	}

	for _, room := range w.rooms {
		room.RemoveClient(id)
	}

	client.Close()
	delete(w.clients, id)

	return nil
}

// Close closes the connection of a client
//
// ws.Close(client.ID)
func (w *Websocket) Close(id string) {
	client := w.GetClient(id)

	if err := w.RemoveClient(id); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

	if err := w.onDisconnect(client); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

// SendJson sends a message to all connected clients
//
// ws.SendJson(message)
func (w *Websocket) SendJson(message Message) error {
	for _, client := range w.clients {
		if err := client.SendJson(message); err != nil {
			return err
		}
	}

	return nil
}

// Emit sends a message to all connected clients
//
// ws.Emit("event", "message to send")
// client.Ws.Emit("event", []string{"message1", "message2"})
func (w *Websocket) Emit(command string, message any) error {
	wsMessage := Message{
		Command: command,
		Message: message,
	}

	return w.SendJson(wsMessage)
}

// EmitError sends an error message to all connected clients
//
// ws.EmitError("error message")
// client.Ws.EmitError("error message")
func (w *Websocket) EmitError(message string) error {
	wsMessage := Message{
		Command: "error",
		Message: message,
	}

	return w.SendJson(wsMessage)
}

// Room return and creates a new room if it doesn't exist
//
// ws.Room("room-name")
// client.Ws.Room("room-name")
func (w *Websocket) Room(name string) *Room {
	if _, ok := w.rooms[name]; !ok {
		w.rooms[name] = NewRoom(name)
	}

	return w.rooms[name]
}

// RemoveRoom removes a room by its name
//
// ws.RemoveRoom("room-name")
// client.Ws.RemoveRoom("room-name")
func (w *Websocket) RemoveRoom(name string) {
	delete(w.rooms, name)
}

// GetRooms returns the list of rooms
//
// ws.GetRooms()
// client.Ws.GetRooms()
func (w *Websocket) GetRooms() map[string]*Room {
	return w.rooms
}
