package main

import "net/http"

func (l *List) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lc := l.Pool.Get().(*ListConn)
	defer l.Pool.Put(lc)
}
