package main

import (
	"net/rpc/jsonrpc"

	"github.com/MJKWoolnough/errors"
	"golang.org/x/net/websocket"
)

func rpcHandler(conn *websocket.Conn) {
	jsonrpc.ServeConn(conn)
}

type RPC struct{}

type RPCPerson struct {
	ID                 uint
	FirstName, Surname string
	Gender             byte
	ChildOf            uint
	SpouseOf           []uint
}

func (RPC) GetPerson(id uint, person *RPCPerson) error {
	p, ok := GedcomData.People[id]
	if !ok {
		return errors.Error("unknown id")
	}
	person.ID = p.ID
	person.FirstName = p.FirstName
	person.Surname = p.Surname
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
		return errors.Error("unknown id")
	}
	family.Husband = f.Husband.ID
	family.Wife = f.Wife.ID
	family.Children = make([]uint, len(f.Children))
	for n, c := range f.Children {
		family.Children[n] = c.ID
	}
	return nil
}
