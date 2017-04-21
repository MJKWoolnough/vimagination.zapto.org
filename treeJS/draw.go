package main

import (
	"strconv"

	"honnef.co/go/js/dom"

	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
)

var (
	rows         [1024]int
	lines, boxes dom.Node
)

func DrawTree(p Person) {
	top := p
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
	rows = [1024]int{3}
	lines = xdom.DocumentFragment()
	boxes = xdom.DocumentFragment()
	boxes.AppendChild(PersonBox(top.ChildOf().Husband(), 0, 0))
	boxes.AppendChild(PersonBox(top.ChildOf().Wife(), 0, 1))
	lines.AppendChild(Marriage(0, 0, 1))
	lines.AppendChild(DownLeft(0, 1))

	Process(top.ChildOf(), 1, 0)

	xjs.RemoveChildren(xjs.Body())
	xjs.AppendChildren(xjs.Body(), lines)
	xjs.AppendChildren(xjs.Body(), boxes)
}

const (
	rowStart = 100
	colStart = 50
	rowGap   = 200
	colGap   = 200
	boxWidth = 150
)

func Process(f Family, row, col int) {
	for _, child := range f.Children() {
		// draw decendants first, left aligned
	}
}

func PersonBox(p Person, row, col int) dom.Node {
	name := xdom.Span()
	name.SetClass("name")
	xjs.SetInnerText(name, p.FirstName+" "+p.Surname)
	d := xdom.Div()
	xjs.AppendChildren(d, name)
	d.SetClass("person sex_" + string(p.Gender))
	style := d.Style()
	style.SetProperty("top", strconv.Itoa(rowStart+row*rowGap)+"px", "")
	style.SetProperty("left", strconv.Itoa(colStart+col*colGap)+"px", "")
	return d
}

func Marriage(row, start, end int) dom.Node {
	d := xdom.Div()
	d.SetClass("spouseLine")
	s := d.Style()
	s.SetProperty("left", strconv.Itoa(colStart+start*colGap)+"px", "")
	s.SetProperty("width", strconv.Itoa((end-start)*colGap)+"px", "")
	s.SetProperty("top", strconv.Itoa(rowStart+row*rowGap)+"px", "")
	return d
}

func DownLeft(row, col int) dom.Node {
	frag := xdom.DocumentFragment()
	downLeft := xdom.Div()
	downLeft.SetClass("downLeft")
	s := downLeft.Style()
	s.SetProperty("top", strconv.Itoa(rowStart+row*rowGap)+"px", "")
	s.SetProperty("left", strconv.Itoa(colStart+col*rowGap-125)+"px", "")
	frag.AppendChild(downLeft)
	down := xdom.Div()
	down.SetClass("downLeft")
	t := down.Style()
	t.SetProperty("top", strconv.Itoa(rowStart+row*rowGap+85)+"px", "")
	t.SetProperty("left", strconv.Itoa(colStart+col*rowGap-125)+"px", "")
	t.SetProperty("width", "0px", "")
	t.SetProperty("height", "100px", "")
	frag.AppendChild(down)
	return frag
}

func DownRight(row, col int) dom.Node {
	frag := xdom.DocumentFragment()
	downRight := xdom.Div()
	downRight.SetClass("downRight")
	s := downRight.Style()
	s.SetProperty("top", strconv.Itoa(rowStart+row*rowGap)+"px", "")
	s.SetProperty("left", strconv.Itoa(colStart+col*rowGap-25)+"px", "")
	frag.AppendChild(downRight)
	down := xdom.Div()
	down.SetClass("downLeft")
	t := down.Style()
	t.SetProperty("top", strconv.Itoa(rowStart+row*rowGap+85)+"px", "")
	t.SetProperty("left", strconv.Itoa(colStart+col*rowGap-25)+"px", "")
	t.SetProperty("width", "0px", "")
	t.SetProperty("height", "100px", "")
	frag.AppendChild(down)
	return frag
}

func SiblingStart(row, start, end int) dom.Node {
	return xdom.DocumentFragment()
}

func Siblings(row, start, end int) dom.Node {
	frag := xdom.DocumentFragment()
	top := strconv.Itoa(rowStart+row*rowGap-85) + "px"
	for i := start; i < end; i++ {
		d := xdom.Div()
		d.SetClass("upLeft")
		s := d.Style()
		s.SetProperty("left", strconv.Itoa(colStart+i*colGap)+"px", "")
		s.SetProperty("top", top, "")
		frag.AppendChild(d)
	}
	return frag
}
