package main

import (
	"github.com/aerospike-labs/stock-exchange/logging"

	"flag"
	"fmt"
	"runtime"
	// "sync"
)

var (
	logch   chan interface{} = make(chan interface{}, 1024)
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
	flag.BoolVar(&logging.Enabled, "verbose", logging.Enabled, "Enable verbose logging")
	flag.Parse()

	// Log listener
	go logging.Listen()
	defer logging.Close()

	// Announce we're running
	logging.Log("Broker is running")

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
