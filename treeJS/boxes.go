package main

type Rows []int

func (r *Rows) GetRow(row int) int {
	if len(*r) <= row {
		return 0
	}
	return (*r)[row]
}

func (r *Rows) RowPP(row int) int {
	for len(*r) <= row {
		*r = append(*r, make(Rows, row+1-len(*r))...)
	}
	c := (*r)[row]
	(*r)[row]++
	return c
}

func (r *Rows) Reset() {
	*r = (*r)[:0]
}

type Box struct {
	Row, Col int
}

func NewBox(row int) Box {
	return Box{
		Row: row,
		Col: rows.RowPP(row),
	}
}

func (b *Box) FirstEmpty() {
	if b.Col < rows.GetRow(b.Row) {
		b.Col = rows.RowPP(b.Row)
	}
}

type Children struct {
	Parents  *Spouse
	Children []Child
}

func NewChildren(f *Family, parents *Spouse, row int) Children {
	children := f.Children()
	c := Children{
		Parents:  parents,
		Children: make([]Child, len(children)),
	}
	for n, child := range children {
		c.Children[n] = NewChild(child, &c, row)
	}
	return c
}

func (c *Children) Shift() {

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
}

func NewSpouses(families []*Family, spouse *Child, row int) Spouses {
	s := Spouses{
		Spouse:  spouse,
		Spouses: make([]Spouse, len(families)),
	}
	for n, f := range families {
		if spouse.Gender == 'F' {
			s.Spouses[n] = NewSpouse(f, f.Husband(), &s, row)
		} else {
			s.Spouses[n] = NewSpouse(f, f.Wife(), &s, row)
		}
	}
	if len(families) > 0 {
		spouse.Col = s.Spouses[0].Col - 1
	}
	return s
}

func (s *Spouses) Shake() {
	for _, spouse := range s.Spouses {
		s.Shake()
	}
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

func (s *Spouse) Shake() {
	s.FirstEmpty()
	s.Children.Shake()
}
