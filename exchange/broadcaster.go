package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	// "io/ioutil"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Broadcaster struct {
	send        chan interface{}
	connections map[string]*websocket.Conn
}

func NewBroadcaster(send chan interface{}) *Broadcaster {
	return &Broadcaster{
		send:        send,
		connections: make(map[string]*websocket.Conn, 1024),
	}
}

func (b *Broadcaster) Listen() error {
	for {
		select {
		case m := <-b.send:
			message, err := json.Marshal(m)
			if err != nil {
				return err
			}
			b.Send(message)
		}
	}
	return nil
}

func (b *Broadcaster) Send(message []byte) error {
	for _, c := range b.connections {
		if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
			return err
		}
	}
	return nil
}

func (b *Broadcaster) Serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}

	b.connections[ws.RemoteAddr().String()] = ws
}
