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

var (
	// broadcast channel
	broadcast chan interface{} = make(chan interface{}, 1024)

	seed   bool   = false
	addr   string = "127.0.0.1"
	port   int    = 7000
	dbHost string = "127.0.0.1"
	dbPort int    = 3000
)

func main() {

	// parse flags
	flag.BoolVar(&seed, "s", seed, "seed db with data and exit")
	flag.StringVar(&addr, "addr", addr, "Exchange listening address")
	flag.IntVar(&port, "port", port, "Exchange listening port")
	flag.StringVar(&dbHost, "dbhost", dbHost, "Aerospike host")
	flag.IntVar(&dbPort, "dbport", dbPort, "Aerospike port")
	flag.Parse()

	listen := fmt.Sprintf("%s:%d", addr, port)

	// defined in db.gp
	connectToDatabase(dbHost, dbPort)

	if seed {
		seed_db()
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
