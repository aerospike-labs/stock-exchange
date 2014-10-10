package main

import (
	"flag"
	"fmt"
	. "github.com/aerospike-labs/stock-exchange/models"
	"log"
	"runtime"
	// "sync"
)

var (
	logch chan interface{} // global logging channel
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	exHost := flag.String("host", "localhost", "Exchange server address")
	exPort := flag.Int("port", 7000, "Exchange server port")
	// dbHost := flag.String("dbhost", "0.0.0.0", "Aerospike server address")
	// dbPort := flag.Int("dbport", 3000, "Aerospike server port")

	// parse flags
	flag.Parse()

	// Channel for receiving log messages
	// We initialize it, then listen for messages
	logch = make(chan interface{}, 1024)
	go listenLog()

	// Announce we're running
	logch <- "Broker is running"

	// Resuse error
	var err error

	ex, err := NewExchangeClient(1, *exHost, uint16(*exPort))
	if err != nil {
		fmt.Printf("err: %#v\n", err.Error())
	}

	// Close this on exit
	defer ex.Close()

	// Listen for notifications
	go ex.Listen()

	for i := 0; i < 1000; i++ {
		a, err := ex.Auctions()
		if err != nil {
			fmt.Printf("err: %#v\n", err.Error())
		} else {
			fmt.Printf("res: %#v\n", a)
		}
	}
}

func listenLog() {
	for {

		select {
		case msg := <-logch:
			switch m := msg.(type) {
			case string:
				log.Println(m)
			case *Request:
				log.Printf("<REQ> %d %s %#v", m.Id, m.Method, m.Params)
			case *RawRequest:
				log.Printf("<REQ> %d %s %#v", m.Id, m.Method, m.Params)
			case *Response:
				log.Printf("<RES> %d %#v %#v", m.Id, m.Result, m.Error)
			case *RawResponse:
				log.Printf("<RES> %d %#v %#v", m.Id, m.Result, m.Error)
			case *Notification:
				log.Printf("<NOT> %s %#v", m.Method, m.Params)
			case *RawNotification:
				log.Printf("<NOT> %s %#v", m.Method, m.Params)
			}
		}
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
