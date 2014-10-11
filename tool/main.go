package main

import (
	. "github.com/aerospike-labs/stock-exchange/client"

	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	broker int    = 1
	host   string = "127.0.0.1"
	port   int    = 7000
)

func main() {

	flag.IntVar(&broker, "broker", broker, "Broker Id")
	flag.StringVar(&host, "host", host, "Exchange host")
	flag.IntVar(&port, "port", port, "Exchange port")
	flag.Parse()

	command := flag.Arg(0)

	switch strings.ToLower(command) {
	case "offer":
		offer()
	case "offers":
		offers()
	case "bid":
		bid()
	case "bids":
		bids()
	case "add-broker":
		addBroker()
	default:
		fmt.Println("Please provide a valid command")
	}

	os.Exit(1)
}

func offer() {

	var err error

	// Connect to exchange
	ex, err := NewExchangeClient(broker, host, uint16(port))
	if err != nil {
		fmt.Printf("err: %#v\n", err.Error())
	}

	// Close connections to the exchange
	defer ex.Close()

	ticker := flag.Arg(1)

	quantity, err := strconv.Atoi(flag.Arg(2))
	exitOnError(err)

	price, err := strconv.Atoi(flag.Arg(3))
	exitOnError(err)

	ttl, err := strconv.Atoi(flag.Arg(4))
	exitOnError(err)

	offerId, err := ex.Offer(ticker, quantity, price, ttl)
	exitOnError(err)

	fmt.Println(offerId)
}

func offers() {

	var err error

	// Connect to exchange
	ex, err := NewExchangeClient(broker, host, uint16(port))
	if err != nil {
		fmt.Printf("err: %#v\n", err.Error())
	}

	// Close connections to the exchange
	defer ex.Close()

	offers, err := ex.Offers()
	exitOnError(err)

	for i, offer := range offers {
		fmt.Printf("%10d  %#v\n", i, offer)
	}
}

func bid() {

	var err error

	// Connect to exchange
	ex, err := NewExchangeClient(broker, host, uint16(port))
	if err != nil {
		fmt.Printf("err: %#v\n", err.Error())
	}

	// Close connections to the exchange
	defer ex.Close()

	offerId, err := strconv.Atoi(flag.Arg(1))
	exitOnError(err)

	price, err := strconv.Atoi(flag.Arg(2))
	exitOnError(err)

	bidId, err := ex.Bid(offerId, price)
	exitOnError(err)

	fmt.Println(bidId)
}

func bids() {

	var err error

	// Connect to exchange
	ex, err := NewExchangeClient(broker, host, uint16(port))
	if err != nil {
		fmt.Printf("err: %#v\n", err.Error())
	}

	// Close connections to the exchange
	defer ex.Close()

	offerId, err := strconv.Atoi(flag.Arg(1))
	exitOnError(err)

	bids, err := ex.Bids(offerId)
	exitOnError(err)

	for i, bid := range bids {
		fmt.Printf("%10d  %#v\n", i, bid)
	}
}

func addBroker() {

	var err error

	// Connect to exchange
	ex, err := NewExchangeClient(broker, host, uint16(port))
	if err != nil {
		fmt.Printf("err: %#v\n", err.Error())
	}

	// Close connections to the exchange
	defer ex.Close()

	id, err := strconv.Atoi(flag.Arg(1))
	exitOnError(err)

	name := flag.Arg(1)

	credit, err := strconv.Atoi(flag.Arg(1))
	exitOnError(err)

	added, err := ex.AddBroker(id, name, credit)
	exitOnError(err)

	if !added {
		os.Exit(1)
	} else {
		fmt.Println("broker added.")
	}
}

func exitOnError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
