package main

import (
	"net/url"
	"strconv"

	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

var selectedID, focusID uint

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
			fID, err := strconv.ParseUint(v.Get("id"), 10, 64)
			if err != nil {
				fID = 1
			}
			if err := InitRPC(); err != nil {
				xjs.Alert("RPC initialisation failed: %s", err)
				return
			}
			focusID = uint(fID)
			me := GetPerson(uint(focusID))
			me.Expand = true
			DrawTree()
		}()
	})
}
