package main

import (
	"fmt"
	m "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"
	"time"
)

// This method is called as go-routine, and will
func Auctioner(offerId int, TTL int) {

	bidderChan := auctionMap.Add(offerId)
	var bestBid *m.Bid

L:
	for {
		select {
		case <-time.After(time.Duration(TTL) * time.Second):

			if bestBid != nil {
				CloseAuction(offerId, bestBid)
				broadcast <- &m.Notification{
					Version: "2.0",
					Method:  "Close",
					Params:  bestBid,
				}
			} else {
				broadcast <- &m.Notification{
					Version: "2.0",
					Method:  "Cancel",
					Params:  offerId,
				}
			}

			break L

		// each bid is sent to this channel after evaluation
		case bid := <-bidderChan:
			if bestBid == nil || bestBid.Price < bid.Price {
				bestBid = bid
			}
		}
	}
}

func CloseAuction(offerId int, bid *m.Bid) {
	auctionMap.Remove(offerId)

	// Check if the broker has enough inventory
	offerKeyId := fmt.Sprintf("%s:%d", OFFERS, offerId)
	offerKey, _ := as.NewKey(NAMESPACE, OFFERS, offerKeyId)
	offerRec, _ := db.Get(readPolicy, offerKey)
	sellerId := offerRec.Bins["broker_id"].(int)

	// add to buyer's inventory and reduce credit
	buyerKeyId := fmt.Sprintf("%s:%d", BROKERS, bid.BrokerId)
	buyerKey, _ := as.NewKey(NAMESPACE, BROKERS, buyerKeyId)
	db.Operate(writePolicy, buyerKey,
		as.AddOp(as.NewBin(offerRec.Bins["ticker"].(string), offerRec.Bins["Quantity"].(int))),
		as.AddOp(as.NewBin("credit", -1*offerRec.Bins["Quantity"].(int)*int(bid.Price))),
	)

	// reduce seller's inventory and add to credit
	sellerKeyId := fmt.Sprintf("%s:%d", BROKERS, sellerId)
	sellerKey, _ := as.NewKey(NAMESPACE, BROKERS, sellerKeyId)
	db.Operate(writePolicy, sellerKey,
		as.AddOp(as.NewBin(offerRec.Bins["ticker"].(string), -1*offerRec.Bins["Quantity"].(int))),
		as.AddOp(as.NewBin("credit", offerRec.Bins["Quantity"].(int)*int(bid.Price))),
	)

	// mark the bid as winner
	bidKeyId := fmt.Sprintf("%s:%d", BIDS, bid.Id)
	bidKey, _ := as.NewKey(NAMESPACE, BIDS, bidKeyId)
	db.Put(writePolicy, bidKey, as.BinMap{"winner": 1})
	db.Put(writePolicy, offerKey, as.BinMap{"finished": 1, "winner": bid.Id})
}
