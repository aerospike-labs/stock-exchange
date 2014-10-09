package main

import (
	// "code.google.com/p/go.net/websocket"
	"flag"
	// "fmt"
	. "github.com/aerospike-labs/stock-exchange/models"
	"log"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	exHost := flag.String("host", "localhost", "Exchange server address")
	exPort := flag.Int("port", 7000, "Exchange server port")
	// dbHost := flag.String("dbhost", "0.0.0.0", "Aerospike server address")
	// dbPort := flag.Int("dbport", 3000, "Aerospike server port")

	// parse flags
	flag.Parse()

	println("I'm a Broker\n\n")

	c, err := NewClient(*exHost, uint16(*exPort))
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()
	c.Listen()

	for i := 0; i < 1024; i++ {
		req := &Request{
			Method: "Arith.Multiply",
			Params: []interface{}{Args{A: i, B: 2}},
			Id:     uint64(i),
		}

		c.Send(req)
	}

	i := 0
	send(c, i)

	for {
		select {
		case message := <-c.messages:
			switch m := message.(type) {
			case *Response:
				log.Printf("Response: %#v \n", m)
				switch v := m.Result.(type) {
				case float64:
					// c.log <- "fuck"
					log.Printf("    sum: %f\n", v)
					i += 1
					send(c, i)
				default:
					log.Printf("    unknown value\n")
				}
			case *Notification:
				log.Printf("Notification: %#v \n", m)
			}
		case <-c.done:
			log.Printf("\n\nDONE!!!! \n\n")
			return
		}
	}
}

func send(c *Client, i int) {

	r := &Request{
		Method: "Arith.Multiply",
		Params: []interface{}{Args{A: i, B: 2}},
		Id:     uint64(i),
	}

	log.Printf("Request: %#v \n", r)
	c.Send(r)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
