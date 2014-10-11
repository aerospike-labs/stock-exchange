package models

import (
	"encoding/json"
)

// Broker is a person who can offer stock for sale or place bids on offers
type Broker struct {
	Id     int
	Name   string
	Credit int
}

type Stock struct {
	Ticker   string
	Quantity int
	Price    int
}

type StockList []Stock

// A parcel of stock to be sold
// This probably should be called parcel, but I don't
// want to change all the code now.
type Offer struct {
	Id       int
	BrokerId int
	TTL      int
	Ticker   string
	Quantity int
	Price    int
}

type OfferList []Offer

// An offer to buy a parcel of stock.
type Bid struct {
	Id       int
	BrokerId int
	OfferId  int
	Price    int
}

type BidList []Bid

// Request to be used for composing
type Request struct {
	Version string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

// Raw allows us to partially unmarshal the request
type RawRequest struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      int             `json:"id"`
}

// Response to be used for composing
type Response struct {
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
	Id      int         `json:"id"`
}

// Raw allows us to partially unmarshal the response
type RawResponse struct {
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   json.RawMessage `json:"error"`
	Id      int             `json:"id"`
}

// Notification to be used for composing
type Notification struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// Raw allows us to partially unmarshal the notification
type RawNotification struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}
