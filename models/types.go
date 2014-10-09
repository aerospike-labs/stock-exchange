package models

import (
	"encoding/json"
)

type OfferType uint8

const (
	BUY OfferType = iota
	SELL
)

const (
	T_OFFER = iota
	T_BROKEROFFER
	T_STOCKLIST
	T_OFFERLIST
	T_TRASNACTION
	T_STOCK
)

type Stock struct {
	Ticker   string
	Quantity uint64
	Price    uint64
}

type StockList []Stock

type Offer struct {
	BrokerId  uint64
	OfferId   uint64
	OfferType OfferType
	TTL       uint32
	Ticker    string
	Quantity  uint64
	Price     uint64
}

type OfferList []Offer

type BrokerOffer struct {
	BrokerId uint64
	OfferId  uint64
}

type Transaction struct {
	Buyer    BrokerOffer
	Seller   BrokerOffer
	Ticker   string
	Quantity uint64
	Price    uint64
}

type Request struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	Id     []uint64      `json:"id"`
}

type Response struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	Id     []uint64    `json:"id"`
}

type Notification struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type Message struct {
	Method string          `json:"method"`
	Params []interface{}   `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
	Id     []uint64        `json:"id"`
}

func (m *Message) IsRequest() bool {
	return len(m.Method) != 0 && m.Id != nil && len(m.Id) == 2
}

func (m *Message) IsNotification() bool {
	return len(m.Method) != 0 && m.Id == nil
}

func (m *Message) IsResponse() bool {
	return len(m.Method) == 0
}
