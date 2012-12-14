package ngforms

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

func (f *Form) Validate(r *app.Request, data interface{}) error {
	// Read the whole body in a buffer
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Req.Body); err != nil {
		return app.Error(err)
	}

	// Create a copy, and store the first read one as the
	// request body
	r.Req.Body = ioutil.NopCloser(&buf)
	nbuf := ioutil.NopCloser(bytes.NewBuffer(buf.Bytes()))

	// Read the map of values from the first copy
	m := make(map[string]interface{})
	if err := r.LoadJsonData(&m); err != nil {
		return err
	}

	// Save the second copy as the new request body
	r.Req.Body = nbuf

	for _, name := range f.FieldNames {
		// Extract the field and the control
		field := f.Fields[name]
		control := getControl(field)
		if control == nil {
			continue
		}

		// Extract the value
		value := normalizeValue(control.Id, m)
		if !field.Validate(value) {
			return app.Errorf("validation failed for field %s: %s", control.Id, value)
		}

		// Save the value if it's correct
		control.value = value
	}

	// If the validation finished succesfully; load the form
	// data into the struct
	if err := r.LoadJsonData(data); err != nil {
		return err
	}

	return nil
}

func normalizeValue(id string, m map[string]interface{}) string {
	// Extract the value
	value, ok := m[id]
	if ok {
		str, ok := value.(string)
		if ok {
			// Trim the value before returning it
			return strings.TrimSpace(str)
		}
	}

	// No value found
	return ""
}

func (f *Form) GetControl(name string) *Control {
	// Obtain the field from the list
	field, ok := f.Fields[name]
	if !ok {
		return nil
	}

	return getControl(field)
}

func getControl(f Field) *Control {
	input, ok := f.(*InputField)
	if ok {
		return input.Control
	}

	sel, ok := f.(*SelectField)
	if ok {
		return sel.Control
	}

	ta, ok := f.(*TextAreaField)
	if ok {
		return ta.Control
	}

	// Not a control
	return nil
}
