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
	logch   chan interface{} // global logging channel
	verbose bool             = false
	exHost  string           = "localhost"
	exPort  int              = 7000
	dbHost  string           = "localhost"
	dbPort  int              = 3000
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// parse flags
	flag.StringVar(&exHost, "host", exHost, "Exchange server address")
	flag.IntVar(&exPort, "port", exPort, "Exchange server port")
	flag.StringVar(&dbHost, "dbhost", dbHost, "Aerospike server address")
	flag.IntVar(&dbPort, "dbport", dbPort, "Aerospike server port")
	flag.BoolVar(&verbose, "verbose", verbose, "Enable verbose logging")
	flag.Parse()

	// Channel for receiving log messages
	// We initialize it, then listen for messages
	logch = make(chan interface{}, 1024)
	go listenLog()

	// Announce we're running
	logch <- "Broker is running"

	// Resuse error
	var err error

	ex, err := NewExchangeClient(1, exHost, uint16(exPort))
	if err != nil {
		fmt.Printf("err: %#v\n", err.Error())
	}

	// Close this on exit
	defer ex.Close()

	// Listen for notifications
	go ex.Listen()

	for i := 0; i < 1000; i++ {
		ex.Auctions()
	}
}

func listenLog() {
	for {

		select {
		case msg := <-logch:
			if verbose {
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
}
