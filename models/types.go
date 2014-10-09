package models

type Stock struct {
	Ticker   string
	Quantity uint64
	Price    uint64
}

type StockList []Stock

type Offer struct {
	BrokerId  uint64
	OfferId   uint64
	OfferType string
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
	Id     uint64        `json:"id"`
}

type Response struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	Id     uint64      `json:"id"`
}

type Notification struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type Message struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	Result interface{}   `json:"result"`
	Error  interface{}   `json:"error"`
	Id     uint64        `json:"id"`
}

func (m *Message) IsRequest() bool {
	return len(m.Method) != 0 && m.Id != 0
}

func (m *Message) IsNotification() bool {
	return len(m.Method) != 0 && m.Id == 0
}

func (m *Message) IsResponse() bool {
	return len(m.Method) == 0
}
