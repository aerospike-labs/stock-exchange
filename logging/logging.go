package logging

import (
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
				log.Printf("<REQ> %d %s %#v", m.Id, m.Method, m.Params)
			case *Response:
				log.Printf("<RES> %d %#v %#v", m.Id, m.Result, m.Error)
			case *RawResponse:
				log.Printf("<RES> %d %#v %#v", m.Id, m.Result, m.Error)
			case *Notification:
				log.Printf("<NOT> %s %#v", m.Method, m.Params)
			case *RawNotification:
				log.Printf("<NOT> %s %#v", m.Method, m.Params)
			}
		}
	}
}

func Close() {
	close(Channel)
}
