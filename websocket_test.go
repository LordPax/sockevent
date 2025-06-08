package sockevent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewWebsocket(t *testing.T) {
	ws := NewWebsocket()
	assert.NotNil(t, ws)
	assert.NotNil(t, ws.clients)
	assert.NotNil(t, ws.events)
	assert.NotNil(t, ws.rooms)
	assert.NotNil(t, ws.onConnect)
	assert.NotNil(t, ws.onDisconnect)
}

func TestGetWebsocket(t *testing.T) {
	ws1 := GetWebsocket()
	ws2 := GetWebsocket()
	assert.Equal(t, ws1, ws2)
}

func TestWebsocket_OnConnect(t *testing.T) {
	ws := NewWebsocket()
	called := false
	ws.OnConnect(func(client *Client, w http.ResponseWriter, r *http.Request) error {
		called = true
		return nil
	})

	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWebsocket_OnDisconnect(t *testing.T) {
	ws := NewWebsocket()
	called := false
	ws.OnDisconnect(func(client *Client) error {
		called = true
		return nil
	})

	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	conn.Close()
	time.Sleep(100 * time.Millisecond)

	assert.True(t, called)
}

func TestWebsocket_On(t *testing.T) {
	ws := NewWebsocket()
	called := false
	ws.On("test", func(client *Client, message any) error {
		called = true
		return nil
	})

	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	err = conn.WriteJSON(Message{Command: "test", Message: "hello"})
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.True(t, called)
}

func TestWebsocket_GetClients(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	clients := ws.GetClients()
	assert.Equal(t, 1, len(clients))
}

func TestWebsocket_GetClient(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	clients := ws.GetClients()
	for id := range clients {
		client := ws.GetClient(id)
		assert.NotNil(t, client)
	}
}

func TestWebsocket_FindClient(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	client := ws.FindClient(func(c *Client) bool {
		return true
	})
	assert.NotNil(t, client)

	client = ws.FindClient(func(c *Client) bool {
		return false
	})
	assert.Nil(t, client)
}

func TestWebsocket_FilterClient(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	clients := ws.FilterClient(func(c *Client) bool {
		return true
	})
	assert.Equal(t, 1, len(clients))

	clients = ws.FilterClient(func(c *Client) bool {
		return false
	})
	assert.Equal(t, 0, len(clients))
}

func TestWebsocket_AddClient(t *testing.T) {
	ws := NewWebsocket()
	client := NewClient(nil)
	ws.AddClient(client)
	assert.Equal(t, 1, len(ws.GetClients()))
}

// TODO: Implement RemoveClient test
// func TestWebsocket_RemoveClient(t *testing.T) {
// 	ws := NewWebsocket()
// 	client := NewClient(nil)
// 	ws.AddClient(client)
// 	err := ws.RemoveClient(client.ID)
// 	assert.NoError(t, err)
// 	assert.Equal(t, 0, len(ws.GetClients()))
// }

func TestWebsocket_Close(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	clients := ws.GetClients()
	for id := range clients {
		ws.Close(id)
	}
	assert.Equal(t, 0, len(ws.GetClients()))
}

func TestWebsocket_SendJson(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	message := Message{Command: "test", Message: "hello"}
	err = ws.SendJson(message)
	assert.NoError(t, err)

	var receivedMessage Message
	err = conn.ReadJSON(&receivedMessage)
	assert.NoError(t, err)
	assert.Equal(t, message, receivedMessage)
}

func TestWebsocket_Emit(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	err = ws.Emit("test", "hello")
	assert.NoError(t, err)

	var receivedMessage Message
	err = conn.ReadJSON(&receivedMessage)
	assert.NoError(t, err)
	assert.Equal(t, "test", receivedMessage.Command)
	assert.Equal(t, "hello", receivedMessage.Message)
}

func TestWebsocket_EmitError(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	err = ws.EmitError("test error")
	assert.NoError(t, err)

	var receivedMessage Message
	err = conn.ReadJSON(&receivedMessage)
	assert.NoError(t, err)
	assert.Equal(t, "error", receivedMessage.Command)
	assert.Equal(t, "test error", receivedMessage.Message)
}

func TestWebsocket_Room(t *testing.T) {
	ws := NewWebsocket()
	room := ws.Room("test-room")
	assert.NotNil(t, room)
	assert.Equal(t, "test-room", room.Name)
}

func TestWebsocket_RemoveRoom(t *testing.T) {
	ws := NewWebsocket()
	ws.Room("test-room")
	assert.Equal(t, 1, len(ws.GetRooms()))
	ws.RemoveRoom("test-room")
	assert.Equal(t, 0, len(ws.GetRooms()))
}

func TestWebsocket_GetRooms(t *testing.T) {
	ws := NewWebsocket()
	ws.Room("room1")
	ws.Room("room2")
	rooms := ws.GetRooms()
	assert.Equal(t, 2, len(rooms))
	assert.NotNil(t, rooms["room1"])
	assert.NotNil(t, rooms["room2"])
}

func TestWebsocket_dispatchCommand(t *testing.T) {
	ws := NewWebsocket()
	called := false
	ws.On("test", func(client *Client, message any) error {
		called = true
		return nil
	})

	client := NewClient(nil)
	time.Sleep(100 * time.Millisecond)
	msg := Message{Command: "test", Message: "hello"}
	err := ws.dispatchCommand(client, msg)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWebsocket_listen(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	called := false
	ws.On("test", func(client *Client, message any) error {
		called = true
		return nil
	})

	err = conn.WriteJSON(Message{Command: "test", Message: "hello"})
	assert.NoError(t, err)

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	assert.True(t, called)
}

func TestWebsocket_connect(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(ws.GetClients()))

	conn.Close()
}

func TestWebsocket_WsHandler(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, len(ws.GetClients()))

	conn.Close()
}
