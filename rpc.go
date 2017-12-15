package main

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"

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

type postRWC struct {
	io.Reader
	io.Closer
	io.Writer
	Written bool
}

func (p *postRWC) Write(b []byte) (int, error) {
	p.Written = true
	return p.Writer.Write(b)
}

type cw struct {
	rpc.ServerCodec
	wg sync.WaitGroup
}

func (cw *cw) ReadRequestHeader(r *rpc.Request) error {
	cw.wg.Add(1)
	return cw.ServerCodec.ReadRequestHeader(r)
}

func (cw *cw) WriteResponse(r *rpc.Response, i interface{}) error {
	cw.wg.Done()
	return cw.ServerCodec.WriteResponse(r, i)
}

func (cw *cw) Close() error {
	cw.wg.Wait()
	return cw.ServerCodec.Close()
}

func rpcPostHandler(w http.ResponseWriter, r *http.Request) {
	data := postRWC{
		Reader: io.LimitReader(r.Body, 1<<12),
		Closer: r.Body,
		Writer: w,
	}
	rpc.ServeCodec(&cw{ServerCodec: jsonrpc.NewServerCodec(conn)})
	if !data.Written {
		w.WriteHeader(http.StatusBadRequest)
	}
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
