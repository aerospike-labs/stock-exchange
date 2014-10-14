package main

import (
	. "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"

	"fmt"
	"time"
)

const (
	NAMESPACE  = "test"
	BIDS       = "broker:bids"
	OFFERS     = "broker:offers"
	TICKERS    = "broker:tickers"
	PRICES     = "broker:prices"
	PORTFOLIOS = "broker:portfolios"
)

var db *as.Client
var readPolicy *as.BasePolicy
var writePolicy *as.WritePolicy
var scanPolicy *as.ScanPolicy
var queryPolicy *as.QueryPolicy

// Store an offer in the database
func storeOffer(offer *Offer) error {

	var err error

	keyId := fmt.Sprintf("%s:%d", OFFERS, offer.Id)
	key, err := as.NewKey(NAMESPACE, OFFERS, keyId)
	if err != nil {
		return err
	}

	rec := as.BinMap{
		"offer_id":  offer.Id,
		"broker_id": offer.BrokerId,
		"ttl":       offer.TTL,
		"ticker":    offer.Ticker,
		"quantity":  offer.Quantity,
		"price":     offer.Price,
	}

	// TODO: use the Put Operation to store the record.

	return nil
}

// Store a bid in the database
func storeBid(bid *Bid) error {

	var err error

	keyId := fmt.Sprintf("%s:%d", BIDS, bid.Id)
	key, err := as.NewKey(NAMESPACE, BIDS, keyId)
	if err != nil {
		return err
	}

	rec := as.BinMap{
		"bid_id":    bid.Id,
		"broker_id": bid.BrokerId,
		"offer_id":  bid.OfferId,
		"price":     bid.Price,
	}

	// TODO: use the Put Operation to store the record.

	return nil
}

// The winning bid for an offer.
// This bid will be stored as the new price for the ticker
// To get the ticker, you need to lookup the offer.
// Also, we will maintain the portfolio for other borkers, so we
// have an idea of what each has.
func storeWinningBid(bid *Bid) error {

	var err error

	// The current time is used a couple places
	ts := time.Now().Unix()

	// Read the offer, to get the ticker_id

	offerKeyId := fmt.Sprintf("%s:%d", OFFERS, bid.OfferId)
	offerKey, err := as.NewKey(NAMESPACE, OFFERS, offerKeyId)
	if err != nil {
		return err
	}

	offerRec, err := db.Get(readPolicy, offerKey)
	if err != nil {
		return err
	}

	if offerRec == nil {
		return fmt.Errorf("Record not found %#v", offerKey)
	}

	ticker := offerRec.Bins["ticker"].(string)
	quantity := offerRec.Bins["quantity"].(int)
	sellerId := offerRec.Bins["broker_id"].(int)

	// Update the current ticker price

	tickerKeyId := fmt.Sprintf("%s:%d", TICKERS, ticker)
	tickerKey, err := as.NewKey(NAMESPACE, TICKERS, tickerKeyId)
	if err != nil {
		return err
	}

	tickerBins := as.BinMap{
		"ticker": ticker,
		"price":  bid.Price,
		"time":   ts,
	}

	// TODO: use the Put operation to update the ticker latest price

	// Store the ticker price for historical prices
	// There is an index on ticker

	priceKeyId := fmt.Sprintf("%s:%d:%d", PRICES, ticker, ts)
	priceKey, err := as.NewKey(NAMESPACE, PRICES, priceKeyId)
	if err != nil {
		return err
	}

	priceBins := as.BinMap{
		"ticker": ticker,
		"price":  bid.Price,
		"time":   ts,
	}

	// TODO: Use the Put operation to store Record the price change

	// Update Porfolio

	sellerKeyId := fmt.Sprintf("%s:%d", PORTFOLIOS, sellerId)
	sellerKey, err := as.NewKey(NAMESPACE, PORTFOLIOS, sellerKeyId)
	if err != nil {
		return err
	}

	buyerKeyId := fmt.Sprintf("%s:%d", PORTFOLIOS, bid.BrokerId)
	buyerKey, err := as.NewKey(NAMESPACE, PORTFOLIOS, buyerKeyId)
	if err != nil {
		return err
	}

	sellerBins := as.BinMap{
		ticker: quantity,
	}

	buyerBins := as.BinMap{
		ticker: quantity,
	}

	// TODO: To do, use the Add operationg to increment the portfolios
	// of the buyer and seller

	return nil
}

// Connect to the database, and initial setup
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
	if _, err := db.CreateIndex(writePolicy, "test", BIDS, "idx:br:1", "offer_id", as.NUMERIC); err != nil {
		fmt.Printf("Create Index Failed: %s\n", err.Error())
	}

	// index on broker_id of bids, so we can find bids by a particular broker
	if _, err := db.CreateIndex(writePolicy, "test", BIDS, "idx:br:2", "broker_id", as.NUMERIC); err != nil {
		fmt.Printf("Create Index Failed: %s\n", err.Error())
	}

	// index on broker_id of offers, so we can find offers by a particular broker
	if _, err := db.CreateIndex(writePolicy, "test", OFFERS, "idx:br:3", "broker_id", as.NUMERIC); err != nil {
		fmt.Printf("Create Index Failed: %s\n", err.Error())
	}

	// index on ticker of prices
	if _, err := db.CreateIndex(writePolicy, "test", PRICES, "idx:br:4", "ticker", as.STRING); err != nil {
		fmt.Printf("Create Index Failed: %s\n", err.Error())
	}

}
