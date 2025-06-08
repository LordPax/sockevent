package sockevent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewRoom(t *testing.T) {
	room := NewRoom("test-room")
	assert.NotNil(t, room)
	assert.Equal(t, "test-room", room.Name)
	assert.NotNil(t, room.clients)
	assert.Empty(t, room.clients)
}

func TestRoom_AddClient(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client := NewClient(conn)
	room := NewRoom("test-room")

	room.AddClient(client)
	assert.Equal(t, 1, len(room.clients))
	assert.Equal(t, client, room.clients[client.ID])
}

func TestRoom_RemoveClient(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client := NewClient(conn)
	room := NewRoom("test-room")

	room.AddClient(client)
	assert.Equal(t, 1, len(room.clients))

	room.RemoveClient(client.ID)
	assert.Equal(t, 0, len(room.clients))
}

func TestRoom_GetClients(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn1, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	conn2, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client1 := NewClient(conn1)
	client2 := NewClient(conn2)
	room := NewRoom("test-room")

	room.AddClient(client1)
	room.AddClient(client2)

	clients := room.GetClients()
	assert.Equal(t, 2, len(clients))
	assert.Equal(t, client1, clients[client1.ID])
	assert.Equal(t, client2, clients[client2.ID])
}

func TestRoom_GetClient(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client := NewClient(conn)
	room := NewRoom("test-room")

	room.AddClient(client)

	foundClient := room.GetClient(client.ID)
	assert.Equal(t, client, foundClient)

	notFoundClient := room.GetClient("non-existent-id")
	assert.Nil(t, notFoundClient)
}

func TestRoom_FindClient(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn1, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	conn2, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client1 := NewClient(conn1)
	client2 := NewClient(conn2)
	room := NewRoom("test-room")

	room.AddClient(client1)
	room.AddClient(client2)

	client1.Set("name", "Alice")
	client2.Set("name", "Bob")

	foundClient := room.FindClient(func(c *Client) bool {
		return c.Get("name") == "Alice"
	})
	assert.Equal(t, client1, foundClient)

	notFoundClient := room.FindClient(func(c *Client) bool {
		return c.Get("name") == "Charlie"
	})
	assert.Nil(t, notFoundClient)
}

func TestRoom_FilterClient(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn1, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	conn2, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	conn3, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client1 := NewClient(conn1)
	client2 := NewClient(conn2)
	client3 := NewClient(conn3)
	room := NewRoom("test-room")

	room.AddClient(client1)
	room.AddClient(client2)
	room.AddClient(client3)

	client1.Set("age", 25)
	client2.Set("age", 30)
	client3.Set("age", 35)

	filteredClients := room.FilterClient(func(c *Client) bool {
		age, ok := c.Get("age").(int)
		return ok && age > 28
	})
	assert.Equal(t, 2, len(filteredClients))
	assert.Contains(t, filteredClients, client2)
	assert.Contains(t, filteredClients, client3)
}

func TestRoom_SendJson(t *testing.T) {
	ws := NewWebsocket()
	ws.On("test", func(client *Client, message any) error {
		client.SendJson(Message{Command: "test", Message: message})
		return nil
	})
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn1, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	conn2, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client1 := NewClient(conn1)
	client2 := NewClient(conn2)
	room := NewRoom("test-room")

	room.AddClient(client1)
	room.AddClient(client2)

	message := Message{Command: "test", Message: "hello"}
	err = room.SendJson(message)
	assert.NoError(t, err)

	var receivedMessage1 Message
	err = conn1.ReadJSON(&receivedMessage1)
	assert.NoError(t, err)
	assert.Equal(t, message, receivedMessage1)

	var receivedMessage2 Message
	err = conn2.ReadJSON(&receivedMessage2)
	assert.NoError(t, err)
	assert.Equal(t, message, receivedMessage2)
}

func TestRoom_Emit(t *testing.T) {
	ws := NewWebsocket()
	ws.On("test", func(client *Client, message any) error {
		client.Emit("test", message)
		return nil
	})
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn1, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	conn2, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client1 := NewClient(conn1)
	client2 := NewClient(conn2)
	room := NewRoom("test-room")

	room.AddClient(client1)
	room.AddClient(client2)

	err = room.Emit("test", "hello")
	assert.NoError(t, err)

	var receivedMessage1 Message
	err = conn1.ReadJSON(&receivedMessage1)
	assert.NoError(t, err)
	assert.Equal(t, Message{Command: "test", Message: "hello"}, receivedMessage1)

	var receivedMessage2 Message
	err = conn2.ReadJSON(&receivedMessage2)
	assert.NoError(t, err)
	assert.Equal(t, Message{Command: "test", Message: "hello"}, receivedMessage2)
}

func TestRoom_EmitError(t *testing.T) {
	ws := NewWebsocket()
	ws.On("error", func(client *Client, message any) error {
		client.EmitError(message.(string))
		return nil
	})
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn1, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	conn2, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client1 := NewClient(conn1)
	client2 := NewClient(conn2)
	room := NewRoom("test-room")

	room.AddClient(client1)
	room.AddClient(client2)

	err = room.EmitError("test error")
	assert.NoError(t, err)

	var receivedMessage1 Message
	err = conn1.ReadJSON(&receivedMessage1)
	assert.NoError(t, err)
	assert.Equal(t, Message{Command: "error", Message: "test error"}, receivedMessage1)

	var receivedMessage2 Message
	err = conn2.ReadJSON(&receivedMessage2)
	assert.NoError(t, err)
	assert.Equal(t, Message{Command: "error", Message: "test error"}, receivedMessage2)
}
