package main

import (
	"strconv"
)

type person struct {
	*Person
	side *Person
	from *person
}

type Links struct {
	First, Second []*Person
	Common        *Person
}

func (l Links) Relationship() string {
	up := len(l.First)
	down := len(l.Second)
	diff := up - down
	if diff < 0 {
		diff = -diff
	}
	big := up
	if down > up {
		big = down
	}
	diff++
	return strconv.Itoa(big) + ordinal(big) + " cousin, " + strconv.Itoa(diff) + " times removed"
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

func Calculate(first, second *Person) Links {
	toFind := make([]*person, 2, 1024)
	toFind[0] = &person{first, first, nil}
	toFind[1] = &person{second, second, nil}
	cache := make(map[uint]*person)
	cache[first.ID] = toFind[0]
	cache[second.ID] = toFind[1]
	var parents [2]*Person
	for len(toFind) > 0 {
		p := toFind[0]
		toFind = toFind[1:]
		parents[0] = p.ChildOf.Husband
		parents[1] = p.ChildOf.Wife
		for _, next := range parents {
			if next.ID == 0 {
				continue
			}
			if got, ok := cache[next.ID]; ok {
				if got.side == p.side {
					continue
				}
				if got.side == first {
					return Links{
						First:  makeRoute(got.from),
						Second: makeRoute(p),
						Common: got.Person,
					}
				}
				return Links{
					First:  makeRoute(p),
					Second: makeRoute(got.from),
					Common: got.Person,
				}
			}
			np := &person{next, p.side, p}
			toFind = append(toFind, np)
			cache[np.ID] = np
		}
	}
	return Links{}
}

func makeRoute(p *person) []*Person {
	route := make([]*Person, 0, 32)
	for {
		route = append(route, p.Person)
		if p.Person == p.side {
			break
		}
		p = p.from
	}
	reverse(route)
	return route
}

func reverse(data []*Person) {
	ld := len(data)
	for i := 0; i < ld>>1; i++ {
		j := ld - i - 1
		data[i], data[j] = data[j], data[i]
	}
}
