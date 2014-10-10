package models

import (
	"encoding/json"
)

type Stock struct {
	Ticker   string
	Quantity int
	Price    int
}

type StockList []Stock

type Offer struct {
	Id       int
	BrokerId int
	TTL      int
	Ticker   string
	Quantity int
	Price    int
}

type OfferList []Offer

type Bid struct {
	Id       int
	BrokerId int
	OfferId  int
	Price    int
}

type BidList []Bid

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
