package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// A user has:
// - user id
// - lobbyToWeb channel
// - webToLobby channel
// - endConnection channel
// - websocket connection object
type user struct {
	id         string
	lobbyToWeb chan []byte
	webToLobby chan []byte
	endConn    chan string
	c          *websocket.Conn
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// User creation function:
// - takes in the user id
// - initializes channels
// - sets connection
// - sets user id
func createUser(id string) user {
	return user{
		id:         id,
		lobbyToWeb: make(chan []byte),
		endConn:    nil,
		c:          nil,
	}
}

// Main handler function:
// - is a method on a user struct
// - sets up connection
// - starts reader and writer functions then ends
func (u *user) runUser(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Upgrading request...")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade problem:", err)
		return
	}
	fmt.Println("Upgraded")

	u.c = c
	go u.webReader()
	go u.webWriter()
}

// WebReader function:
// - takes a pointer to the connection object and the webToLobby channel
// - waits for messages from the websocket
// - passes them to the channel
// - if websocket is closed, send 'close' message to lobby and return
func (u *user) webReader() {
	defer u.c.Close()
	for {
		_, message, err := u.c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("web reader error:", err)
			}
			break
		}
		fmt.Println("Read message from user:", string(message))
		u.webToLobby <- message
	}
}

// WebWriter function:
// - takes a pointer to the connection object and the lobbyToWeb channel
// - selects on the lobbyToWeb channel and endConnection channel
// - takes messages and writes them to websocket connection
// - if endConnection triggers, return
func (u *user) webWriter() {
	defer u.c.Close()
	for {
		select {
		case message := <-u.lobbyToWeb:
			w, err := u.c.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			fmt.Print("writing message...", string(message))
			w.Write(message)
			if err := w.Close(); err != nil {
				fmt.Println("writer error:", err)
			}
			fmt.Println("...finished writing")
		case <-u.endConn:
			return
		}
	}
}
