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
	Route    []*Person
	Up, Down int
	Half     bool
}

func (l *Links) Relationship() string {
	diff := l.Up - l.Down
	if diff < 0 {
		diff = -diff
	}
	big := l.Up
	if l.Down > l.Up {
		big = l.Down
	}
	diff++
	return strconv.Itoa(big) + ordinal(big) + " cousin, " + strconv.Itoa(diff) + " times removed"
}

func (l *Links) Reverse() *Links {
	route := make([]*Person, len(l.Route))
	copy(route, l.Route)
	reverse(route)
	return &Links{
		Route: route,
		Down:  l.Up,
		Up:    l.Down,
		Half:  l.Half,
	}
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
	for len(toFind) > 0 {
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

func makeLinks(first, second *person) *Links {
	var (
		up, down int
		half     bool
		p        *person
	)

	route := make([]*Person, 0, 32)

	p = first
	for {
		route = append(route, p.Person)
		if p.Person == p.side {
			break
		}
		p = p.from
		up++
	}

	reverse(route)

	p = second
	for {
		if p.Person == p.side {
			break
		}
		p = p.from
		route = append(route, p.Person)
		down++
	}

	return &Links{
		Route: route,
		Up:    up,
		Down:  down,
		Half:  half,
	}
}

func reverse(data []*Person) {
	ld := len(data)
	for i := 0; i < ld>>1; i++ {
		j := ld - i - 1
		data[i], data[j] = data[j], data[i]
	}
}
