package main

import (
	"github.com/aerospike-labs/stock-exchange/logging"
	. "github.com/aerospike-labs/stock-exchange/models"

	"flag"
	"fmt"
	"runtime"
	// "sync"
)

var (
	logch   chan interface{} = make(chan interface{}, 1024)
	verbose bool             = false
	exHost  string           = "127.0.0.1"
	exPort  int              = 7000
	dbHost  string           = "127.0.0.1"
	dbPort  int              = 3000
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// parse flags
	flag.StringVar(&exHost, "host", exHost, "Exchange server address")
	flag.IntVar(&exPort, "port", exPort, "Exchange server port")
	flag.StringVar(&dbHost, "dbhost", dbHost, "Aerospike server address")
	flag.IntVar(&dbPort, "dbport", dbPort, "Aerospike server port")
	flag.BoolVar(&logging.Enabled, "verbose", logging.Enabled, "Enable verbose logging")
	flag.Parse()

	// Log listener
	go logging.Listen()
	defer logging.Close()

	// Announce we're running
	logging.Log("Broker is running")

	// Resuse error
	var err error

	// Connect to database
	connectToDatabase(dbHost, dbPort)

	// Connect to exchange
	ex, err := NewExchangeClient(1, exHost, uint16(exPort))
	if err != nil {
		fmt.Printf("err: %#v\n", err.Error())
	}

	// Close connections to the exchange
	defer ex.Close()

	// Open connections to the exchange
	go ex.Listen()

	// Process notifications
	go processNotifications(ex)

	// Run custom logic
	run(ex)

	// Wait for done to exit
	<-ex.Done
}

// Process Notifications
func processNotifications(ex *ExchangeClient) {
	for {
		select {
		case message := <-ex.Messages:
			switch message.Method {
			case "Offer":

				offer := message.Params.(Offer)

				// store the offer
				storeOffer(&offer)

				// additional processing of the offer

			case "Bid":
				bid := message.Params.(Bid)

				// store the bid
				storeBid(&bid)

				// additional processing of the bid

			case "Close":
				bid := message.Params.(Bid)

				// store the winning bid
				storeWinningBid(&bid)

				// additional processing of the close bid

			}
		}
	}
}

// Custom Logic
func run(ex *ExchangeClient) {

	ex.Offer("GOOG", 100, 1, 5)
}
