package main

import (
	"strconv"
)

const (
	relParent relationship = iota
	relChild
	relSibling
	relMarriage
)

type relationship uint8

func (r relationship) Reverse() relationship {
	switch r {
	case relParent:
		return relChild
	case relChild:
		return relParent
	default:
		return r
	}
}

type Calculator map[uint]*person

type person struct {
	id, side uint
	from     *person
}

type personData struct {
	Gender             byte
	RelationshipToNext relationship
	ID                 uint
	FName, LName       string
}

type links struct {
	Route    []personData
	Up, Diff int
}

func (l *links) String() string {
	return l.Route[0].FName + " " + l.Route[0].LName + " is the " + l.Relationship() + " of " + l.Route[len(l.Route)-1].FName + " " + l.Route[len(l.Route)-1].LName
}

func (l *links) Relationship() string {
	diff := l.Diff
	if diff < 0 {
		diff = -diff
	}
	diff++
	return strconv.Itoa(l.Up) + ordinal(l.Up) + " cousin, " + strconv.Itoa(diff) + " times removed"
}

func ordinal(num int) string {
	switch num % 100 {
	case 11, 12, 13:
		return "th"
	default:
		switch num % 10 {
		case 1:
			return "st"
		case 2:
			return "nd"
		case 3:
			return "rd"
		default:
			return "th"
		}
	}
}

func GetParents(id uint) (uint, uint) {
	p, ok := GedcomData.People[id]
	if !ok {
		return 0, 0
	}
	var mother, father uint
	if p.ChildOf != nil {
		if p.ChildOf.Wife != nil {
			mother = p.ChildOf.Wife.ID
		}
		if p.ChildOf.Husband != nil {
			father = p.ChildOf.Husband.ID
		}
	}
	return mother, father
}

func (f finder) findDirectRelationship(fid, sid uint) (*links, error) {
	tf := make([]person, 2, 1024)
	tf[0] = person{fid, fid, nil}
	tf[1] = person{sid, sid, nil}
	for len(tf) > 0 {
		p := tf[0]
		tf = tf[1:]
		mother, father, err := GetParents(p.id)
		for _, parent := range [2]uint{mother, father} {
			if parent == 0 {
				continue
			}
			pid := uint(parent.Int64)
			if pid == 0 {
				continue
			}
			if got, ok := f.data[pid]; ok {
				if got.side == p.side {
					continue
				}
				var first, second *person
				if got.side == fid {
					first = got
					second = &p
				} else {
					first = &p
					second = got
				}
				var data []personData
				cf, err := f.followRouteDown(first, &data)
				if err != nil {
					return nil, err
				}
				reverse(data)
				cs, err := f.followRouteDown(second, &data)
				if err != nil {
					return nil, err
				}
				diff := cf - cs
				if cs > cf {
					cs, cf = cf, cs
				}
				return &links{data, cs - 2, diff}, nil
			}
			tf = append(tf, person{pid, p.side, &p})
			f.data[pid] = &tf[len(tf)-1]
		}
	}
	return nil, nil
}

func (f finder) followRouteDown(from *person, data *[]personData) (int, error) {
	count := 0
	for ; from != nil; from = from.from {
		var fname, lname, sex string

		err := f.getData.QueryRow(from.id).Scan(&fname, &lname, &sex)
		if err != nil {
			return count, err
		}
		*data = append(*data, personData{
			ID:                 from.id,
			FName:              fname,
			LName:              lname,
			Gender:             sex[0],
			RelationshipToNext: relParent,
		})
		count++
	}
	return count, nil
}

func reverse(data []personData) {
	ld := len(data)
	for i := 0; i < ld>>1; i++ {
		j := ld - i - 1
		data[i].FName, data[j].FName = data[j].FName, data[i].FName
		data[i].LName, data[j].LName = data[j].LName, data[i].LName
		data[i].Gender, data[j].Gender = data[j].Gender, data[i].Gender
		data[i].ID, data[j].ID = data[j].ID, data[i].ID
		data[i].RelationshipToNext, data[j].RelationshipToNext = data[j].RelationshipToNext.Reverse(), data[i].RelationshipToNext.Reverse()
	}
	if ld&1 == 1 {
		pos := ld>>1 - 1
		data[pos].RelationshipToNext = data[pos].RelationshipToNext.Reverse()
	}
}
