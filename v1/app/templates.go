package app

import (
	"html"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"appengine"

	"github.com/ernestokarim/gaelib/v1/errors"
)

var (
	templatesMutex = &sync.Mutex{}
	templatesCache = map[string]*template.Template{}
	templatesFuncs = template.FuncMap{}
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

	// Convert any string into an HTML snnipet
	// Take care of not introducing a security error
	AddTemplateFunc("bhtml", func(s string) template.HTML {
		return template.HTML(s)
	})

	// Converts new lines to their equivalent in HTML
	AddTemplateFunc("nl2br", func(s string) template.HTML {
		s = html.EscapeString(s)
		s = strings.Replace(s, "\n", "<br>", -1)
		return template.HTML(s)
	})
}

func AddTemplateFunc(name string, f interface{}) {
	templatesFuncs[name] = f
}

type TemplateConfig struct {
	LeftDelim, RightDelim string
	Names                 []string
	W                     io.Writer
	Data                  interface{}
	Dir                   string
}

func Template(w io.Writer, names []string, data interface{}) error {
	return ExecTemplate(&TemplateConfig{
		LeftDelim:  "{{",
		RightDelim: "}}",
		Names:      names,
		W:          w,
		Data:       data,
		Dir:        "templates",
	})
}

func TemplateDelims(w io.Writer, names []string, data interface{}, leftDelim, rightDelim string) error {
	return ExecTemplate(&TemplateConfig{
		LeftDelim:  leftDelim,
		RightDelim: rightDelim,
		Names:      names,
		W:          w,
		Data:       data,
		Dir:        "templates",
	})
}

func ExecTemplate(c *TemplateConfig) error {
	templatesMutex.Lock()
	defer templatesMutex.Unlock()

	cname := ""
	for i, name := range c.Names {
		c.Names[i] = filepath.Join(c.Dir, name+".html")
		cname += name
	}

	t, ok := templatesCache[cname]
	if !ok || appengine.IsDevAppServer() {
		var err error
		t = template.New(cname).Delims(c.LeftDelim, c.RightDelim)

		t, err = t.Funcs(templatesFuncs).ParseFiles(c.Names...)
		if err != nil {
			return errors.New(err)
		}
		templatesCache[cname] = t
	}

	if err := t.ExecuteTemplate(c.W, "base", c.Data); err != nil {
		return errors.New(err)
	}

	return nil
}
