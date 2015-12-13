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
		for i := byte('A'); i <= 'Z'; i++ {
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
	Pool     *ConnPool
}

func (l *List) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const perPage = 10
	lc := l.Pool.Get().(*Conn)
	defer l.Pool.Put(lc)
	var index IndexVars
	r.ParseForm()
	form.Parse(&index, r.Form)
	if index.Page == 0 {
		index.Page = 1
	}
	var (
		rows    []Row
		num     uint
		urlBase string
	)
	if index.Query != "" {
		index.Letter = ""
		var err error
		num, rows, err = lc.Search(index.Query, perPage, index.Page-1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		urlBase = "?query=" + template.HTMLEscapeString(index.Query) + "&amp;page="
	} else if index.Letter != "" {
		var err error
		num, rows, err = lc.Index(index.Letter, perPage, index.Page-1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		urlBase = "?letter=" + index.Letter + "&amp;page="
	}
	totalPages := num / perPage
	tv := struct {
		IndexVars
		HasRows    bool
		Rows       []Row
		Pagination template.HTML
	}{
		index,
		len(rows) > 0,
		rows,
		template.HTML(pagination.New().Get(index.Page-1, totalPages).HTML(urlBase)),
	}

	l.Template.Execute(w, tv)
}
