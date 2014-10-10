package main

import (
	. "github.com/aerospike-labs/stock-exchange/models"

	"github.com/aerospike-labs/stock-exchange/logging"

	"code.google.com/p/go.net/websocket"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var requestSeq int64 = 0

type ExchangeClient struct {
	BrokerId int
	host     string
	port     uint16
	rpc      *http.Client
	rpcUrl   string
	ws       *websocket.Conn
	wsUrl    string
	Messages chan *RawNotification
	Done     chan bool
}

func NewExchangeClient(borkerId int, host string, port uint16) (*ExchangeClient, error) {

	wsUrl := fmt.Sprintf("ws://%s:%d/ws", host, port)
	rpcUrl := fmt.Sprintf("http://%s:%d/rpc", host, port)

	ws, err := websocket.Dial(wsUrl, "", rpcUrl)
	if err != nil {
		return nil, err
	}

	ex := &ExchangeClient{
		BrokerId: borkerId,
		host:     host,
		port:     port,
		rpc:      &http.Client{},
		rpcUrl:   rpcUrl,
		ws:       ws,
		wsUrl:    wsUrl,
		Messages: make(chan *RawNotification, 1024),
		Done:     make(chan bool),
	}

	return ex, nil
}

// Listen for messages
func (ex *ExchangeClient) Listen() {
	for {
		var raw []byte

		if err := websocket.Message.Receive(ex.ws, &raw); err != nil {
			fmt.Printf("GOT ERROR %#v\n\n", err)
			ex.Done <- true
			return
		}

		var notice RawNotification

		json.Unmarshal([]byte(raw), &notice)

		logging.Log(&notice)
		ex.Messages <- &notice
	}
}

// Close the connection
func (ex *ExchangeClient) Close() {
	ex.ws.Close()
	close(ex.Messages)
	close(ex.Done)
}

// Offer a stock to the ex
// Returns the OfferId for the offer.
func (ex *ExchangeClient) CreateAuction(ticker string, quantity int, price int, ttl uint32) (bool, error) {

	reqId := 1

	offer := &Offer{
		BrokerId: ex.BrokerId,
		OfferId:  0,
		TTL:      ttl,
		Ticker:   ticker,
		Quantity: quantity,
		Price:    price,
	}

	res, err := ex.call("Command.CreateAuction", offer, int(reqId))
	if err != nil {
		return false, err
	}

	result := true
	json.Unmarshal(res, &result)
	return result, nil
}

// Issue a buy offer
// Returns the BidId for the big
func (ex *ExchangeClient) Bid(auctionId int, price int) (bool, error) {

	reqId := 1

	bid := &Offer{
		BrokerId: ex.BrokerId,
		OfferId:  auctionId,
		TTL:      0,
		Ticker:   "",
		Quantity: 0,
		Price:    price,
	}

	res, err := ex.call("Command.Bid", bid, int(reqId))
	if err != nil {
		return false, err
	}

	result := true
	json.Unmarshal(res, &result)
	return result, nil
}

// Issue a buy offer
// Returns the BidId for the big
func (ex *ExchangeClient) Stocks() (StockList, error) {

	reqId := 1

	// var bid interface{} = nil

	res, err := ex.call("Command.Stocks", nil, int(reqId))
	if err != nil {
		return nil, err
	}

	result := StockList{}
	json.Unmarshal(res, &result)
	return result, nil
}

// Issue a buy offer
// Returns the BidId for the big
func (ex *ExchangeClient) Auctions() (OfferList, error) {

	reqId := 1

	// var bid interface{} = nil

	res, err := ex.call("Command.Auctions", nil, int(reqId))
	if err != nil {
		return nil, err
	}

	result := OfferList{}
	json.Unmarshal(res, &result)
	return result, nil
}

// List the current stock prices
func (ex *ExchangeClient) call(method string, params interface{}, id int) (json.RawMessage, error) {

	req := Request{
		Version: "2.0",
		Method:  method,
		Params:  []interface{}{params},
		Id:      id,
	}

	res := RawResponse{}
	if err := ex.send(&req, &res); err != nil {
		return nil, err
	}

	var reserr interface{}
	json.Unmarshal(res.Error, &reserr)

	if reserr != nil {
		return nil, fmt.Errorf("Command failed: %#v", reserr)
	}

	return res.Result, nil
}

// List the current stock prices
func (ex *ExchangeClient) send(req *Request, res *RawResponse) error {

	logging.Log(req)

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	hreq, err := http.NewRequest("POST", ex.rpcUrl, bytes.NewReader(body))
	if err != nil {
		return err
	}

	hreq.Header.Add("Content-Type", "application/json")
	hreq.Header.Add("Content-Length", string(len(body)))

	hres, err := ex.rpc.Do(hreq)
	if err != nil {
		return err
	}

	hbody, err := ioutil.ReadAll(hres.Body)
	if err != nil {
		return err
	}

	json.Unmarshal(hbody, res)

	logging.Log(res)

	return nil
}
