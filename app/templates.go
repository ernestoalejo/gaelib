package app

import (
	"html/template"
	"io"
	"strings"
	"time"

	"appengine"
)

var (
	templatesCache = map[string]*template.Template{}
	templatesFuncs = template.FuncMap{}
	leftDelim      = "{{"
	rightDelim     = "}}"
)

func init() {
	// Compare two items to see if they're equal
	AddTemplateFunc("equals", func(a, b interface{}) bool {
		return a == b
	})

	// Returns true if the current iteration is the last one of the loop
	AddTemplateFunc("last", func(max, i int) bool {
		return i == max-1
	})

	// Quick format for a date & time
	AddTemplateFunc("datetime", func(t time.Time) string {
		return t.Format("02/01/2006 15:04:05")
	})

	// Converst any string into an HTML snnipet
	AddTemplateFunc("bhtml", func(s string) template.HTML {
		return template.HTML(s)
	})

	// Converts new lines to their equivalent in HTML
	AddTemplateFunc("nl2br", func(s string) template.HTML {
		return template.HTML(strings.Replace(s, "\n", "<br>", -1))
	})
}

func AddTemplateFunc(name string, f interface{}) {
	templatesFuncs[name] = f
}

func SetDelims(left, right string) {
	leftDelim = left
	rightDelim = right
}

func Template(w io.Writer, names []string, data interface{}) error {
	// Build the key for this template
	cname := ""
	for i, name := range names {
		names[i] = "templates/" + name + ".html"
		cname += name
	}

	// Parse the templates
	t, ok := templatesCache[cname]
	if !ok || appengine.IsDevAppServer() {
		var err error
		t = template.New(cname).Delims(leftDelim, rightDelim)

		t, err = t.Funcs(templatesFuncs).ParseFiles(names...)
		if err != nil {
			return Error(err)
		}
		templatesCache[cname] = t
	}

	// Execute them
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		return Error(err)
	}

	return nil
}
