package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Broadcaster struct {
	send        chan interface{}
	connections map[int]*websocket.Conn
}

func NewBroadcaster(send chan interface{}) *Broadcaster {
	return &Broadcaster{
		send:        send,
		connections: make(map[int]*websocket.Conn, 1024),
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
			fmt.Printf("SENDING SOCKET MESSAGE %#v\n", message)
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
		fmt.Println(err)
		return
	}

	b.connections[1] = ws
}
