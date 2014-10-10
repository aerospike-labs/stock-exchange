package main

import (
	m "github.com/aerospike-labs/stock-exchange/models"
	as "github.com/aerospike/aerospike-client-go"
	"time"
)

// This method is called as go-routine, and will
func Auctioner(auctionId int, TTL int) {
	// broadcast auction
	broadcast <- &m.Broadcast{
		Type:    m.AUCTION_BEGIN,
		Auction: *findAuction(int(auctionId)),
	}

	bidderChan := auctionMap.Add(int64(auctionId))
	var bestBid *m.Offer

L:
	for {
		select {
		case <-time.After(time.Duration(TTL) * time.Second):
			b := &m.Broadcast{
				Type:    m.AUCTION_ENDED,
				Auction: *findAuction(auctionId),
			}

			if bestBid != nil {
				b.Bid = *findBid(bestBid.OfferId)
			}
			broadcast <- b

			if bestBid != nil {
				CloseAuction(auctionId, bestBid)
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
