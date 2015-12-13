package main

import (
	"database/sql"
	"runtime"
	"strconv"
	"strings"
	"sync"

	_ "github.com/mxk/go-sqlite/sqlite3"
)

type ConnPool struct {
	sync.Pool
}

type Conn struct {
	db                                     *sql.DB
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
	err := c.person.QueryRow(id).Scan(&p.ID, &p.FirstName, &p.LastName, &p.Sex, &p.Dead, &famc, &fams)
	if err != nil {
		return nil, err
	}
	var (
		mother, father uint
		siblingsStr    string
	)
	if famc > 0 {
		err = c.family.QueryRow(famc).Scan(&father, &mother, &siblingsStr)
		if err != nil {
			return nil, err
		}
	}
	p.Parents = make([]uint, 0, 2)
	p.Spouses = make([]uint, 0, 2)
	p.Siblings = make([]uint, 0, 6)
	p.Children = make([]uint, 0, 6)
	if mother != 0 {
		p.Parents = append(p.Parents, mother)
	}
	if father != 0 {
		p.Parents = append(p.Parents, father)
	}
	for _, sid := range strings.Split(strings.TrimSpace(siblingsStr), " ") {
		if sid == "" {
			continue
		}
		pid, err := strconv.ParseUint(sid, 10, 0)
		if err != nil {
			return nil, err
		}
		if uint(pid) != id {
			p.Siblings = append(p.Siblings, uint(pid))
		}
	}
	for _, fam := range strings.Split(strings.TrimSpace(fams), " ") {
		if fam == "" {
			continue
		}
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
			if cid == "" {
				continue
			}
			kid, err := strconv.ParseUint(cid, 10, 0)
			if err != nil {
				return nil, err
			}
			p.Children = append(p.Children, uint(kid))
		}
	}

	return p, nil
}

func (c *Conn) Index(char string, perPage, page uint) (uint, []Row, error) {
	return c.getIndex(c.indexCount, c.index, char, perPage, page)
}

func (c *Conn) Search(query string, perPage, page uint) (uint, []Row, error) {
	return c.getIndex(c.searchCount, c.search, query, perPage, page)
}

type PersonCache struct {
	c     *Conn
	cache map[uint]*Person
}

func (pc *PersonCache) Get(ids ...uint) ([]*Person, error) {
	toRet := make([]*Person, len(ids))
	var err error
	for n, pid := range ids {
		p, ok := pc.cache[pid]
		if !ok {
			p, err = pc.c.Person(pid)
			if err != nil {
				return nil, err
			}
			pc.cache[pid] = p
		}
		toRet[n] = p
	}
	return toRet, nil
}

func (c *Conn) getIndex(count, query *sql.Stmt, queryStr string, perPage, page uint) (uint, []Row, error) {
	var num uint
	err := count.QueryRow(queryStr).Scan(&num)
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
	cache := PersonCache{
		c,
		make(map[uint]*Person),
	}

	for _, id := range ids {
		p, err := cache.Get(id)
		if err != nil {
			return 0, nil, err
		}

		parents, err := cache.Get(p[0].Parents...)
		if err != nil {
			return 0, nil, err
		}
		siblings, err := cache.Get(p[0].Siblings...)
		if err != nil {
			return 0, nil, err
		}
		spouses, err := cache.Get(p[0].Spouses...)
		if err != nil {
			return 0, nil, err
		}
		children, err := cache.Get(p[0].Children...)
		if err != nil {
			return 0, nil, err
		}

		r := Row{
			Person:   p[0],
			Parents:  parents,
			Siblings: siblings,
			Spouses:  spouses,
			Children: children,
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
				countIndex, _ := db.Prepare("SELECT COUNT(1) FROM [People] WHERE [lname] LIKE ? || '%';")
				index, _ := db.Prepare("SELECT [id] FROM [People] WHERE [lname] LIKE ? || '%' ORDER BY [fname] ASC, [lname] ASC LIMIT ? OFFSET ?;")
				countSearch, _ := db.Prepare("SELECT COUNT(1) FROM [People] WHERE [fname] || ' ' || [lname] LIKE '%' ||  ? || '%';")
				search, _ := db.Prepare("SELECT [id] FROM [People] WHERE [fname] || ' ' || [lname] LIKE '%' || ? || '%' ORDER BY [fname] ASC, [lname] ASC LIMIT ? OFFSET ?;")
				person, _ := db.Prepare("SELECT [id], [fname], [lname], [sex], CASE [deathdate] WHEN '' THEN 0 ELSE 1 END AS [isdead], [famc], [fams] FROM [People] WHERE [id] = ?;")
				family, _ := db.Prepare("SELECT [husband], [wife], [children] FROM [Fams] WHERE [id] = ?;")
				l := &Conn{
					db:          db,
					indexCount:  countIndex,
					index:       index,
					searchCount: countSearch,
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
	l.indexCount.Close()
	l.index.Close()
	l.searchCount.Close()
	l.search.Close()
	l.person.Close()
	l.family.Close()
	l.db.Close()
}
