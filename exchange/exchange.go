package main

import (
	// "code.google.com/p/go.net/websocket"
	"errors"
	"flag"
	"fmt"
	m "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"
	rpc "github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
	"net/http"
	// _ "net/http/pprof"
	// "net/rpc"
	// "net/rpc/jsonrpc"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

var db *as.Client
var readPolicy *as.BasePolicy
var writePolicy *as.WritePolicy
var scanPolicy *as.ScanPolicy

var broadcast chan interface{} = make(chan interface{}, 1024)

const (
	NAMESPACE = "test"
	STOCKS    = "stocks"
	AUCTIONS  = "auctions"
	BIDS      = "bids"
	BROKERS   = "brokers"
	LOG       = "log"
)

type Command struct{}

var counter int = 0

func (command *Command) Stocks(r *http.Request, args *Command, reply *[]m.Stock) error {

	recordset, err := db.ScanAll(scanPolicy, NAMESPACE, STOCKS)
	if err != nil {
		return err
	}

	for rec := range recordset.Records {
		*reply = append(*reply, m.Stock{
			Ticker:   rec.Bins["ticker"].(string),
			Quantity: int(rec.Bins["quantity"].(int)),
			Price:    int(rec.Bins["price"].(int)),
		})
	}

	// fmt.Printf("STOCKS: %#v\n\n", *reply)

	broadcast <- &m.Notification{
		Version: "2.0",
		Method:  "Oh.Yeah",
		Params:  []interface{}{"OH YEAH! STOCKS"},
	}

	return nil
}

func (command *Command) Auctions(r *http.Request, args *Command, reply *[]m.Offer) error {

	recordset, err := db.ScanAll(scanPolicy, NAMESPACE, AUCTIONS)
	if err != nil {
		return err
	}

	for rec := range recordset.Records {
		*reply = append(*reply, m.Offer{
			BrokerId: int(rec.Bins["broker_id"].(int)),
			TTL:      uint32(rec.Bins["ttl"].(int)),
			Ticker:   rec.Bins["ticker"].(string),
			Quantity: int(rec.Bins["quantity"].(int)),
			Price:    int(rec.Bins["price"].(int)),
		})
	}

	// fmt.Printf("AUCTIONS: %#v\n\n", *reply)

	broadcast <- &m.Notification{
		Version: "2.0",
		Method:  "Oh.Yeah",
		Params:  []interface{}{"OH YEAH! AUCTIONS"},
	}

	return nil
}

var auctionId int64 = 0

func (command *Command) CreateAuction(r *http.Request, args *m.Offer, reply *bool) error {

	// Check if the broker has enough inventory
	brokerKey, _ := as.NewKey(NAMESPACE, BROKERS, int(args.BrokerId))
	rec, err := db.Get(readPolicy, brokerKey, args.Ticker, args.Ticker+"_os")
	if err != nil {
		return err
	}

	fmt.Printf("%#v\n\n", rec)

	if rec.Bins == nil || len(rec.Bins) == 0 {
		return errors.New("Broker does not have any inventory of the stock")
	} else if _, exists := rec.Bins[args.Ticker+"_os"]; !exists {
		// set the outstanding as much as the inventory - bin quantity
		err := db.Put(writePolicy, brokerKey, as.BinMap{args.Ticker + "_os": int(int(rec.Bins[args.Ticker].(int)) - args.Quantity)})
		if err != nil {
			return err
		}
	} else {
		rec, err := db.Operate(
			writePolicy,
			brokerKey,
			as.AddOp(as.NewBin(args.Ticker+"_os", -1*int(args.Quantity))),
			as.GetOp(),
		)
		if err != nil {
			return err
		}

		if inventory, exists := rec.Bins[args.Ticker+"_os"]; !exists || int(inventory.(int)) < args.Quantity {
			db.Add(writePolicy, brokerKey, as.BinMap{args.Ticker + "_os": args.Quantity})
			return errors.New("Not enough inventory")
		}
	}

	aId := atomic.AddInt64(&auctionId, 1)

	// put the offer up
	key, _ := as.NewKey(NAMESPACE, AUCTIONS, aId)
	if err := db.Put(writePolicy, key, as.BinMap{
		"auction_id": aId,
		"broker_id":  args.BrokerId,
		"ticker":     args.Ticker,
		"quantity":   args.Quantity,
		"price":      args.Price,
		"ttl":        args.TTL,
	}); err != nil {
		return err
	}

	go Auctioner(int(auctionId), args.TTL)
	return nil
}

var auctionMap AuctionMap

func Auctioner(auctionId int, TTL uint32) {
	bidderChan := auctionMap.Add(int64(auctionId))

	var bestBid *m.Offer
	for {
		select {
		case <-time.After(time.Duration(TTL) * time.Second):
			CloseAuction(auctionId, bestBid)
		case bid := <-bidderChan:
			if bestBid == nil || bestBid.Price > bid.Price {
				bestBid = bid
			}
		}
	}

}

var bidId int64 = 0

func (command *Command) Bid(r *http.Request, args *m.Offer, reply *bool) error {
	// check if auction exists
	offerKey, _ := as.NewKey(NAMESPACE, BIDS, args.OfferId)
	exists, err := db.Exists(readPolicy, offerKey)
	bidChan := auctionMap.Get(auctionId)
	if err != nil || !exists || bidChan == nil {
		errors.New("Auction has finished.")
	}

	bId := atomic.AddInt64(&bidId, 1)
	bidKey, _ := as.NewKey(NAMESPACE, BIDS, bId)
	if err := db.Put(writePolicy, bidKey, as.BinMap{
		"bid_id":     bId,
		"auction_id": args.OfferId,
		"broker_id":  args.BrokerId,
		"price":      args.Price,
	}); err != nil {
		return err
	}

	bidChan <- args
	return nil
}

func CloseAuction(auctionId int, bid *m.Offer) {
	auctionMap.Remove(int64(auctionId))

	// Check if the broker has enough inventory
	auctionKey, _ := as.NewKey(NAMESPACE, AUCTIONS, auctionId)
	auction, _ := db.Get(readPolicy, auctionKey)
	sellerId := auction.Bins["broker_id"].(int)

	// add to buyer's inventory and reduce credit
	keyBuyer, _ := as.NewKey(NAMESPACE, BROKERS, bid.BrokerId)
	db.Operate(writePolicy, keyBuyer,
		as.AddOp(as.NewBin(auction.Bins["ticker"].(string), auction.Bins["Quantity"].(int))),
		as.AddOp(as.NewBin("credit", -1*auction.Bins["Quantity"].(int)*int(bid.Price))),
	)

	// reduce seller's inventory and add to credit
	sellerKey, _ := as.NewKey(NAMESPACE, BROKERS, sellerId)
	db.Operate(writePolicy, sellerKey,
		as.AddOp(as.NewBin(auction.Bins["ticker"].(string), -1*auction.Bins["Quantity"].(int))),
		as.AddOp(as.NewBin("credit", auction.Bins["Quantity"].(int)*int(bid.Price))),
	)

	// mark the bid as winner
	bidKey, _ := as.NewKey(NAMESPACE, BIDS, bid.OfferId)
	db.Put(writePolicy, bidKey, as.BinMap{"winner": 1})
	db.Put(writePolicy, auctionKey, as.BinMap{"finished": 1, "winner": bid.OfferId})
}

func main() {
	var seed = flag.Bool("s", false, "seed db with data")
	var broker = flag.Int("b", 0, "add a broker")
	var brokerName = flag.String("bn", "", "broker name")
	var host = flag.String("h", "127.0.0.1", "Aerospike server seed hostnames or IP addresses")
	var port = flag.Int("p", 3000, "Aerospike server seed hostname or IP address port number.")

	listen := fmt.Sprintf("%s:%d", *host, 8080)
	var err error
	// connect to the db
	if db, err = as.NewClient(*host, *port); err != nil {
		panic(err)
	}

	flag.Parse()

	if *seed {
		seed_db()
		os.Exit(0)
	} else if *broker > 0 {
		seed_broker(*broker, *brokerName)
		os.Exit(0)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	readPolicy = as.NewPolicy()
	scanPolicy = as.NewScanPolicy()

	// ---
	//
	// START SERVER
	//
	// ---

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
