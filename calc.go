package main

import (
	"strconv"
	"strings"
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
	var relationship string
	if up > 0 && down > 0 {
		if l.First[len(l.First)-1].ChildOf != l.Second[len(l.Second)-1].ChildOf {
			relationship = "Half-"
		}
	}
	switch up {
	case 0:
		switch down {
		case 0:
			return "Clone"
		case 1:
		default:
			greats := down - 2
			if greats > 3 {
				relationship = strconv.Itoa(greats) + " x Great-Grand-"
			} else {
				relationship = strings.Repeat("Great-", greats) + "Grand-"
			}
		}
		switch l.Common.Gender {
		case 'M':
			relationship += "Father"
		case 'F':
			relationship += "Mother"
		default:
			relationship += "Parent"
		}
	case 1:
		switch down {
		case 0:
			switch l.First[0].Gender {
			case 'M':
				relationship = "Son"
			case 'F':
				relationship = "Daughter"
			default:
				relationship = "Child"
			}
		case 1:
			switch l.First[0].Gender {
			case 'M':
				relationship += "Brother"
			case 'F':
				relationship += "Sister"
			default:
				relationship += "Sibling"
			}
		default:
			greats := down - 2
			if greats > 3 {
				relationship += strconv.Itoa(greats) + " x Great-"
			} else {
				relationship += strings.Repeat("Great-", greats)
			}
			switch l.First[0].Gender {
			case 'M':
				relationship += "Uncle"
			case 'F':
				relationship += "Aunt"
			default:
				relationship += "Pibling"
			}
		}
	default:
		switch down {
		case 0:
			greats := up - 2
			if greats > 3 {
				relationship = strconv.Itoa(greats) + " x Great-Grand-"
			} else {
				relationship = strings.Repeat("Great-", greats) + "Grand-"
			}
			switch l.First[0].Gender {
			case 'M':
				relationship += "Son"
			case 'F':
				relationship += "Daughter"
			default:
				relationship += "Child"
			}
		case 1:
			greats := up - 2
			if greats > 3 {
				relationship += strconv.Itoa(greats) + " x Great-"
			} else {
				relationship += strings.Repeat("Great-", greats)
			}
			switch l.First[0].Gender {
			case 'M':
				relationship += "Nephew"
			case 'F':
				relationship += "Neice"
			default:
				relationship += "Nibling"
			}
		default:
			var small, diff int
			if up > down {
				small = down - 1
				diff = up - down
			} else {
				small = up - 1
				diff = down - up
			}
			relationship += strconv.Itoa(small) + ordinal(small) + " Cousin"
			if diff > 0 {
				relationship += ", " + removed(diff) + " removed,"
			}
		}
	}
	return relationship
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

func removed(diff int) string {
	switch diff {
	case 1:
		return "once"
	case 2:
		return "twice"
	case 3:
		return "thrice"
	}
	return strconv.Itoa(diff) + " times"
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
		if p == nil {
			break
		}
		route = append(route, p.Person)
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
