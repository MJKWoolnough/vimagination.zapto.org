package main

import (
	"github.com/MJKWoolnough/gopherjs/rpc"

	"honnef.co/go/js/dom"

	"github.com/gopherjs/gopherjs/js"
)

var RPC jRPC

type jRPC struct {
	rpc *rpc.Client
}

func InitRPC() error {
	conn, err := rpc.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/FH/rpc")
	if err != nil {
		return err
	}
	dom.GetWindow().AddEventListener("beforeunload", false, func(dom.Event) {
		conn.Close()
	})
	RPC.rpc = conn
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
