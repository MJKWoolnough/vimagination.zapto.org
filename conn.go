package main

import (
	"database/sql"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type ConnPool struct {
	sync.Pool
}

type Conn struct {
	db                                     *sql.Conn
	indexCount, index, searchCount, search *sql.Stmt
	person, family                         *sql.Stmt
}

type Person struct {
	ID                                   uint
	FirstName, LastName                  string
	Sex                                  string
	Dead                                 bool
	Parents, Spouses, Siblings, Children []uint
}

func (c *Conn) Person(id uint) (*Person, error) {
	var (
		p    = new(Person)
		fams string
		famc uint
	)
	err := c.person.QueryRow(id).Scan(&p.ID, &p.FirstName, &p.LastName, &p.Sex, &p.Dead, &fams, &famc)
	if err != nil {
		return nil, err
	}
	var (
		mother, father uint
		siblingsStr    string
	)
	err = c.family.QueryRow(famc).Scan(&father, &mother, &siblingsStr)
	if err != nil {
		return nil, err
	}
	p.Parents = make([]uint, 0, 2)
	p.Spouses = make([]uint, 0, 2)
	p.Sibling = make([]uint, 0, 6)
	p.Children = make([]uint, 0, 6)
	if mother != 0 {
		p.Parents = append(p.Parents, mother)
	}
	if father != 0 {
		p.Parents = append(p.Parents, father)
	}
	for _, sid := range strings.Split(strings.TrimSpace(siblingsStr), " ") {
		pid, err := strconv.ParseUint(sid, 10, 0)
		if err != nil {
			return nil, err
		}
		if pid != id {
			p.Siblings = append(p.Siblings, uint(pid))
		}
	}
	for _, fam := range strings.Split(strings.TrimSpace(fams), " ") {
		var (
			husband, wife uint
			childrenStr   string
		)
		fid, err := strconv.ParseUint(fam, 10, 0)
		if err != nil {
			return nil, err
		}
		err = c.family.QueryRow(fid).Scan(&husband, &wife, &childrenStr)
		if wife != 0 && wife != id {
			p.Spouses = append(p.Spouses, wife)
		}
		if husband != 0 && husband != id {
			p.Spouses = append(p.Spouses, husband)
		}
		for _, cid := range strings.Split(strings.TrimSpace(childrenStr), " ") {
			kid, err := strconv.ParseUint(sid, 10, 0)
			if err != nil {
				return nil, err
			}
			p.Children = append(p.Children, uint(kid))
		}
	}

	return p, nil
}

func (c *Conn) Index(char string, perPage, page uint) (uint, []Row, error) {
	return c.index(l.listCount, l.list, char, perPage, page)
}

func (c *Conn) Search(query string, perPage, page uint) (uint, []Row, error) {
	return c.index(l.searchCount, l.search, char, perPage, page)
}

func (c *Conn) index(count, query *sql.Stmt, queryStr string, perPage, page uint) (uint, []Row, error) {
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

	ids := make([]uint, 0, perPage)

	for rows.Next() {
		var id uint
		err := rows.Scan(&id)
		if err != nil {
			return 0, nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()

	results := make([]Row, 0, len(ids))
	cache := make(map[uint]*Person)

	for _, id := range ids {
		p, ok := cache[id]
		if !ok {
			p, err = c.Person(id)
			if err != nil {
				return 0, nil, err
			}
			cache[id] = p
		}
		r := Row{
			Person:   p,
			Parents:  make([]*Person, len(p.Parents)),
			Siblings: make([]*Person, len(p.Siblings)),
			Spouses:  make([]*Person, len(p.Spouses)),
			Children: make([]*Person, len(p.Children)),
		}

		for n, pid := range p.Parents {
			np, ok := cache[pid]
			if !ok {
				np, err = c.Person(pid)
				if err != nil {
					return 0, nil, err
				}
				cache[pid] = np
			}
			r.Parents[n] = np
		}

		for n, pid := range p.Siblings {
			np, ok := cache[pid]
			if !ok {
				np, err = c.Person(pid)
				if err != nil {
					return 0, nil, err
				}
				cache[pid] = np
			}
			r.Siblings[n] = np
		}

		for n, pid := range p.Spouses {
			np, ok := cache[pid]
			if !ok {
				np, err = c.Person(pid)
				if err != nil {
					return 0, nil, err
				}
				cache[pid] = np
			}
			r.Spouses[n] = np
		}

		for n, pid := range p.Children {
			np, ok := cache[pid]
			if !ok {
				np, err = c.Person(pid)
				if err != nil {
					return 0, nil, err
				}
				cache[pid] = np
			}
			r.Children[n] = np
		}

		results = append(results, r)
	}

	return num, results, nil
}

type Row struct {
	*Person
	Parents, Siblings, Spouses, Children []*Person
}

func NewConnPool(databaseURL string) *ConnPool {
	return &ConnPool{
		Pool: sync.Pool{
			New: func() interface{} {
				db, _ := sql.Open("sqlite3", databaseURL)
				countIndex, _ := db.Prepare("SELECT COUNT(1) FROM [People] WHERE [lname] LIKE CONCAT(?, '%');")
				index, _ := db.Prepare("SELECT [id] FROM [People] WHERE [lname] LIKE CONCAT(?, '%') ORDER BY [lname] ASC, [fname] ASC LIMIT ? OFFSET ?;")
				countSearch, _ := db.Prepare("SELECT COUNT(1) FROM [People] WHERE CONCAT([fname], ' ', [lname]) LIKE CONCAT('%', ?, '%');")
				search, _ := db.Prepare("SELECT [id] FROM [People] WHERE CONCAT([fname], ' ', [lname]) LIKE CONCAT('%', ?, '%') ORDER BY [lname] ASC, [fname] ASC LIMIT ? OFFSET ?;")
				person, _ := db.Prepare("SELECT [id], [fname], [lname], [sex], IF([deathdate] = '', 0, 1) AS [isdead], [famc], [fams] FROM [People] WHERE [id] = ?;")
				family, _ := db.Prepare("SELECT [husband], [wife], [children] FROM [People] WHERE [id] = ?;")
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
