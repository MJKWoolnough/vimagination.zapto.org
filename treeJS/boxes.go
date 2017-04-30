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
		*r = append(*r, 0)
	}
}

func (r *Rows) Reset() {
	*r = (*r)[:0]
}

type Box struct {
	Row, Col int
}

func (b *Box) Init(row int) {
	b.Row = row
	b.Col = rows.RowPP(row)
}

func (b *Box) SetCol(col int) {
	b.Col = col
	if col >= rows.GetRow(b.Row) {
		rows.SetRow(b.Row, col+1)
	}
}

func (b *Box) AddCol(diff int) {
	b.SetCol(b.Col + diff)
}

type Children struct {
	Parents  *Spouse
	Children []Child
}

func (c *Children) Init(f *Family, parents *Spouse, row int) {
	children := f.Children()
	c.Parents = parents
	c.Children = make([]Child, len(children))
	for n, child := range children {
		c.Children[n].Init(child, c, row)
	}
	if parents.Spouses != nil {
		c.Shift(0)
	}
}

func (c *Children) Shift(diff int) bool {
	if len(c.Children) > 0 {
		if pDiff := c.Parents.Col + diff - 1 - c.Children[len(c.Children)-1].LastX(); pDiff > 0 {
			for i := len(c.Children) - 1; i >= 0; i-- {
				if !c.Children[i].Shift(pDiff) {
					return false
				}
			}
		}
	}
	return true
}

func (c *Children) Draw() {
	if len(c.Children) > 1 {
		SiblingLine(c.Children[0].Row, c.Children[0].Col, c.Children[len(c.Children)-1].Col)
	}
	for _, child := range c.Children {
		child.Draw()
	}
}

type Child struct {
	Siblings *Children
	*Person
	Spouses
	Box
}

func (c *Child) Init(p *Person, siblings *Children, row int) {
	c.Siblings = siblings
	c.Person = p
	c.Box.Init(row)
	if p.Expand {
		c.Spouses.Init(p.SpouseOf(), c, row)
	}
}

func (c *Child) LastX() int {
	if len(c.Spouses.Spouses) > 0 {
		return c.Spouses.Spouses[len(c.Spouses.Spouses)-1].Col
	}
	return c.Col
}

func (c *Child) Shift(diff int) bool {
	if len(c.Spouses.Spouses) > 0 {
		if !c.Spouses.Shift(diff) {
			return false
		}
		c.SetCol(c.Spouses.Spouses[0].Col - 1)
	} else {
		c.AddCol(diff)
	}
	return true
}

func (c *Child) Draw() {
	SiblingUp(c.Row, c.Col)
	PersonBox(c.Person, c.Row, c.Col, false)
	c.Spouses.Draw()
}

type Spouses struct {
	Spouse  *Child
	Spouses []Spouse
}

func (s *Spouses) Init(families []*Family, spouse *Child, row int) {
	s.Spouse = spouse
	s.Spouses = make([]Spouse, len(families))
	for n, f := range families {
		if spouse.Gender == 'F' {
			s.Spouses[n].Init(f, f.Husband(), s, row)
		} else {
			s.Spouses[n].Init(f, f.Wife(), s, row)
		}
	}
	if len(families) > 0 {
		spouse.Col = s.Spouses[0].Col - 1
	}
}
func (s *Spouses) Shift(diff int) bool {
	for i := len(s.Spouses) - 1; i >= 0; i-- {
		if !s.Spouses[i].Shift(diff) {
			return false
		}
	}
	return true
}

func (s *Spouses) Draw() {
	if len(s.Spouses) > 0 {
		Marriage(s.Spouse.Row, s.Spouse.Col, s.Spouses[len(s.Spouses)-1].Col)
		for _, spouse := range s.Spouses {
			spouse.Draw()
		}
	}
}

type Spouse struct {
	Spouses *Spouses
	*Person
	Children
	Box
}

func (s *Spouse) Init(f *Family, p *Person, spouses *Spouses, row int) {
	s.Spouses = spouses
	s.Person = p
	s.Box.Init(row)
	s.Children.Init(f, s, row+1)
	if len(f.ChildrenIDs) > 0 {
		firstChildPos := s.Children.Children[0].Col
		if s.Col < firstChildPos {
			s.SetCol(firstChildPos)
		}
	}
}

func (s *Spouse) Shift(diff int) bool {
	all := true
	if len(s.Children.Children) > 0 {
		all = s.Children.Shift(diff)
	}
	s.AddCol(diff)
	return all
}

func (s *Spouse) Draw() {
	PersonBox(s.Person, s.Row, s.Col, true)
	if len(s.Children.Children) > 0 {
		if s.Col == s.Children.Children[0].Col {
			DownRight(s.Row, s.Col)
		} else if s.Col > s.Children.Children[len(s.Children.Children)-1].Col {
			DownLeft(s.Row, s.Children.Children[len(s.Children.Children)-1].Col+1, s.Col)
		} else {
			DownLeft(s.Row, s.Col, s.Col)
		}
		s.Children.Draw()
	}
}
