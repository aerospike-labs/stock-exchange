package main

import (
	"errors"
	"fmt"
	m "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"
	"net/http"
	"sync/atomic"
)

// sequence for Bid Ids
var bidIdSeq int64 = 0

// sequence for Offer Ids
var offerIdSeq int64 = 0

type Command struct{}

// name of database entities
const (
	NAMESPACE = "test"
	STOCKS    = "stocks"
	OFFERS    = "offers"
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

	return nil
}

func (command *Command) Offers(r *http.Request, args *Command, reply *[]m.Offer) error {

	recordset, err := db.ScanAll(scanPolicy, NAMESPACE, OFFERS)
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

	return nil
}

func (command *Command) Offer(r *http.Request, offer *m.Offer, offerId *int) error {

	var err error

	*offerId = 0

	// Check if the broker has enough inventory
	brokerKeyId := fmt.Sprintf("%s:%d", BROKERS, offer.BrokerId)
	brokerKey, err := as.NewKey(NAMESPACE, BROKERS, brokerKeyId)
	if err != nil {
		return err
	}

	brokerRec, err := db.Get(readPolicy, brokerKey, offer.Ticker, offer.Ticker+"_os")
	if err != nil {
		return err
	}

	if brokerRec == nil {
		return fmt.Errorf("Broker not found %d", offer.BrokerId)
	}

	// fmt.Printf("%#v\n\n", brokerRec)

	if brokerRec.Bins == nil || len(brokerRec.Bins) == 0 {
		return errors.New("Broker does not have any inventory of the stock")
	} else if _, exists := brokerRec.Bins[offer.Ticker+"_os"]; !exists {
		// set the outstanding as much as the inventory - bin quantity
		err := db.Put(writePolicy, brokerKey, as.BinMap{offer.Ticker + "_os": int(int(brokerRec.Bins[offer.Ticker].(int)) - offer.Quantity)})
		if err != nil {
			return err
		}
	} else {
		opRec, err := db.Operate(
			writePolicy,
			brokerKey,
			as.AddOp(as.NewBin(offer.Ticker+"_os", -1*int(offer.Quantity))),
			as.GetOp(),
		)
		if err != nil {
			return err
		}

		if inventory, exists := opRec.Bins[offer.Ticker+"_os"]; !exists || int(inventory.(int)) < offer.Quantity {
			db.Add(writePolicy, brokerKey, as.BinMap{offer.Ticker + "_os": offer.Quantity})
			return errors.New("Not enough inventory")
		}
	}

	offer.Id = int(atomic.AddInt64(&offerIdSeq, 1))

	// put the offer up
	offerKeyId := fmt.Sprintf("%s:%d", OFFERS, offer.Id)
	offerKey, err := as.NewKey(NAMESPACE, OFFERS, offerKeyId)
	if err != nil {
		return err
	}

	offerBins := as.BinMap{
		"offer_id":  offer.Id,
		"broker_id": offer.BrokerId,
		"ticker":    offer.Ticker,
		"quantity":  offer.Quantity,
		"price":     offer.Price,
		"ttl":       offer.TTL,
	}

	if err = db.Put(writePolicy, offerKey, offerBins); err != nil {
		return err
	}

	*offerId = offer.Id

	broadcast <- &m.Notification{
		Version: "2.0",
		Method:  "Offer",
		Params:  offer,
	}

	go Auctioner(offer.Id, offer.TTL)
	return nil
}

func (command *Command) Bid(r *http.Request, bid *m.Bid, bidId *int) error {

	var err error

	offerKeyId := fmt.Sprintf("%s:%d", OFFERS, bid.OfferId)
	offerKey, err := as.NewKey(NAMESPACE, OFFERS, offerKeyId)
	if err != nil {
		return err
	}

	exists, err := db.Exists(readPolicy, offerKey)
	bidChan := auctionMap.Get(int(offerIdSeq))
	if err != nil || !exists || bidChan == nil {
		errors.New("Auction has finished.")
	}

	bid.Id = int(atomic.AddInt64(&bidIdSeq, 1))

	bidKeyId := fmt.Sprintf("%s:%d", BIDS, bid.Id)
	bidKey, err := as.NewKey(NAMESPACE, BIDS, bidKeyId)
	if err != nil {
		return err
	}

	bidBins := as.BinMap{
		"bid_id":     bid.Id,
		"auction_id": bid.OfferId,
		"broker_id":  bid.BrokerId,
		"price":      bid.Price,
	}

	if err := db.Put(writePolicy, bidKey, bidBins); err != nil {
		return err
	}

	*bidId = bid.Id

	bidChan <- bid

	broadcast <- &m.Notification{
		Version: "2.0",
		Method:  "Bid",
		Params:  bid,
	}

	return nil
}
