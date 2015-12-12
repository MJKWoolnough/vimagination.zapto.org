package main

import (
	"database/sql"
	"net/http"
	"runtime"
	"sync"

	_ "github.com/mxk/go-sqlite/sqlite3"
)

type List struct {
	Pool sync.Pool
}

type ListConn struct {
	db                                   *sql.DB
	listCount, list, searchCount, search *sql.Stmt
}

func (l *ListConn) List(char string, perPage, page uint) (uint, []ListRow, error) {
	return l.exec(l.listCount, l.list, char, perPage, page)
}

func (l *ListConn) Search(query string, perPage, page uint) (uint, []ListRow, error) {
	return l.exec(l.searchCount, l.search, char, perPage, page)
}

func (l *ListConn) exec(count, query *sql.Stmt, queryStr string, perPage, page uint) {
	var num uint
	err := count.QueryRow(char).Scan(&num)
	if err != nil {
		return 0, nil, err
	}
	rows, err := query.Query(queryStr, perPage, page*perPage)
	if err != nil {
		return 0, nil, err
	}
}

type ListPerson struct {
	ID                  int
	FirstName, LastName string
	Sex                 string
	Dead                bool
	Famc                int
	Fams                string
}

type ListRow struct {
	*Person
	Parents, Siblings, Children []*ListPerson
}

func NewList(databaseURL string) *List {
	return &List{
		Pool: sync.Pool{
			New: func() interface{} {
				db, _ := sql.Open("sqlite3", databaseURL)
				countList, _ := db.Prepare("SELECT COUNT(1) FROM [People] WHERE [lname] LIKE CONCAT(?, '%');")
				list, _ := db.Prepare("SELECT [id], [fname], [lname], [sex], IF([deathdate] = '', 0, 1) AS [isdead], [famc], [fams] FROM [People] WHERE [lname] LIKE CONCAT(?, '%') ORDER BY [lname] ASC, [fname] ASC LIMIT ? OFFSET ?;")
				countSearch, _ := db.Prepare("SELECT COUNT(1) FROM [People] WHERE CONCAT([fname], ' ', [lname]) LIKE CONCAT('%', ?, '%');")
				search, _ := db.Prepare("SELECT [id], [fname], [lname], [sex], IF([deathdate] = '', 0, 1) AS [isdead], [famc], [fams] FROM [People] WHERE CONCAT([fname], ' ', [lname]) LIKE CONCAT('%', ?, '%') ORDER BY [lname] ASC, [fname] ASC LIMIT ? OFFSET ?;")
				l := &ListConn{
					db:          db,
					countList:   countList,
					list:        list,
					countSearch: countSearch,
					search:      search,
				}
				runtime.SetFinalizer(l, closeListConn)
				return l
			},
		},
	}
}

func closeListConn(l *ListConn) {
	l.listCount.Close()
	l.list.Close()
	l.searchCount.Close()
	l.search.Close()
	l.db.Close()
}

func (l *List) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lc := l.Pool.Get().(*ListConn)
	defer l.Pool.Put(lc)
}
