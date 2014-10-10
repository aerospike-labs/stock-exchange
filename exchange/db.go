package main

import (
	"fmt"
	m "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"
)

var db *as.Client
var readPolicy *as.BasePolicy
var writePolicy *as.WritePolicy
var scanPolicy *as.ScanPolicy

func findOffer(offerId int) *m.Offer {

	offerKeyId := fmt.Sprintf("%s:%d", OFFERS, offerId)
	offerKey, err := as.NewKey(NAMESPACE, OFFERS, offerKeyId)
	if err != nil {
		return nil
	}

	rec, err := db.Get(readPolicy, offerKey)
	if err != nil {
		return nil
	}

	return &m.Offer{
		Id:       rec.Bins["offer_id"].(int),
		BrokerId: rec.Bins["broker_id"].(int),
		TTL:      rec.Bins["ttl"].(int),
		Ticker:   rec.Bins["ticker"].(string),
		Quantity: rec.Bins["quantity"].(int),
		Price:    rec.Bins["price"].(int),
	}
}

func findBid(bidId int) *m.Bid {

	bidKeyId := fmt.Sprintf("%s:%d", BIDS, bidId)
	bidKey, err := as.NewKey(NAMESPACE, BIDS, bidKeyId)
	if err != nil {
		return nil
	}

	rec, err := db.Get(readPolicy, bidKey)
	if err != nil {
		return nil
	}

	return &m.Bid{
		Id:       rec.Bins["bid_id"].(int),
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
