package sockevent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client := NewClient(conn)

	assert.NotNil(t, client)
	assert.Equal(t, conn, client.Conn)
	assert.NotEmpty(t, client.ID)
	assert.NotNil(t, client.data)
	assert.NotNil(t, client.Ws)
}

func TestClient_SetAndGet(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client := NewClient(conn)

	client.Set("key", "value")
	assert.Equal(t, "value", client.Get("key"))

	client.Set("int", 123)
	assert.Equal(t, 123, client.Get("int"))

	assert.Nil(t, client.Get("nonexistent"))
}

func TestClient_SendJson(t *testing.T) {
	ws := NewWebsocket()
	ws.On("test", func(client *Client, message any) error {
		client.SendJson(Message{Command: "test", Message: message})
		return nil
	})
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client := NewClient(conn)

	message := Message{Command: "test", Message: "hello"}
	err = client.SendJson(message)
	assert.NoError(t, err)

	var receivedMessage Message
	err = conn.ReadJSON(&receivedMessage)
	assert.NoError(t, err)
	assert.Equal(t, message, receivedMessage)
}

func TestClient_Emit(t *testing.T) {
	ws := NewWebsocket()
	ws.On("test", func(client *Client, message any) error {
		client.Emit("test", message)
		return nil
	})
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client := NewClient(conn)

	err = client.Emit("test", "hello")
	assert.NoError(t, err)

	var receivedMessage Message
	err = conn.ReadJSON(&receivedMessage)
	assert.NoError(t, err)
	assert.Equal(t, Message{Command: "test", Message: "hello"}, receivedMessage)
}

func TestClient_EmitError(t *testing.T) {
	ws := NewWebsocket()
	ws.On("error", func(client *Client, message any) error {
		client.EmitError(message.(string))
		return nil
	})
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client := NewClient(conn)

	err = client.EmitError("test error")
	assert.NoError(t, err)

	var receivedMessage Message
	err = conn.ReadJSON(&receivedMessage)
	assert.NoError(t, err)
	assert.Equal(t, Message{Command: "error", Message: "test error"}, receivedMessage)
}

func TestClient_Close(t *testing.T) {
	ws := NewWebsocket()
	server := httptest.NewServer(http.HandlerFunc(ws.WsHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	client := NewClient(conn)
	client.Close()

	_, _, err = conn.ReadMessage()
	assert.Error(t, err)
}
