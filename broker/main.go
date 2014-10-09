package main

import (
	"flag"
	"fmt"
	. "github.com/aerospike-labs/stock-exchange/models"
	"log"
	"runtime"
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

	// Exchange client handles communitcation with the exchange
	ex, err := NewExchangeClient(1, *exHost, uint16(*exPort))
	if err != nil {
		log.Fatal(err)
	}

	// Close the client when the function exits
	defer ex.Close()

	// List for messages from the exchange
	ex.Listen()

	// First transaction is to get a list of stocks in the market
	ex.Stocks()

	for {
		select {
		case message := <-ex.Messages():
			switch m := message.(type) {
			case *Response:
				switch r := m.Result.(type) {
				case StockList:
					// A stock list is received, so we will process it
					logch <- fmt.Sprintf("Stocks: %#v", r)
					ex.Offers()
				case OfferList:
					// A stock list is received, so we will process it
					logch <- fmt.Sprintf("Offers: %#v", r)
					ex.Stocks()
				default:
					// Unhandled result type.
					// This should never be reached
					logch <- fmt.Sprintf("Unhandled: %#v", r)
				}
			case *Notification:
				logch <- m
			}
		case <-ex.Done():
			logch <- fmt.Sprintf("\n\nDONE!!!! \n\n")
			return
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
				log.Printf("<REQ> %d %s %#v", m.Id[0], m.Method, m.Params)
			case *Response:
				log.Printf("<RES> %d %#v %#v", m.Id[0], m.Result, m.Error)
			case *Notification:
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
