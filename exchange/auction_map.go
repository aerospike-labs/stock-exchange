package main

import (
	m "github.com/aerospike-labs/stock-exchange/models"
	"sync"
)

type AuctionMap struct {
	auctionMap map[int64]chan *m.Offer
	rwmutex    sync.RWMutex
}

func NewAuctionMap() *AuctionMap {
	return &AuctionMap{
		auctionMap: map[int64]chan *m.Offer{},
	}
}

func (am *AuctionMap) Add(auctionId int64) chan *m.Offer {
	am.rwmutex.Lock()
	defer am.rwmutex.Unlock()

	ch := make(chan *m.Offer, 64)
	am.auctionMap[auctionId] = ch
	return ch
}

func (am *AuctionMap) Get(auctionId int64) chan *m.Offer {
	am.rwmutex.RLock()
	defer am.rwmutex.Unlock()

	return am.auctionMap[auctionId]
}

func (am *AuctionMap) Remove(auctionId int64) {
	am.rwmutex.Lock()
	defer am.rwmutex.Unlock()

	delete(am.auctionMap, auctionId)
}
