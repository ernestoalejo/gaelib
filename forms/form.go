package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ernestokarim/gaelib/app"
)

type Field interface {
	Build() string
}

type Form struct {
	Name           string
	Method, Action string
	FieldNames     []string
	Fields         map[string]Field

	// True if we want to show a message at the top of the form
	// each time a validation error occurs
	ShowError bool
}

func New(action string) *Form {
	return &Form{
		Method:     "POST",
		Action:     action,
		FieldNames: make([]string, 0),
		Fields:     make(map[string]Field),
	}
}

func (f *Form) AddField(name string, field Field) {
	f.FieldNames = append(f.FieldNames, name)
	f.Fields[name] = field
}

func (f *Form) Build() string {
	// Set the error class if needed
	out := ""
	withError := false
	for _, name := range f.FieldNames {
		out += f.Fields[name].Build()

		ctrl := f.GetControl(name)
		if ctrl != nil && ctrl.Error != "" {
			withError = true
		}
	}

	if withError && f.ShowError {
		out = `
			<div class="alert alert-error">
				Hay errores en el formulario, revísalos y guarda de nuevo
			</div><br>
		` + out
	}

	// Set the form legend if needed
	legend := ""
	if f.Name != "" {
		legend = "<legend>" + f.Name + "</legend>"
	}

	return fmt.Sprintf(`
		<form action="%s" method="%s" class="form-horizontal">
			<fieldset>%s%s</fieldset>
		</form>
	`, f.Action, f.Method, legend, out)
}

func (f *Form) Validate(r *app.Request, data interface{}) (bool, error) {
	if err := r.Req.ParseForm(); err != nil {
		return false, app.Error(err)
	}

	failed := false
	for _, name := range f.FieldNames {
		field := f.Fields[name]
		if control := getControl(field); control != nil {
			// Extract the control value and assign it
			value := normalizeValue(control.Id, r.Req.Form)
			control.Value = value

			// Run each validation for this field
			for _, val := range control.Validations {
				if err := val.Func(value); err != "" {
					failed = true

					control.Error = err
					if control.ResetValue {
						control.Value = ""
					}

					break
				}
			}
		}
	}

	if !failed {
		if err := r.LoadData(data); err != nil {
			return true, err
		}
	}

	return !failed, nil
}

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
