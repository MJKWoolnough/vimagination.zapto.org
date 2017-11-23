package main

import (
	"net/http"

	"github.com/MJKWoolnough/form"
)

type CalcVars struct {
	First  uint
	Chosen uint
}

func (c *CalcVars) ParserList() form.ParserList {
	return form.ParserList{
		"first":  form.Uint{&c.First},
		"chosen": form.Uint{&c.Chosen},
	}
}

type CalcTemplateVars struct {
	Found         bool
	Links         Links
	First, Second *Person
}

func (l *List) Calculator(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var cv CalcVars
	form.Parse(&cv, r.Form)
	first := GedcomData.People[0]
	second := GedcomData.People[0]
	if cv.First > 0 {
		var ok bool
		first, ok = GedcomData.People[cv.First]
		if !ok {
			cv.First = 0
		}
	}
	if cv.Chosen == cv.First {
		cv.Chosen = 0
	}
	if cv.Chosen > 0 {
		var ok bool
		second, ok = GedcomData.People[cv.Chosen]
		if !ok {
			cv.Chosen = 0
		}
	}
	if cv.First == 0 || cv.Chosen == 0 {
		l.serveList(w, r, true, cv.Chosen+cv.First)
		return
	}
	var reverse bool
	if first.ID > second.ID {
		reverse = true
		first, second = second, first
	}
	key := [2]uint{first.ID, second.ID}
	GedcomData.RWMutex.RLock()
	links, ok := GedcomData.RelationshipCache[key]
	GedcomData.RWMutex.RUnlock()
	if !ok {
		links = Calculate(first, second)
		GedcomData.RWMutex.Lock()
		GedcomData.RelationshipCache[key] = links
		GedcomData.RWMutex.Unlock()
	}
	if reverse {
		first, second = second, first
		links.First, links.Second = links.Second, links.First
	}
	ctv := CalcTemplateVars{
		Found:  links.Common != nil,
		Links:  links,
		First:  first,
		Second: second,
	}
	w.Header().Set("Content-Type", "text/html")
	l.RelationTemplate.Execute(w, ctv)
}
