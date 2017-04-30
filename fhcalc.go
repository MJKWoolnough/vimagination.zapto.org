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
	Links         *Links
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

	key := [2]uint{first.ID, second.ID}
	GedcomData.RWMutex.RLock()
	links, ok := GedcomData.RelationshipCache[key]
	GedcomData.RWMutex.RUnlock()
	if !ok {
		links = Calculate(first, second)
		GedcomData.RWMutex.Lock()
		rev := links.Reverse()
		GedcomData.RelationshipCache[key] = links
		GedcomData.RelationshipCache[[2]uint{second.ID, first.ID}] = rev
		GedcomData.RWMutex.Unlock()
	}
	ctv := CalcTemplateVars{
		Found:  links != nil,
		Links:  links,
		First:  first,
		Second: second,
	}
	l.RelationTemplate.Execute(w, ctv)
}
