package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	. "github.com/aerospike-labs/stock-exchange/models"
	"log"
)

type Client struct {
	conn     *websocket.Conn
	messages chan interface{}
	done     chan bool
}

func NewClient(host string, port uint16) (*Client, error) {

	url := fmt.Sprintf("ws://%s:%d/conn", host, port)
	origin := fmt.Sprintf("http://%s", host)

	conn, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	client := &Client{
		conn:     conn,
		messages: make(chan interface{}, 1),
		done:     make(chan bool),
	}

	return client, nil
}

func (c *Client) Listen() {
	go c.listen()
}

func (c *Client) listen() {
	for {
		var message Message

		if err := websocket.JSON.Receive(c.conn, &message); err != nil {
			c.done <- true
			return
		}

		if message.IsResponse() {
			c.messages <- &Response{
				Result: message.Result,
				Error:  message.Error,
				Id:     message.Id,
			}
		} else if message.IsNotification() {
			c.messages <- &Notification{
				Method: message.Method,
				Params: message.Params,
			}
		} else {
			c.done <- true
			return
		}
	}
}

func (c *Client) Conn() *websocket.Conn {
	return c.conn
}

func (c *Client) Send(request *Request) error {
	return websocket.JSON.Send(c.conn, request)
}

func (c *Client) Close() {
	c.conn.Close()
	close(c.messages)
}
