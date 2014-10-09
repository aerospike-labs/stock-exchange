package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	. "github.com/aerospike-labs/stock-exchange/models"
	"log"
)

type WebSocketClient struct {
	conn     *websocket.Conn
	messages chan interface{}
	done     chan bool
}

func NewWebSocketClient(host string, port uint16) (*WebSocketClient, error) {

	url := fmt.Sprintf("ws://%s:%d/conn", host, port)
	origin := fmt.Sprintf("http://%s", host)

	conn, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	client := &WebSocketClient{
		conn:     conn,
		messages: make(chan interface{}, 1),
		done:     make(chan bool),
	}

	return client, nil
}

func (c *WebSocketClient) listen() {
	for {
		var raw []byte

		if err := websocket.Message.Receive(c.conn, &raw); err != nil {
			fmt.Printf("GOT ERROR %#v\n\n", err)
			c.done <- true
			return
		}

		msg := Unmarshal(raw)

		if msg == nil {
			c.done <- true
			return
		}

		logch <- msg
		c.messages <- msg
	}
}

func (c *WebSocketClient) send(request *Request) error {
	logch <- request
	return websocket.JSON.Send(c.conn, request)
}

func (c *WebSocketClient) close() {
	c.conn.Close()
	close(c.messages)
}
