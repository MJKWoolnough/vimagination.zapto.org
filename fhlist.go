package main

import (
	"html/template"
	"net/http"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/pagination"
)

type Letter struct {
	data *byte
}

func (l Letter) Parse(d []string) error {
	if len(d[0]) == 1 {
		for i := byte('A'); i <= 'Z'; i++ {
			if d[0][0] == i {
				*l.data = i - 'A' + 1
				break
			}
		}
	}
	return nil
}

type IndexVars struct {
	Page   uint
	Letter byte
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
}

func (l *List) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const perPage = 20
	var iv IndexVars
	r.ParseForm()
	form.Parse(&iv, r.Form)
	var (
		urlBase        string
		index          Index
		paginationHTML template.HTML
	)
	if iv.Query != "" {
		iv.Letter = 0
		// store/restore with session storage???
		index = GedcomData.Search(iv.Query)
		urlBase = "?query=" + template.HTMLEscapeString(iv.Query) + "&amp;page="
	} else if iv.Letter > 0 {
		index = GedcomData.Indexes[iv.Letter-1]
		urlBase = "?letter=" + string([]byte{iv.Letter + 'A' - 1}) + "&amp;page="
	}
	if iv.Page != 0 {
		iv.Page--
	}
	if iv.Page*perPage > uint(len(index)) {
		iv.Page = 0
	}
	if index != nil {
		numPages := uint(len(index)) / perPage
		if numPages > 0 && len(index)%perPage == 0 {
			numPages--
		}
		first := iv.Page * perPage
		last := (iv.Page + 1) * perPage
		if first > uint(len(index)) {
			first = 0
			last = 0
		}
		if last > uint(len(index)) {
			index = index[first:]
		} else {
			index = index[first:last]
		}
		paginationHTML = template.HTML(pagination.New().Get(iv.Page, numPages).HTML(urlBase))
	}
	tv := struct {
		IndexVars
		HasRows    bool
		Rows       Index
		Pagination template.HTML
	}{
		iv,
		index != nil,
		index,
		paginationHTML,
	}

	l.Template.Execute(w, tv)
}
