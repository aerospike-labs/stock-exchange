package main

import (
	. "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"

	"fmt"
	"time"
)

const (
	NAMESPACE      = "test"
	SET_BIDS       = "bids"
	SET_OFFERS     = "offers"
	SET_TICKERS    = "tickers"
	SET_PRICES     = "prices"
	SET_PORTFOLIOS = "portfolios"
)

var db *as.Client
var readPolicy *as.BasePolicy
var writePolicy *as.WritePolicy
var scanPolicy *as.ScanPolicy

func storeOffer(offer *Offer) error {

	var err error

	key, err := as.NewKey(NAMESPACE, SET_OFFERS, offer.Id)
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

	if err = db.Put(writePolicy, key, rec); err != nil {
		return nil
	}

	return nil
}

func storeBid(bid *Bid) error {

	var err error

	key, err := as.NewKey(NAMESPACE, SET_BIDS, bid.Id)
	if err != nil {
		return err
	}

	rec := as.BinMap{
		"bid_id":    bid.Id,
		"broker_id": bid.BrokerId,
		"offer_id":  bid.OfferId,
		"price":     bid.Price,
	}

	if err = db.Put(writePolicy, key, rec); err != nil {
		return nil
	}

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

	offerId := fmt.Sprintf("%s:%d", SET_OFFERS, bid.OfferId)
	offerKey, err := as.NewKey(NAMESPACE, SET_OFFERS, offerId)
	if err != nil {
		return nil
	}

	offerRec, err := db.Get(readPolicy, offerKey)
	if err != nil {
		return nil
	}

	ticker := offerRec.Bins["ticker"]
	quantity := offerRec.Bins["quantity"]

	// Update the current ticker price

	tickerId := fmt.Sprintf("%s:%d", SET_TICKERS, ticker)
	tickerKey, err := as.NewKey(NAMESPACE, SET_TICKERS, tickerId)
	if err != nil {
		return err
	}

	tickerBins := as.BinMap{
		"ticker": ticker,
		"price":  bid.Price,
		"time":   ts,
	}

	if err = db.Put(writePolicy, tickerKey, tickerBins); err != nil {
		return err
	}

	// Store the ticker price for historical prices
	// There is an index on ticker

	priceId := fmt.Sprintf("%s:%d:%d", SET_PRICES, ticker, ts)
	priceKey, err := as.NewKey(NAMESPACE, SET_PRICES, priceId)
	if err != nil {
		return err
	}

	priceBins := as.BinMap{
		"ticker": ticker,
		"price":  bid.Price,
		"time":   ts,
	}

	if err = db.Put(writePolicy, priceKey, priceBins); err != nil {
		return err
	}

	// Update Porfolio

	portfolioId := fmt.Sprintf("%s:%d", SET_PORTFOLIOS, bid.BrokerId)
	portfolioKey, err := as.NewKey(NAMESPACE, SET_PORTFOLIOS, portfolioId)
	if err != nil {
		return err
	}

	portfolioBins := as.BinMap{
		ticker.(string): quantity,
	}

	if err = db.Add(writePolicy, portfolioKey, portfolioBins); err != nil {
		return err
	}

	return nil
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
