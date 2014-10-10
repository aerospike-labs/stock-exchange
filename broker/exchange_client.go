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
	"sync/atomic"
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
	Messages chan *Notification
	Done     chan bool
}

func NewExchangeClient(borkerId int, host string, port uint16) (*ExchangeClient, error) {

	wsUrl := fmt.Sprintf("ws://%s:%d/ws", host, port)
	rpcUrl := fmt.Sprintf("http://%s:%d/rpc", host, port)

	ex := &ExchangeClient{
		BrokerId: borkerId,
		host:     host,
		port:     port,
		rpc:      &http.Client{},
		rpcUrl:   rpcUrl,
		ws:       nil,
		wsUrl:    wsUrl,
		Messages: make(chan *Notification, 1024),
		Done:     make(chan bool),
	}

	return ex, nil
}

// Listen for messages
func (ex *ExchangeClient) Listen() error {

	ws, err := websocket.Dial(ex.wsUrl, "", ex.rpcUrl)
	if err != nil {
		return err
	}

	ex.ws = ws

	for {
		var raw []byte

		if err := websocket.Message.Receive(ex.ws, &raw); err != nil {
			fmt.Printf("GOT ERROR %#v\n\n", err)
			ex.Done <- true
			return err
		}

		var rawNotice RawNotification

		json.Unmarshal([]byte(raw), &rawNotice)

		notice := &Notification{
			Version: rawNotice.Version,
			Method:  rawNotice.Method,
			Params:  nil,
		}

		switch rawNotice.Method {
		case "Offer":
			notice.Params = Offer{}
			json.Unmarshal(rawNotice.Params, &notice.Params)

		case "Bid":
			notice.Params = Bid{}
			json.Unmarshal(rawNotice.Params, &notice.Params)

		case "Close":
			notice.Params = Bid{}
			json.Unmarshal(rawNotice.Params, &notice.Params)

		}

		logging.Log(&notice)
		ex.Messages <- notice
	}

	return nil
}

// Close the connection
func (ex *ExchangeClient) Close() {
	if ex.ws != nil {
		ex.ws.Close()
	}
	close(ex.Messages)
	close(ex.Done)
}

// Offer a stock to the ex
// Returns the OfferId for the offer.
func (ex *ExchangeClient) Offer(ticker string, quantity int, price int, ttl int) (int, error) {

	offer := &Offer{
		Id:       0, // Set to 0, b/c it will be assigned by exchange
		BrokerId: ex.BrokerId,
		TTL:      ttl,
		Ticker:   ticker,
		Quantity: quantity,
		Price:    price,
	}

	res, err := ex.call("Command.Offer", offer)
	if err != nil {
		return 0, err
	}

	var result int = 0
	json.Unmarshal(res, &result)
	return result, nil
}

// Issue a buy offer
// Returns the BidId for the big
func (ex *ExchangeClient) Bid(offerId int, price int) (int, error) {

	bid := &Bid{
		Id:       0, // Set to 0, b/c it will be assigned by exchange
		BrokerId: ex.BrokerId,
		OfferId:  offerId,
		Price:    price,
	}

	res, err := ex.call("Command.Bid", bid)
	if err != nil {
		return 0, err
	}

	var result int = 0
	json.Unmarshal(res, &result)
	return result, nil
}

// Issue a buy offer
// Returns the BidId for the big
func (ex *ExchangeClient) Stocks() (StockList, error) {

	res, err := ex.call("Command.Stocks", nil)
	if err != nil {
		return nil, err
	}

	result := StockList{}
	json.Unmarshal(res, &result)
	return result, nil
}

// Issue a buy offer
// Returns the BidId for the big
func (ex *ExchangeClient) Offers() (OfferList, error) {

	res, err := ex.call("Command.Offers", nil)
	if err != nil {
		return nil, err
	}

	result := OfferList{}
	json.Unmarshal(res, &result)
	return result, nil
}

// List the current stock prices
func (ex *ExchangeClient) call(method string, params interface{}) (json.RawMessage, error) {

	req := Request{
		Version: "2.0",
		Method:  method,
		Params:  []interface{}{params},
		Id:      int(atomic.AddInt64(&requestSeq, 1)),
	}

	res := RawResponse{}
	if err := ex.send(&req, &res); err != nil {
		return nil, err
	}

	var reserr interface{}
	json.Unmarshal(res.Error, &reserr)

	if reserr != nil {
		logging.Log(reserr)

		return nil, fmt.Errorf("Command failed: %#v", reserr)
	}
	logging.Log(res.Result)

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

	return nil
}
