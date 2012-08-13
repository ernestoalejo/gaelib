package app

import (
	"fmt"
	"html/template"
	"io"
	"time"

	"appengine"
)

const DATETIME_FORMAT = "02/01/2006 15:04:05"

var (
	templatesCache = map[string]*template.Template{}
	templatesFuncs = template.FuncMap{
		"equals":   func(a, b interface{}) bool { return a == b },
		"last":     func(max, i int) bool { return i == max-1 },
		"datetime": func(t time.Time) string { return t.Format(DATETIME_FORMAT) },
		"bhtml":    func(s string) template.HTML { return template.HTML(s) },
	}
)

func RawExecuteTemplate(w io.Writer, names []string, data interface{}) error {
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
		t, err = template.New(cname).Funcs(templatesFuncs).ParseFiles(names...)
		if err != nil {
			return fmt.Errorf("cannot parse the template: %s", err)
		}
		templatesCache[cname] = t
	}

	// Execute them
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("cannot execute the template: %s", err)
	}

	return nil
}
