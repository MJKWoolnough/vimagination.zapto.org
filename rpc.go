package main

import (
	"net/http"
	"net/rpc/jsonrpc"

	"golang.org/x/net/websocket"
)

type rpcSwitch struct {
	Websocket, Post http.Handler
}

func (rpc *rpcSwitch) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		rpc.Post.ServeHTTP(w, r)
		return
	}
	rpc.Websocket.ServeHTTP(w, r)
}

func rpcWebsocketHandler(conn *websocket.Conn) {
	jsonrpc.ServeConn(conn)
}

type RPC struct{}

type RPCPerson struct {
	ID                 uint
	FirstName, Surname string
	DOB, DOD           string
	Gender             byte
	ChildOf            uint
	SpouseOf           []uint
}

func (RPC) GetPerson(id uint, person *RPCPerson) error {
	p, ok := GedcomData.People[id]
	if !ok {
		p = GedcomData.People[0]
	}
	person.ID = p.ID
	person.FirstName = p.FirstName
	person.Surname = p.Surname
	person.DOB = p.DOB
	person.DOD = p.DOD
	switch p.Gender {
	case 'M':
		person.Gender = 'M'
	case 'F':
		person.Gender = 'F'
	default:
		person.Gender = 'U'
	}
	person.ChildOf = p.ChildOf.ID
	person.SpouseOf = make([]uint, len(p.SpouseOf))
	for n, f := range p.SpouseOf {
		person.SpouseOf[n] = f.ID
	}
	return nil
}

type RPCFamily struct {
	Husband, Wife uint
	Children      []uint
}

func (RPC) GetFamily(id uint, family *RPCFamily) error {
	f, ok := GedcomData.Families[id]
	if !ok {
		f = GedcomData.Families[0]
	}
	family.Husband = f.Husband.ID
	family.Wife = f.Wife.ID
	family.Children = make([]uint, len(f.Children))
	for n, c := range f.Children {
		family.Children[n] = c.ID
	}
	return nil
}
