package main

import (
	"net/rpc"
	"net/rpc/jsonrpc"

	"honnef.co/go/js/dom"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
)

var RPC jRPC

type jRPC struct {
	rpc *rpc.Client
}

func InitRPC() error {
	conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/FH/rpc")
	if err != nil {
		return err
	}
	dom.GetWindow().AddEventListener("beforeunload", false, func(dom.Event) {
		conn.Close()
	})
	RPC.rpc = jsonrpc.NewClient(conn)
	return nil
}

func (j jRPC) GetPerson(id uint) Person {
	var p Person
	if err := j.rpc.Call("RPC.GetPerson", id, &p); err != nil {
		panic(err)
	}
	return p
}

func (j jRPC) GetFamily(id uint) Family {
	var f Family
	if err := j.rpc.Call("RPC.GetFamily", id, &f); err != nil {
		panic(err)
	}
	return f
}
