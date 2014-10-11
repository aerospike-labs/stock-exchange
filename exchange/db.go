package main

import (
	"fmt"
	m "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"
)

// name of database entities
const (
	NAMESPACE = "test"
	STOCKS    = "exchange:stocks"
	OFFERS    = "exchange:offers"
	BIDS      = "exchange:bids"
	BROKERS   = "exchange:brokers"
	SEQUENCES = "exchange:seq"
	LOG       = "log"
)

var (
	db          *as.Client
	readPolicy  *as.BasePolicy
	writePolicy *as.WritePolicy
	scanPolicy  *as.ScanPolicy
)

func nextSeq(seq string) (int, error) {

	keyId := fmt.Sprintf("%s:%s", SEQUENCES, seq)
	key, err := as.NewKey(NAMESPACE, SEQUENCES, keyId)
	if err != nil {
		return 0, err
	}

	rec, err := db.Operate(
		writePolicy,
		key,
		as.AddOp(as.NewBin("seq", 1)),
		as.GetOpForBin("seq"),
	)

	if err != nil {
		return 0, err
	}

	if seq, exists := rec.Bins["seq"]; !exists || int(seq.(int)) <= 0 {
		return 0, fmt.Errorf("sequence not found: %s", seq)
	}

	return rec.Bins["seq"].(int), nil
}

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

	// index on offer_id of bids, so we can find bids on a given offer
	if _, err := db.CreateIndex(writePolicy, "test", BIDS, "idx:ex:1", "offer_id", as.NUMERIC); err != nil {
		fmt.Printf("Create Index Failed: %s\n", err.Error())
	}

	// index on broker_id of bids, so we can find bids by a particular broker
	if _, err := db.CreateIndex(writePolicy, "test", BIDS, "idx:ex:2", "broker_id", as.NUMERIC); err != nil {
		fmt.Printf("Create Index Failed: %s\n", err.Error())
	}

	// index on broker_id of offers, so we can find offers by a particular broker
	if _, err := db.CreateIndex(writePolicy, "test", OFFERS, "idx:ex:3", "broker_id", as.NUMERIC); err != nil {
		fmt.Printf("Create Index Failed: %s\n", err.Error())
	}

}
