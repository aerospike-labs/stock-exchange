package main

import (
	. "github.com/aerospike-labs/stock-exchange/client"
	"github.com/aerospike-labs/stock-exchange/logging"
	. "github.com/aerospike-labs/stock-exchange/models"

	"flag"
	"fmt"
	"runtime"
	// "sync"
)

var (
	verbose bool   = false
	broker  int    = 0
	exHost  string = "127.0.0.1"
	exPort  int    = 7000
	dbHost  string = "127.0.0.1"
	dbPort  int    = 3000
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// parse flags
	flag.IntVar(&broker, "broker", broker, "Broker Id")
	flag.StringVar(&exHost, "host", exHost, "Exchange host")
	flag.IntVar(&exPort, "port", exPort, "Exchange port")
	flag.StringVar(&dbHost, "dbhost", dbHost, "Aerospike host")
	flag.IntVar(&dbPort, "dbport", dbPort, "Aerospike port")
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
	ex, err := NewExchangeClient(broker, exHost, uint16(exPort))
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

				// TODO: Process of the Offer.
				//
				// Ideally, using the data avilable to you in
				// the bid, and your local database, determine
				// how to bid for the offer.
				//

			case "Bid":
				bid := message.Params.(Bid)

				// store the bid
				storeBid(&bid)

				// TODO: Processing of the Bid.
				//
				// Ideally, using the data available to you in
				// the bid, and your local database, determine
				// how to counter bid.
				//

			case "Close":
				bid := message.Params.(Bid)

				// store the winning bid
				storeWinningBid(&bid)

				// TODO: Process the closing of the offer
				//
				// Ideally, using the data from the winning bid,
				// you will know how to offer stocks of a similar price.
				//

			case "Cancel":
				// offerId := message.Params.(int)

				// TODO: Process the cancelled of the offer
				//
				// Ideally, using the data from the failed offering,
				// you will know what stock or price is not ideal.
				//

			}
		}
	}
}

// Custom initial logic
// This is not necessary, but is available
// for you to do something on startup
func run(ex *ExchangeClient) {

	// TODO: anything you want

}
