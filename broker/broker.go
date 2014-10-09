package main

import (
	. "github.com/aerospike-labs/stock-exchange/models"
)

// this is for the test
type Args struct {
	A int
	B int
}

type Broker struct {
	Id         uint64
	Client     Client
	OfferSeq   uint64
	RequestSeq uint64
}

// Issue a sell offer
func (b *Broker) Sell(ticker string, quantity uint64, price uint64, ttl uint32) (uint64, error) {

	reqId := b.RequestSeq // atomic increment
	offerId := b.OfferSeq // atomic increment

	sell := &Offer{
		BrokerId:  b.Id,
		OfferId:   offerId,
		OfferType: "sell",
		TTL:       ttl,
		Ticker:    ticker,
		Quantity:  quantity,
		Price:     price,
	}

	req := &Request{
		Method: "Offer.Sell",
		Params: []interface{}{sell},
		Id:     reqId,
	}

	return reqId, b.Client.Send(req)
}

// Issue a buy offer
func (b *Broker) Buy(ticker string, quantity uint64, price uint64, ttl uint32) (uint64, error) {

	reqId := b.RequestSeq // atomic increment
	offerId := b.OfferSeq // atomic increment

	buy := &Offer{
		BrokerId:  b.Id,
		OfferId:   offerId,
		OfferType: "buy",
		TTL:       ttl,
		Ticker:    ticker,
		Quantity:  quantity,
		Price:     price,
	}

	req := &Request{
		Method: "Offer.Buy",
		Params: []interface{}{buy},
		Id:     reqId,
	}

	return reqId, b.Client.Send(req)
}

// Cancel an offer
func (b *Broker) Cancel(offerId uint64) (uint64, error) {

	reqId := b.RequestSeq // atomic increment

	cancel := &BrokerOffer{
		BrokerId: b.Id,
		OfferId:  offerId,
	}

	req := &Request{
		Method: "Offer.Cancel",
		Params: []interface{}{cancel},
		Id:     reqId,
	}

	return reqId, b.Client.Send(req)
}

// List the outstanding offers
func (b *Broker) Offers() (uint64, error) {

	reqId := b.RequestSeq // atomic increment

	req := &Request{
		Method: "Offer.List",
		Params: []interface{}{},
		Id:     reqId,
	}

	return reqId, b.Client.Send(req)
}

// List the current stock prices
func (b *Broker) Stocks() (uint64, error) {

	reqId := b.RequestSeq // atomic increment

	req := &Request{
		Method: "Stock.List",
		Params: []interface{}{},
		Id:     reqId,
	}

	return reqId, b.Client.Send(req)
}
