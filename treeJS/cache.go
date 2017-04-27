package main

var (
	personCache = map[uint]*Person{
		0: &Person{FirstName: "?", Surname: "?", Gender: 'U', ChildOfID: 0},
	}
	familyCache = map[uint]*Family{
		0: &Family{},
	}
)

func GetPerson(id uint) *Person {
	if p, ok := personCache[id]; ok {
		return p
	}
	p := RPC.GetPerson(id)
	if expandAll {
		p.Expand = true
	}
	personCache[id] = &p
	return &p
}

func GetFamily(id uint) *Family {
	if f, ok := familyCache[id]; ok {
		return f
	}
	f := RPC.GetFamily(id)
	familyCache[id] = &f
	return &f
}
