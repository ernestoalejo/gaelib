package app

import (
	"html/template"
	"io"
	"path/filepath"
	"sync"

	"appengine"

	"github.com/ernestokarim/gaelib/v1/errors"
)

var (
	templatesMutex = &sync.Mutex{}
	templatesCache = map[string]*template.Template{}
)

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

		t, err = t.ParseFiles(c.Names...)
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
