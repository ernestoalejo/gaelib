package app

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"sync"

	"appengine"
)

var (
	templatesMutex = &sync.Mutex{}
	templatesCache = map[string]*template.Template{}
)

type TemplateConfig struct {
	Names                 []string
	W                     io.Writer
	Data                  interface{}
	Dir                   string
}

func Template(w io.Writer, names []string, data interface{}) error {
	return ExecTemplate(&TemplateConfig{
		Names:      names,
		W:          w,
		Data:       data,
		Dir:        "templates",
	})
}

func ExecTemplate(c *TemplateConfig) error {
	cname := ""
	for i, name := range c.Names {
		c.Names[i] = filepath.Join(c.Dir, name+".html")
		cname += name
	}

	templatesMutex.Lock()
	defer templatesMutex.Unlock()

	t, ok := templatesCache[cname]
	if !ok || appengine.IsDevAppServer() {
		var err error
		t, err = template.New(cname).ParseFiles(c.Names...)
		if err != nil {
			return fmt.Errorf("templates parsing failed: %s", err)
		}
		templatesCache[cname] = t
	}

	if err := t.ExecuteTemplate(c.W, "base", c.Data); err != nil {
		return fmt.Errorf("exec templates failed: %s", err)
	}

	return nil
}
