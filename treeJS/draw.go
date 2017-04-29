package main

import (
	"strconv"

	"honnef.co/go/js/dom"

	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/gopherjs/gopherjs/js"
)

var (
	rows         = new(Rows)
	lines, boxes dom.Node
)

func DrawTree(scroll bool) {
	top := GetPerson(focusID)
	for {
		f := top.ChildOf()
		if h := f.Husband(); h.Expand {
			top = h
			continue
		} else if w := f.Wife(); w.Expand {
			top = w
		} else {
			break
		}
	}
	topFam := top.ChildOf()
	if len(topFam.ChildrenIDs) == 0 {
		topFam = &Family{
			ChildrenIDs: []uint{top.ID},
		}
	}

	lines = xdom.DocumentFragment()
	boxes = xdom.DocumentFragment()

	PersonBox(topFam.Husband(), 0, 0, false)
	PersonBox(topFam.Wife(), 0, 1, true)
	Marriage(0, 0, 1)
	DownLeft(0, 1, 1)
	(&Box{}).Init(0)

	var tree Children
	s := new(Spouse)
	s.Box.Init(0)
	tree.Init(topFam, s, 1)
	tree.Draw()

	xjs.RemoveChildren(xjs.Body())
	xjs.AppendChildren(xjs.Body(), lines)
	xjs.AppendChildren(xjs.Body(), boxes)

	rows.Reset()

	if scroll {
		do := js.Global.Get("document").Get("documentElement")
		js.Global.Call("scrollTo", chosenX-do.Get("clientWidth").Int()/2, chosenY-do.Get("clientHeight").Int()/2)
	}
}

const (
	rowStart = 100
	colStart = 50
	rowGap   = 150
	colGap   = 200
	boxWidth = 150
)

var chosenX, chosenY int

func PersonBox(p *Person, row, col int, spouse bool) {
	d := xdom.Div()
	style := d.Style()
	y, x := rowStart+row*rowGap, colStart+col*colGap
	style.SetProperty("top", strconv.Itoa(y)+"px", "")
	style.SetProperty("left", strconv.Itoa(x)+"px", "")
	class := "person sex_" + string(p.Gender)
	if p.ID == focusID {
		d.SetID("chosen")
		chosenX, chosenY = x, y
	}
	for _, h := range highlight {
		if p.ID == h {
			class += " highlight"
			break
		}
	}
	if len(p.SpouseOfIDs) > 0 {
		collapseExpand := xdom.Div()
		if !p.Expand || spouse {
			collapseExpand.SetClass("expand")
		} else {
			collapseExpand.SetClass("collapse")
		}
		d.AppendChild(collapseExpand)
		d.AddEventListener("click", true, expandCollapse(p, !p.Expand, spouse))
		class += " clicky"
	}
	name := xdom.Div()
	name.SetClass("name")
	xjs.SetInnerText(name, p.FirstName+" "+p.Surname)
	d.AppendChild(name)
	if p.DOB != "" {
		dob := xdom.Div()
		xjs.SetInnerText(dob, "DOB: "+p.DOB)
		dob.SetClass("dob")
		d.AppendChild(dob)
	}
	if p.DOD != "" {
		dod := xdom.Div()
		xjs.SetInnerText(dod, "DOD: "+p.DOD)
		dod.SetClass("dod")
		d.AppendChild(dod)
	}
	d.SetClass(class)
	boxes.AppendChild(d)
}

func expandCollapse(p *Person, expand, spouse bool) func(dom.Event) {
	if spouse {
		return func(dom.Event) {
			focusID = p.ID
			p.Expand = true
			go DrawTree(true)
		}
	}
	return func(dom.Event) {
		p.Expand = expand
		go DrawTree(false)
	}
}

func Marriage(row, start, end int) {
	d := xdom.Div()
	d.SetClass("spouseLine")
	s := d.Style()
	s.SetProperty("left", strconv.Itoa(colStart+start*colGap)+"px", "")
	s.SetProperty("width", strconv.Itoa((end-start)*colGap)+"px", "")
	s.SetProperty("top", strconv.Itoa(rowStart+row*rowGap)+"px", "")
	lines.AppendChild(d)
}

func DownLeft(row, start, end int) {
	downLeft := xdom.Div()
	downLeft.SetClass("downLeft")
	s := downLeft.Style()
	s.SetProperty("top", strconv.Itoa(rowStart+row*rowGap)+"px", "")
	s.SetProperty("left", strconv.Itoa(colStart+start*colGap-125)+"px", "")
	s.SetProperty("width", strconv.Itoa((end-start)*colGap+100)+"px", "")
	lines.AppendChild(downLeft)
}

func DownRight(row, col int) {
	downRight := xdom.Div()
	downRight.SetClass("downRight")
	s := downRight.Style()
	s.SetProperty("top", strconv.Itoa(rowStart+row*rowGap)+"px", "")
	s.SetProperty("left", strconv.Itoa(colStart+col*colGap-25)+"px", "")
	lines.AppendChild(downRight)
}

func SiblingUp(row, col int) {
	down := xdom.Div()
	down.SetClass("downLeft")
	t := down.Style()
	t.SetProperty("top", strconv.Itoa(rowStart+row*rowGap-50)+"px", "")
	t.SetProperty("left", strconv.Itoa(colStart+col*colGap+75)+"px", "")
	t.SetProperty("width", "0px", "")
	t.SetProperty("height", "50px", "")
	lines.AppendChild(down)
}

func SiblingLine(row, start, end int) {
	down := xdom.Div()
	down.SetClass("downLeft")
	t := down.Style()
	t.SetProperty("top", strconv.Itoa(rowStart+row*rowGap-50)+"px", "")
	t.SetProperty("left", strconv.Itoa(colStart+start*colGap+75)+"px", "")
	t.SetProperty("width", strconv.Itoa((end-start)*colGap)+"px", "")
	t.SetProperty("height", "0px", "")
	lines.AppendChild(down)
}
