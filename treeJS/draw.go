package main

import (
	"strconv"

	"honnef.co/go/js/dom"

	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
)

var (
	rows         Rows
	lines, boxes dom.Node
)

type Rows []int

func (r *Rows) GetRow(row int) int {
	if len(*r) <= row {
		return 0
	}
	return (*r)[row]
}

func (r *Rows) SetRow(row, pos int) {
	if len(*r) <= row {
		*r = append(*r, make(Rows, row+1-len(*r))...)
	}
	(*r)[row] = pos
}

func (r *Rows) Reset() {
	*r = (*r)[:0]
}

type Box struct {
	Row            int
	MinCol, MaxCol int
}

func NewBox(row int) Box {
	return Box{
		Row:    row,
		MaxCol: 65535,
	}
}

type Children struct {
	Parents  *Spouse
	Children []Child
	Box
}

func NewChildren(f *Family, parents *Spouse, row int) Children {
	children := f.Children()
	c := Children{
		Parents:  parents,
		Children: make([]Child, len(children)),
		Box:      NewBox(row),
	}
	for n, child := range children {
		c.Children[n] = NewChild(child, &c, row)
	}
	return c
}

type Child struct {
	Siblings *Children
	*Person
	Spouses
	Box
}

func NewChild(p *Person, siblings *Children, row int) Child {
	c := Child{
		Siblings: siblings,
		Person:   p,
		Box:      NewBox(row),
	}
	if p.Expand {
		c.Spouses = NewSpouses(p.SpouseOf(), &c, row)
	}
	return c
}

type Spouses struct {
	Spouse  *Child
	Spouses []Spouse
	Box
}

func NewSpouses(families []*Family, spouse *Child, row int) Spouses {
	s := Spouses{
		Spouse:  spouse,
		Spouses: make([]Spouse, len(families)),
		Box:     NewBox(row),
	}
	for n, f := range families {
		if spouse.Gender == 'F' {
			s.Spouses[n] = NewSpouse(f, f.Husband(), &s, row)
		} else {
			s.Spouses[n] = NewSpouse(f, f.Wife(), &s, row)
		}
	}
	return s
}

type Spouse struct {
	Spouses *Spouses
	*Person
	Children
	Box
}

func NewSpouse(f *Family, p *Person, spouses *Spouses, row int) Spouse {
	s := Spouse{
		Spouses: spouses,
		Person:  p,
		Box:     NewBox(row),
	}
	s.Children = NewChildren(f, &s, row+1)
	return s
}

func DrawTree() {
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
	lines = xdom.DocumentFragment()
	boxes = xdom.DocumentFragment()
	topFam := top.ChildOf()
	boxes.AppendChild(PersonBox(topFam.Husband(), 0, 0, false))
	boxes.AppendChild(PersonBox(topFam.Wife(), 0, 1, true))
	lines.AppendChild(Marriage(0, 0, 1))
	lines.AppendChild(DownLeft(0, 1))

	tree := NewChildren(topFam, nil, 0)
	_ = tree

	xjs.RemoveChildren(xjs.Body())
	xjs.AppendChildren(xjs.Body(), lines)
	xjs.AppendChildren(xjs.Body(), boxes)

	rows.Reset()
}

const (
	rowStart = 100
	colStart = 50
	rowGap   = 200
	colGap   = 200
	boxWidth = 150
)

func PersonBox(p *Person, row, col int, spouse bool) dom.Node {
	name := xdom.Span()
	name.SetClass("name")
	xjs.SetInnerText(name, p.FirstName+" "+p.Surname)
	d := xdom.Div()
	style := d.Style()
	style.SetProperty("top", strconv.Itoa(rowStart+row*rowGap)+"px", "")
	style.SetProperty("left", strconv.Itoa(colStart+col*colGap)+"px", "")
	class := "person sex_" + string(p.Gender)
	if p.ID == focusID {
		class += " chosen"
	} else if len(p.SpouseOfIDs) > 0 {
		collapseExpand := xdom.Div()
		if p.Expand && !spouse {
			collapseExpand.SetClass("collapse")
		}
		d.AppendChild(collapseExpand)
		d.AddEventListener("click", true, expandCollapse(p, !p.Expand, spouse))
		class += " clicky"
		if p.ID == selectedID {
			class += " selected"
		}
	}
	d.SetClass(class)
	d.AppendChild(name)
	return d
}

func expandCollapse(p *Person, expand, spouse bool) func(dom.Event) {
	if spouse {
		return func(dom.Event) {
			focusID = p.ID
			p.Expand = true
			DrawTree()
		}
	} else {
		return func(dom.Event) {
			selectedID = p.ID
			p.Expand = expand
			DrawTree()
		}
	}
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
