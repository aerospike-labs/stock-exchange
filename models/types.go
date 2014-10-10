package models

import (
	"encoding/json"
)

const (
	T_OFFER = iota
	T_BROKEROFFER
	T_STOCKLIST
	T_OFFERLIST
	T_TRANSACTION
	T_STOCK
	T_ID
)

type Stock struct {
	Ticker   string
	Quantity int
	Price    int
}

type StockList []Stock

type Offer struct {
	OfferId  int
	BrokerId int
	TTL      int
	Ticker   string
	Quantity int
	Price    int
}

type OfferList []Offer

type BrokerOffer struct {
	BrokerId int
	OfferId  int
}

type Transaction struct {
	Buyer    BrokerOffer
	Seller   BrokerOffer
	Ticker   string
	Quantity int
	Price    int
}

type BroadCastType int

const (
	AUCTION_BEGIN BroadCastType = iota
	AUCTION_ENDED
	BID_HAPPENNED
)

type Broadcast struct {
	Type    BroadCastType
	Auction Offer
	Bid     Offer
}

type Request struct {
	Version string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type RawRequest struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      int             `json:"id"`
}

type Response struct {
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
	Id      int         `json:"id"`
}
type RawResponse struct {
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   json.RawMessage `json:"error"`
	Id      int             `json:"id"`
}

type Notification struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type RawNotification struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}
