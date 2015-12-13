package main

import (
	"html/template"
	"net/http"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/pagination"
)

type Letter struct {
	data *string
}

func (l Letter) Parse(d []string) error {
	if len(d[0]) == 1 {
		for i := 'A'; i <= 'Z'; i++ {
			if d[0][0] == i {
				*l.data = d[0]
				break
			}
		}
	}
	return nil
}

type IndexVars struct {
	Page   uint
	Letter string
	Query  string
}

func (i *IndexVars) ParserList() form.ParserList {
	return form.ParserList{
		"page":   form.Uint{&i.Page},
		"letter": Letter{&i.Letter},
		"query":  form.String{&i.Query},
	}
}

type List struct {
	Template *template.Template
	Conn     *Conn
}

func (l *List) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const perPage = 10
	lc := l.Pool.Get().(*ListConn)
	defer l.Pool.Put(lc)
	var index IndexVars
	form.Parse(&l, r.Form)
	var (
		rows       []Row
		totalPages uint
	)
	if index.Query != "" {
		index.Letter = ""
		var err error
		rows, err = l.Conn.Search(index.Query, perPage, index.Page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else if index.Letter != "" {
		var err error
		rows, err = l.Conn.Index(index.Letter, perPage, index.Page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	tv := struct {
		IndexVars
		Rows       []Row
		Pagination template.HTML
	}{
		index,
		rows,
		template.HTML(pagination.New().Get(index.Page, totalPages).HTML()),
	}

	l.Template.Execute(w, tv)
}
