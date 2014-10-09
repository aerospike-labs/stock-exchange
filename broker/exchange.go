package main

import (
	. "github.com/aerospike-labs/stock-exchange/models"
	"sync/atomic"
)

type ExchangeClient struct {
	BrokerId   uint64
	client     *WebSocketClient
	OfferSeq   uint64
	RequestSeq uint64
}

func NewExchangeClient(borkerId uint64, host string, port uint16) (*ExchangeClient, error) {

	client, err := NewWebSocketClient(host, port)
	if err != nil {
		return nil, err
	}

	exchange := &ExchangeClient{
		BrokerId:   borkerId,
		client:     client,
		OfferSeq:   0,
		RequestSeq: 0,
	}

	return exchange, nil
}

func (exchange *ExchangeClient) Listen() {
	go exchange.client.listen()
}

func (exchange *ExchangeClient) Close() {
	exchange.client.close()
}

func (exchange *ExchangeClient) Messages() chan interface{} {
	return exchange.client.messages
}

func (exchange *ExchangeClient) Done() chan bool {
	return exchange.client.done
}

// Issue a sell offer
func (exchange *ExchangeClient) Sell(ticker string, quantity uint64, price uint64, ttl uint32) (uint64, error) {

	reqId := atomic.AddUint64(&exchange.RequestSeq, 1)
	offerId := atomic.AddUint64(&exchange.OfferSeq, 1)

	sell := &Offer{
		BrokerId:  exchange.BrokerId,
		OfferId:   offerId,
		OfferType: "sell",
		TTL:       ttl,
		Ticker:    ticker,
		Quantity:  quantity,
		Price:     price,
	}

	req := &Request{
		Method: "Command.Sell",
		Params: []interface{}{sell},
		Id:     []uint64{reqId, T_OFFER},
	}

	return reqId, exchange.client.send(req)
}

// Issue a buy offer
func (exchange *ExchangeClient) Buy(ticker string, quantity uint64, price uint64, ttl uint32) (uint64, error) {

	reqId := atomic.AddUint64(&exchange.RequestSeq, 1)
	offerId := atomic.AddUint64(&exchange.OfferSeq, 1)

	buy := &Offer{
		BrokerId:  exchange.BrokerId,
		OfferId:   offerId,
		OfferType: "buy",
		TTL:       ttl,
		Ticker:    ticker,
		Quantity:  quantity,
		Price:     price,
	}

	req := &Request{
		Method: "Command.Buy",
		Params: []interface{}{buy},
		Id:     []uint64{reqId, T_OFFER},
	}

	return reqId, exchange.client.send(req)
}

// Cancel an offer
func (exchange *ExchangeClient) Cancel(offerId uint64) (uint64, error) {

	reqId := atomic.AddUint64(&exchange.RequestSeq, 1)

	cancel := &BrokerOffer{
		BrokerId: exchange.BrokerId,
		OfferId:  offerId,
	}

	req := &Request{
		Method: "Command.Cancel",
		Params: []interface{}{cancel},
		Id:     []uint64{reqId, T_BROKEROFFER},
	}

	return reqId, exchange.client.send(req)
}

// List the outstanding offers
func (exchange *ExchangeClient) Offers() (uint64, error) {

	reqId := atomic.AddUint64(&exchange.RequestSeq, 1)

	req := &Request{
		Method: "Command.Offers",
		Params: []interface{}{},
		Id:     []uint64{reqId, T_OFFERLIST},
	}

	return reqId, exchange.client.send(req)
}

// List the current stock prices
func (exchange *ExchangeClient) Stocks() (uint64, error) {

	reqId := atomic.AddUint64(&exchange.RequestSeq, 1)

	req := &Request{
		Method: "Command.Stocks",
		Params: []interface{}{},
		Id:     []uint64{reqId, T_STOCKLIST},
	}

	return reqId, exchange.client.send(req)
}
