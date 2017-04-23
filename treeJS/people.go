package main

type Person struct {
	ID                 uint
	FirstName, Surname string
	Gender             byte
	ChildOfID          uint   `json:"ChildOf"`
	SpouseOfIDs        []uint `json:"SpouseOf"`
	Expand             bool
}

func (p Person) SpouseOf() []*Family {
	families := make([]*Family, len(p.SpouseOfIDs))
	for n, fid := range p.SpouseOfIDs {
		families[n] = GetFamily(fid)
	}
	return families
}

func (p Person) ChildOf() *Family {
	return GetFamily(p.ChildOfID)
}

type Family struct {
	HusbandID   uint   `json:"Husband"`
	WifeID      uint   `json:"Wife"`
	ChildrenIDs []uint `json:"Children"`
}

func (f Family) Husband() *Person {
	return GetPerson(f.HusbandID)
}

func (f Family) Wife() *Person {
	return GetPerson(f.WifeID)
}

func (f Family) Children() []*Person {
	people := make([]*Person, len(f.ChildrenIDs))
	for n, pid := range f.ChildrenIDs {
		people[n] = GetPerson(pid)
	}
	return people
}
