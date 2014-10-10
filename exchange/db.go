package main

import (
	m "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"
)

var db *as.Client
var readPolicy *as.BasePolicy
var writePolicy *as.WritePolicy
var scanPolicy *as.ScanPolicy

func findAuction(auctionId int) *m.Offer {
	auctionKey, _ := as.NewKey(NAMESPACE, AUCTIONS, auctionId)
	rec, err := db.Get(readPolicy, auctionKey)
	if err != nil {
		return nil
	}

	return &m.Offer{
		OfferId:  rec.Bins["auction_id"].(int),
		BrokerId: rec.Bins["broker_id"].(int),
		TTL:      rec.Bins["ttl"].(int),
		Ticker:   rec.Bins["ticker"].(string),
		Quantity: rec.Bins["quantity"].(int),
		Price:    rec.Bins["price"].(int),
	}
}

func findBid(bidId int) *m.Offer {
	auctionKey, _ := as.NewKey(NAMESPACE, BIDS, bidId)
	rec, err := db.Get(readPolicy, auctionKey)
	if err != nil {
		return nil
	}

	return &m.Offer{
		OfferId:  rec.Bins["bid_id"].(int),
		BrokerId: rec.Bins["broker_id"].(int),
		Price:    rec.Bins["price"].(int),
	}
}

func connectToDatabase(host string, port int) {
	var err error

	// connect to the db
	if db, err = as.NewClient(host, port); err != nil {
		panic(err)
	}

	readPolicy = as.NewPolicy()
	writePolicy = as.NewWritePolicy(0, 0)
	scanPolicy = as.NewScanPolicy()
}
