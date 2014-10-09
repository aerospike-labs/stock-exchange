package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	m "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"
	"net/http"
	_ "net/http/pprof"
	"net/rpc"
	"net/rpc/jsonrpc"
	"runtime"
)

var db *as.Client
var readPolicy *as.BasePolicy
var writePolicy *as.WritePolicy
var scanPolicy *as.ScanPolicy

const (
	NAMESPACE = "test"
	STOCKS    = "stocks"
	OFFERS    = "offers"
)

type Args struct {
	A int
	B int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	println(args.A, "+", args.B, "=", *reply)
	return nil
}

type Command struct{}

var counter int = 0

func (command *Command) Stocks(args *Args, reply *[]m.Stock) error {

	recordset, err := db.ScanAll(scanPolicy, NAMESPACE, STOCKS)
	if err != nil {
		return err
	}

	for rec := range recordset.Records {
		*reply = append(*reply, m.Stock{
			Ticker:   rec.Bins["ticker"].(string),
			Quantity: uint64(rec.Bins["quantity"].(int)),
			Price:    uint64(rec.Bins["price"].(int)),
		})
	}

	return nil
}

func (command *Command) Offers(args *Args, reply *[]m.Offer) error {

	recordset, err := db.ScanAll(scanPolicy, NAMESPACE, OFFERS)
	if err != nil {
		return err
	}

	for rec := range recordset.Records {
		*reply = append(*reply, m.Offer{
			BrokerId:  uint64(rec.Bins["broker_id"].(int)),
			OfferId:   uint64(rec.Bins["offer_id"].(int)),
			OfferType: m.OfferType(rec.Bins["offer_type"].(int)),
			TTL:       uint32(rec.Bins["ttl"].(int)),
			Ticker:    rec.Bins["ticker"].(string),
			Quantity:  uint64(rec.Bins["quantity"].(int)),
			Price:     uint64(rec.Bins["price"].(int)),
		})
	}

	return nil
}

func (command *Command) Sell(args *m.Offer, reply *bool) error {
	// Buy(ticker string, quantity uint64, price uint64, ttl uint32) (uint64, error) {

	// Check if the user has that much inventory

	key, _ := as.NewKey(NAMESPACE, OFFERS, fmt.Sprintf("%v:%v", args.BrokerId, args.OfferId))
	wpolicy := as.NewWritePolicy(0, 0)
	if args.TTL > 0 {
		wpolicy.Expiration = int(args.TTL)
	}
	return db.Put(writePolicy, key, as.BinMap{
		"broker_id":  args.BrokerId,
		"offer_id":   args.OfferId,
		"offer_type": int(args.OfferType),
		"ticker":     args.Ticker,
		"quantity":   args.Quantity,
		"price":      args.Price,
	})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var err error
	// connect to the db
	if db, err = as.NewClient("127.0.0.1", 3000); err != nil {
		panic(err)
	}
	readPolicy = as.NewPolicy()
	scanPolicy = as.NewScanPolicy()

	rpc.Register(new(Arith))
	rpc.Register(new(Command))

	http.Handle("/conn", websocket.Handler(serve))
	http.ListenAndServe("0.0.0.0:7000", nil)
}

func serve(ws *websocket.Conn) {
	jsonrpc.ServeConn(ws)
}
