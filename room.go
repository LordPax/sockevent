package sockevent

// Room is the struct for managing a room of clients
type Room struct {
	Name    string
	clients map[string]*Client
}

// NewRoom creates a new room
func NewRoom(name string) *Room {
	return &Room{
		Name:    name,
		clients: make(map[string]*Client),
	}
}

// AddClient adds a client to the room
//
// room.AddClient(client)
// client.Ws.Room("room-name").AddClient(client)
func (r *Room) AddClient(client *Client) {
	r.clients[client.ID] = client
}

// RemoveClient removes a client from the room
//
// room.RemoveClient(client.ID)
// client.Ws.Room("room-name").RemoveClient(client.ID)
func (r *Room) RemoveClient(id string) {
	delete(r.clients, id)
}

// GetClients returns the list of clients in the room
//
// room.GetClients()
// clients := client.Ws.Room("room-name").GetClients()
func (r *Room) GetClients() map[string]*Client {
	return r.clients
}

// GetClient returns a client by its ID
//
// client := room.GetClient(client.ID)
// client := client.Ws.Room("room-name").GetClient(client.ID)
func (r *Room) GetClient(id string) *Client {
	return r.clients[id]
}

// FindClient returns a client by a callback function
//
//	room.FindClient(func(client *Client) bool {
//		if !c.Get("logged").(bool) {
//			return false
//		}
//		return c.Get("user").(models.User).ID == 2
//	})
func (r *Room) FindClient(cb CbClient) *Client {
	for _, client := range r.clients {
		if cb(client) {
			return client
		}
	}

	return nil
}

// FilterClient returns a list of clients by a callback function
//
//	room.FilterClient(func(client *Client) bool {
//		return c.Get("logged").(bool)
//	})
func (r *Room) FilterClient(cb CbClient) []*Client {
	var clients []*Client

	for _, client := range r.clients {
		if cb(client) {
			clients = append(clients, client)
		}
	}

	return clients
}

// SendJson sends a message to all clients in the room
func (r *Room) SendJson(message Message) error {
	for _, client := range r.clients {
		if err := client.SendJson(message); err != nil {
			return err
		}
	}

	return nil
}

// Emit sends a message to all clients in the room
//
// room.Emit("event", "message to send")
// client.Ws.Room("room-name").Emit("event", []string{"message1", "message2"})
func (r *Room) Emit(command string, message any) error {
	msg := Message{
		Command: command,
		Message: message,
	}

	return r.SendJson(msg)
}

// EmitError sends an error message to all clients in the room
//
// room.EmitError("error message")
// client.Ws.Room("room-name").EmitError("error message")
func (r *Room) EmitError(message string) error {
	errorMsg := Message{
		Command: "error",
		Message: message,
	}

	return r.SendJson(errorMsg)
}
