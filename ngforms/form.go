package ngforms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ernestokarim/gaelib/app"
)

type Field interface {
	Build() string
	Validate(value string) bool
}

type Form struct {
	Name       string
	FieldNames []string
	Fields     map[string]Field
}

func New() *Form {
	return &Form{
		FieldNames: make([]string, 0),
		Fields:     make(map[string]Field),
	}
}

func (f *Form) AddField(name string, field Field) {
	f.FieldNames = append(f.FieldNames, name)
	f.Fields[name] = field
}

func (f *Form) Build() string {
	// Build each field
	out := ""
	for _, name := range f.FieldNames {
		out += f.Fields[name].Build()
	}

	// Set the form legend if present
	legend := ""
	if f.Name != "" {
		legend = "<legend>" + f.Name + "</legend>"
	}

	// Build the form output
	return fmt.Sprintf(`
		<form class="form-horizontal" name="f" novalidate ng-init="val = false;"
			ng-submit="f.$valid && submit()">
			<fieldset>%s%s</fieldset>
		</form>
	`, legend, out)
}

func (f *Form) Validate(r *app.Request, data interface{}) (bool, error) {
	if err := r.Req.ParseForm(); err != nil {
		return false, app.Error(err)
	}

	for _, name := range f.FieldNames {
		// Extract the field and the control
		field := f.Fields[name]
		control, ok := field.(*Control)
		if !ok {
			continue
		}

		// Extract the the value
		value := normalizeValue(control.Id, r.Req.Form)
		if !field.Validate(value) {
			return false, nil
		}
	}

	if err := r.LoadData(data); err != nil {
		return false, err
	}

	return true, nil
}

func normalizeValue(id string, v url.Values) string {
	// Extract the value
	values, ok := v[id]
	if ok {
		// Trim the value
		v[id] = []string{strings.TrimSpace(values[0])}

		// Return the value
		return v[id][0]
	}

	// No value found
	return ""
}

/*
func (f *Form) GetControl(name string) *Control {
	field, ok := f.Fields[name]
	if !ok {
		return nil
	}

	return getControl(field)
}

func getControl(f Field) *Control {
	// Control for inputs
	input, ok := f.(*InputField)
	if ok {
		return input.Control
	}

	// Control for selects
	sel, ok := f.(*SelectField)
	if ok {
		return sel.Control
	}

	// Control for textarea
	textarea, ok := f.(*TextAreaField)
	if ok {
		return textarea.Control
	}

	// Not a control
	return nil
}*/
