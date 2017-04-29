package main

import (
	"strconv"
)

const (
	relParent relationship = false
	relChild  relationship = true
)

type relationship bool

type person struct {
	*Person
	side *Person
	from *personData
}

type personData struct {
	*Person
	RelationshipToNext relationship
}

type Links struct {
	Route    []personData
	Up, Down int
	Half     bool
}

func (l *Links) String() string {
	return l.Route[0].FName + " " + l.Route[0].LName + " is the " + l.Relationship() + " of " + l.Route[len(l.Route)-1].FName + " " + l.Route[len(l.Route)-1].LName
}

func (l *Links) Relationship() string {
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

func Calculate(first, second *Person) *Links {
	toFind := make([]*person, 2, 1024)
	toFind[0] = &person{first, first, nil}
	toFind[1] = &person{second, second, nil}
	cache := make(map[uint]*person)
	cache[first.ID] = toFind[0]
	cache[second.ID] = toFind[1]
	var parents [2]*Person
	for len(tf) > 0 {
		p := toFind[0]
		toFind = toFind[1:]
		parents[0] = p.ChildOf.Husband
		parents[1] = p.ChildOf.Wife
		for _, next := range parents {
			if next.ID == 0 {
				continue
			}
			np := &person{next, p.side, p}
			if got, ok := cache[next.ID]; ok {
				if got.side == p.side {
					continue
				}
				if got.side == first {
					return makeLinks(got, np)
				}
				return makeLinks(np, got)
			}
			toFind = append(toFind, np)
			cache[np.ID] = np
		}
	}
	return nil
}

func makeLinks(first, second *person) Links {
	return Links{}
}

func reverse(data []personData) {
	ld := len(data)
	for i := 0; i < ld>>1; i++ {
		j := ld - i - 1
		data[i], data[j] = data[j], data[i]
		data[i].RelationshipToNext = !data[i].RelationshipToNext
		data[j].RelationshipToNext = !data[j].RelationshipToNext
	}
	if ld&1 == 1 {
		pos := ld>>1 - 1
		data[pos].RelationshipToNext = !data[pos].RelationshipToNext
	}
}
