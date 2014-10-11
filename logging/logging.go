package logging

import (
	"encoding/json"
	. "github.com/aerospike-labs/stock-exchange/models"
	"log"
)

var (
	Channel chan interface{} = make(chan interface{}, 1024)
	Enabled bool             = false
)

func Log(msg interface{}) {
	if Enabled {
		Channel <- msg
	}
}

func Listen() {
	if !Enabled {
		return
	}
	for {
		select {
		case msg := <-Channel:
			switch m := msg.(type) {
			case string:
				log.Println(m)
			case *Request:
				log.Printf("<REQ> %d %s %#v", m.Id, m.Method, m.Params)
			case *RawRequest:
				var params interface{}
				json.Unmarshal(m.Params, &params)
				log.Printf("<REQ> %d %s %#v", m.Id, m.Method, params)
			case *Response:
				log.Printf("<RES> %d %#v %#v", m.Id, m.Result, m.Error)
			case *RawResponse:
				var err interface{}
				var res interface{}

				json.Unmarshal(m.Error, &err)
				json.Unmarshal(m.Result, &res)

				log.Printf("<RES> %d %#v %#v", m.Id, res, err)
			case RawResponse:
				var err interface{}
				var res interface{}

				json.Unmarshal(m.Error, &err)
				json.Unmarshal(m.Result, &res)

				log.Printf("<RES> %d %#v %#v", m.Id, res, err)
			case *Notification:
				log.Printf("<MSG> %s %#v", m.Method, m.Params)
			case *RawNotification:
				var params interface{}
				json.Unmarshal(m.Params, &params)
				log.Printf("<MSG> %s %#v", m.Method, params)
			}
		}
	}
}

func Close() {
	close(Channel)
}
