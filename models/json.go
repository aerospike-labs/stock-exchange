package models

import (
// "encoding/json"
)

func Unmarshal(raw []byte) *Response {

	// println("RAW", string(raw))

	// var message Message
	// json.Unmarshal([]byte(raw), &message)

	// // if message.IsResponse() {

	// res := &Response{
	// 	Result: nil,
	// 	Error:  nil,
	// 	Id:     message.Id,
	// }

	// switch message.Id[1] {
	// case T_STOCKLIST:
	// 	var stocks StockList
	// 	json.Unmarshal(message.Result, &stocks)
	// 	res.Result = stocks
	// case T_OFFERLIST:
	// 	var offers OfferList
	// 	json.Unmarshal(message.Result, &offers)
	// 	res.Result = offers
	// case T_OFFER:
	// 	var offer Offer
	// 	json.Unmarshal(message.Result, &offer)
	// 	res.Result = offer
	// case T_ID:
	// 	res.Result = true
	// }

	// return res
	// } else if message.IsNotification() {
	// 	return &Notification{
	// 		Method: message.Method,
	// 		Params: message.Params,
	// 	}
	// } else {
	// 	return nil
	// }
	return nil
}
