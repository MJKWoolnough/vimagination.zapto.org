package main

type Rows []int

func (r *Rows) GetRow(row int) int {
	if len(*r) <= row {
		return 0
	}
	return (*r)[row]
}

func (r *Rows) RowPP(row int) int {
	r.size(row)
	c := (*r)[row]
	(*r)[row]++
	return c
}

func (r *Rows) SetRow(row, col int) {
	r.size(row)
	(*r)[row] = col
}

func (r *Rows) size(row int) {
	for len(*r) <= row {
		*r = append(*r, make(Rows, row+1-len(*r))...)
	}
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

func (b *Box) SetCol(col int) {
	if rows.GetRow(b.Row) < col {
		rows.SetRow(b.Row, col)
	}
}

func (b *Box) AddCol(diff int) {
	b.SetCol(b.Col + diff)
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
	if parents != nil {
		c.Shift(diff)
	}
	return c
}

func (c *Children) Shift(diff int) bool {
	if len(c.Children) > 0 {
		if pDiff := parents.Col + diff - 1 - c.Children[len(c.Children)-1].LastX(); pDiff > 0 {
			for i := len(c.Children) - 1; c >= 0; c-- {
				if !c.Children[i].Shift(pDiff) {
					return false
				}
			}
		}
	}
	return true
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

func (c *Child) LastX() int {
	if len(c.Spouses.Spouses) > 0 {
		return c.Spouses.Spouses[len(c.Spouses.Spouses)-1].Col
	}
	return c.Col
}

func (c *Child) Shift(diff int) bool {
	if !c.Spouses.Shift(diff) {
		return false
	}
	if len(s.Spouses) > 0 {
		s.Spouse.Col = s.Spouses[0].Col - 1
	} else {
		c.AddCol(diff)
	}
	return true
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
func (s *Spouses) Shift(diff int) bool {
	for i := len(c.Spouses.Spouses) - 1; i >= 0; i-- {
		if !c.Spouses.Spouses[i].Shift(diff) {
			return false
		}
	}
	return true
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
	if len(f.ChildrenIDs) > 0 {
		lastChildPos := s.Children.Children[len(s.Children.Children)-1].Col
		if s.Col+1 > lastChildPos {
			s.SetCol(lastChildPos + 1)
		}
	}
	return s
}

func (s *Spouse) Shift(diff int) bool {
	if !s.Children.Shift(diff) {
		return false
	}
	s.AddCol(diff)
	return true
}
