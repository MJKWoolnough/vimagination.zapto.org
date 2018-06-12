package main // import "vimagination.zapto.org/vimagination.zapto.org/treeJS"

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
	"vimagination.zapto.org/gopherjs/xjs"
)

var (
	focusID   uint
	highlight = []uint{}
	expandAll bool
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
			fID, err := strconv.ParseUint(v.Get("id"), 10, 64)
			if err != nil {
				fID = 1
			}
			if err := InitRPC(); err != nil {
				xjs.Alert("RPC initialisation failed: %s", err)
				return
			}
			if e := v.Get("highlight"); e != "" {
				ids := strings.Split(e, ",")
				for _, id := range ids {
					uid, err := strconv.ParseUint(id, 10, 64)
					if err != nil {
						continue
					}
					p := GetPerson(uint(uid))
					p.Expand = true
					highlight = append(highlight, uint(uid))
				}
			}
			focusID = uint(fID)
			me := GetPerson(uint(focusID))
			if me.ID != 0 {
				expandAll = v.Get("expandAll") != ""
				me.Expand = true
			}
			DrawTree(true)
		}()
	})
}
