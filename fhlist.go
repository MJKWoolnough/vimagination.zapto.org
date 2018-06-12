package main

import (
	"html/template"
	"net/http"
	"strconv"

	"vimagination.zapto.org/form"
	"vimagination.zapto.org/pagination"
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
	ListTemplate     *template.Template
	RelationTemplate *template.Template
}

func (l *List) List(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	l.serveList(w, r, false, 0)
}

type IndexTemplateVars struct {
	IndexVars
	HasRows    bool
	Rows       Index
	Pagination template.HTML
	Calc       bool
	First      uint
}

var letters = [26]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

func (IndexTemplateVars) Letters() [26]string {
	return letters
}

func (l *List) serveList(w http.ResponseWriter, r *http.Request, calc bool, first uint) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	const perPage = 20
	var iv IndexVars
	form.Parse(&iv, r.Form)
	var (
		index          Index
		paginationHTML template.HTML
	)
	urlBase := "?"
	if iv.Query != "" {
		iv.Letter = 0
		// store/restore with session storage???
		index = GedcomData.Search(iv.Query)
		urlBase += "query=" + template.HTMLEscapeString(iv.Query) + "&amp;"
	} else if iv.Letter > 0 {
		index = GedcomData.Indexes[iv.Letter-1]
		urlBase += "letter=" + string([]byte{iv.Letter + 'A' - 1}) + "&amp;"
	}
	if first > 0 {
		urlBase += "first=" + strconv.FormatUint(uint64(first), 10) + "&amp;"
	}
	urlBase += "&amp;page="
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
	tv := IndexTemplateVars{
		iv,
		index != nil,
		index,
		paginationHTML,
		calc,
		first,
	}

	l.ListTemplate.Execute(w, tv)
}
