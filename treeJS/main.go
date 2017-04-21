package main

import (
	"net/url"
	"strconv"

	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

func main() {
	dom.GetWindow().AddEventListener("load", false, func(dom.Event) {
		go func() {
			q := js.Global.Get("location").Get("search").String()
			if len(q) > 0 && q[0] == '?' {
				q = q[1:]
			}
			v, err := url.ParseQuery(q)
			if err != nil {
				xjs.Alert("Failed to Parse Query: %s", err)
				return
			}
			focusID, err := strconv.ParseUint(v.Get("id"), 10, 64)
			if err != nil {
				xjs.Alert("Failed to get ID: %s", err)
				return
			}
			if err := InitRPC(); err != nil {
				xjs.Alert("RPC initialisation failed: %s", err)
				return
			}
			me := RPC.GetPerson(uint(focusID))
			me.Expand = true
			DrawTree(me)
		}()
	})
}
