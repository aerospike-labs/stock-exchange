package main

import (
	"errors"
	"fmt"
	m "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"
	"net/http"
	"sync/atomic"
)

// uniqueue Id to assign to Bids
var bidId int64 = 0

// unique Id to assign to auctions
var auctionId int64 = 0

type Command struct{}

// name of database entities
const (
	NAMESPACE = "test"
	STOCKS    = "stocks"
	AUCTIONS  = "auctions"
	BIDS      = "bids"
	BROKERS   = "brokers"
	LOG       = "log"
)

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
			TTL:      rec.Bins["ttl"].(int),
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
	broadcast <- &m.Broadcast{
		Type:    m.AUCTION_ENDED,
		Auction: *findAuction(args.OfferId),
		Bid:     *findBid(int(bId)),
	}
	return nil
}
