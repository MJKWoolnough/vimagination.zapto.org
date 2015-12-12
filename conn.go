package main

import (
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/mxk/go-sqlite/sqlite3"
)

type ConnPool struct {
	sync.Pool
}

type Conn struct {
	db                                     *sqlite3.Conn
	indexCount, index, searchCount, search *sqlite3.Stmt
	person, family                         *sqlite3.Stmt
}

func (l *Conn) Index(char string, perPage, page uint) (uint, []Row, error) {
	return l.index(l.listCount, l.list, char, perPage, page)
}

func (l *Conn) Search(query string, perPage, page uint) (uint, []Row, error) {
	return l.index(l.searchCount, l.search, char, perPage, page)
}

func (l *Conn) index(count, query *sqllit3.Stmt, queryStr string, perPage, page uint) (uint, []Row, error) {
	var num uint
	err := count.QueryRow(char).Scan(&num)
	if err != nil {
		return 0, nil, err
	}
	if num == 0 || num < page*perPage {
		return num, []Row{}, nil
	}
	rows, err := query.Query(queryStr, perPage, page*perPage)
	if err != nil {
		return 0, nil, err
	}

	result := make([]Row, 0, perPage)
	cache := make(map[uint]*Person)

	getPerson := func(id uint) (*Person, error) {
		if p, ok := cache[id]; ok {
			return p, nil
		}
		var fams string
		p := new(Person)
		err := l.person.QueryRow.Scan(&p.ID, &p.FirstName, &p.LastName, &p.Sex, &p.Dead, &p.Famc, &fams)
		if err != nil {
			return nil, err
		}
		for _, fam := range strings.Split(strings.TrimSpace(fams), " ") {
			id, err := strconv.ParseUint(fam, 10, 0)
			if err != nil {
				return 0, nil, err
			}
			p.Fams = append(p.Fams, uint(id))
		}
		cache[id] = p
		return p, nil
	}

	for rows.Next() {
		var fams string
		p := new(Person)
		err := rows.Scan(&p.ID, &p.FirstName, &p.LastName, &p.Sex, &p.Dead, &p.Famc, &fams)
		if err != nil {
			return 0, nil, err
		}
		for _, fam := range strings.Split(strings.TrimSpace(fams), " ") {
			id, err := strconv.ParseUint(fam, 10, 0)
			if err != nil {
				return 0, nil, err
			}
			p.Fams = append(p.Fams, uint(id))
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, err
	}
	for n := range result {
		var (
			father, mother uint
			siblings       string
		)
		err = l.family.QueryRow(result[n].Famc).Scan(&father, &mother, &siblings)
		if err != nil {
			return 0, nil, err
		}
		for _, fid := range result[n].Fams {
			var (
				husband, wife uint
				children      string
			)
			err = l.family.QueryRow(fid).Scan(&husband, &wife, &children)
			if err != nil {
				return 0, nil, err
			}

		}
	}
	return num, result, nil
}

type Person struct {
	ID                  uint
	FirstName, LastName string
	Sex                 string
	Dead                bool
	Famc                uint
	Fams                []uint
}

type Row struct {
	*Person
	Parents, Siblings, Children, Spouses []*Person
}

func NewConnPool(databaseURL string) *ConnPool {
	return &ConnPool{
		Pool: sync.Pool{
			New: func() interface{} {
				const (
					personTerms = "[id], [fname], [lname], [sex], IF([deathdate] = '', 0, 1) AS [isdead], [famc], [fams]"
					familyTerms = "[husband], [wife], [children]"
				)
				db, _ := sqlite3.Open(databaseURL)
				countIndex, _ := db.Prepare("SELECT COUNT(1) FROM [People] WHERE [lname] LIKE CONCAT(?, '%');")
				index, _ := db.Prepare("SELECT " + personTerms + " FROM [People] WHERE [lname] LIKE CONCAT(?, '%') ORDER BY [lname] ASC, [fname] ASC LIMIT ? OFFSET ?;")
				countSearch, _ := db.Prepare("SELECT COUNT(1) FROM [People] WHERE CONCAT([fname], ' ', [lname]) LIKE CONCAT('%', ?, '%');")
				search, _ := db.Prepare("SELECT " + personTerms + " FROM [People] WHERE CONCAT([fname], ' ', [lname]) LIKE CONCAT('%', ?, '%') ORDER BY [lname] ASC, [fname] ASC LIMIT ? OFFSET ?;")
				person, _ := db.Prepare("SELECT " + personTerms + " FROM [People] WHERE [id] = ?;")
				family, _ := db.Prepare("SELECT " + familyTerms + " FROM [People] WHERE [id] = ?;")
				l := &Conn{
					db:          db,
					countIndex:  countIndex,
					index:       index,
					countSearch: countSearch,
					search:      search,
					person:      person,
					family:      family,
				}
				runtime.SetFinalizer(l, closeConn)
				return l
			},
		},
	}
}

func closeConn(l *Conn) {
	l.listCount.Close()
	l.list.Close()
	l.searchCount.Close()
	l.search.Close()
	l.person.Close()
	l.family.Close()
	l.db.Close()
}
