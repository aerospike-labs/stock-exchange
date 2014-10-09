package main

import (
	"code.google.com/p/go.net/websocket"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Args struct {
	A int
	B int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	println(args.A, "+", args.B, "=", *reply)
	return nil
}

func main() {
	rpc.Register(new(Arith))

	http.Handle("/conn", websocket.Handler(serve))
	http.ListenAndServe("localhost:7000", nil)
}

func serve(ws *websocket.Conn) {
	jsonrpc.ServeConn(ws)
}
