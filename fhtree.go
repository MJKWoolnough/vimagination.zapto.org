package main

import (
	"html/template"
	"net/http"

	"github.com/MJKWoolnough/form"
)

type Tree struct {
	HTMLTemplate, JSTemplate *template.Template
}

type TreeVars struct {
	ID uint
}

func (t *TreeVars) ParserList() form.ParserList {
	return form.ParserList{
		"id": form.Uint{&t.ID},
	}
}

func (t *Tree) HTML(w http.ResponseWriter, r *http.Request) {
	var tv TreeVars
	r.ParseForm()
	form.Parse(&tv, r.Form)
	person := GedcomData.People[tv.ID]
	if person == nil {
		return
	}
	t.HTMLTemplate.Execute(w, person)
}

func (t *Tree) JS(w http.ResponseWriter, r *http.Request) {
	var tv TreeVars
	r.ParseForm()
	form.Parse(&tv, r.Form)
	w.Header().Set("Content-Type", "text/javascript")
	person := GedcomData.People[tv.ID]
	if person == nil {
		return
	}
	t.JSTemplate.Execute(w, person)
}
