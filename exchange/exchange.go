package main

import (
	"flag"
	"fmt"
	rpc "github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

// broadcast channel
var broadcast chan interface{} = make(chan interface{}, 1024)

func main() {
	var seed = flag.Bool("s", false, "seed db with data and exit")
	var broker = flag.Int("b", 0, "add a broker and exit")
	var brokerName = flag.String("bn", "", "broker name")
	var host = flag.String("h", "127.0.0.1", "Aerospike server seed hostnames or IP addresses")
	var port = flag.Int("p", 3000, "Aerospike server seed hostname or IP address port number.")

	listen := fmt.Sprintf("%s:%d", *host, 7000)

	// parse flags
	flag.Parse()

	// defined in db.gp
	connectToDatabase(*host, *port)

	if *seed {
		seed_db()
		os.Exit(0)
	} else if *broker > 0 {
		seed_broker(*broker, *brokerName)
		os.Exit(0)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	///////////////////////////////////////////////////////////////////////////////////
	//
	// START SERVER
	//
	///////////////////////////////////////////////////////////////////////////////////

	// Use this for broadcasting messages to all brokers
	broadcaster := NewBroadcaster(broadcast)
	go broadcaster.Listen()

	// services
	command := new(Command)

	// export services
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterService(command, "")

	// routes
	httpRouter := http.NewServeMux()
	httpRouter.Handle("/rpc", rpcServer)
	httpRouter.HandleFunc("/ws", broadcaster.Serve)

	// server
	httpServer := &http.Server{
		Addr:           listen,
		Handler:        httpRouter,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   1 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// start
	log.Printf("Starting HTTP on http://%s\n", listen)
	fmt.Fprintf(os.Stdout, "Starting HTTP on http://%s\n", listen)

	log.Panic(httpServer.ListenAndServe())
}
