package main

import (
	m "github.com/aerospike-labs/stock-exchange/models"
	"sync"
)

// stores
var auctionMap *AuctionMap = NewAuctionMap()

type AuctionMap struct {
	auctionMap map[int]chan *m.Bid
	rwmutex    sync.RWMutex
}

func NewAuctionMap() *AuctionMap {
	return &AuctionMap{
		auctionMap: make(map[int]chan *m.Bid),
	}
}

// Add creates and adds an auction with a bid channel to receive the bids
func (am *AuctionMap) Add(auctionId int) chan *m.Bid {
	am.rwmutex.Lock()
	defer am.rwmutex.Unlock()

	ch := make(chan *m.Bid, 64)
	am.auctionMap[auctionId] = ch
	return ch
}

// Get finds a bid channel in the map and returns it
func (am *AuctionMap) Get(auctionId int) chan *m.Bid {
	am.rwmutex.RLock()
	defer am.rwmutex.RUnlock()

	return am.auctionMap[auctionId]
}

// Remove removes the auction and its bid channel from the map
func (am *AuctionMap) Remove(auctionId int) {
	am.rwmutex.Lock()
	defer am.rwmutex.Unlock()

	delete(am.auctionMap, auctionId)
}
