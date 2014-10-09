package models

import (
	"encoding/json"
)

func Unmarshal(raw []byte) interface{} {

	stocks := StockList{}

	res := &Response{
		Result: stocks,
		Error:  nil,
		Id:     message.Id,
	}

	return res

	// var message Message
	// json.Unmarshal([]byte(raw), &message)

	// if message.IsResponse() {

	// 	res := &Response{
	// 		Result: nil,
	// 		Error:  nil,
	// 		Id:     message.Id,
	// 	}

	// 	switch message.Id[1] {
	// 	case T_STOCKLIST:
	// 		var stocks StockList
	// 		json.Unmarshal(message.Result, &stocks)
	// 		res.Result = stocks
	// 	case T_OFFERLIST:
	// 		var offers OfferList
	// 		json.Unmarshal(message.Result, &offers)
	// 		res.Result = offers
	// 	case T_OFFER:
	// 		var offer Offer
	// 		json.Unmarshal(message.Result, &offer)
	// 		res.Result = offer
	// 	}

	// 	return res
	// } else if message.IsNotification() {
	// 	return &Notification{
	// 		Method: message.Method,
	// 		Params: message.Params,
	// 	}
	// } else {
	// 	return nil
	// }
}
