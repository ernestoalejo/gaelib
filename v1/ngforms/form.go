package ngforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ernestokarim/gaelib/v1/errors"
)

type Field interface {
	Build(form Form) string
}

type FieldList []Field
type ValidationMap map[string][]*Validator

type FormData struct {
	// Name of the controller of the form
	Name string

	// Function called when the form passed all the validations
	// and is sent. Without the () pair
	Submit string

	// Function called each time the user try to send the form
	// Without the () pair
	TrySubmit string

	// Name of the client side object that will be scoped
	// with the values of the form
	ObjName string
}

type Form interface {
	// General data about the form
	// Extending *BaseForm gives you an implementation of this
	// Override when needed
	Data() *FormData

	// List of items of the form
	Fields() FieldList
	Validations() ValidationMap

	// Store the value of each field for validations
	// Extending *BaseForm gives you an implementation of these
	Value(key string) string
	SetValue(key, value string)
}

// ==================================================================

type BaseForm struct {
	values map[string]string
}

func (f *BaseForm) Data() *FormData {
	return nil
}

func (f *BaseForm) Value(key string) string {
	return f.values[key]
}

func (f *BaseForm) SetValue(key, value string) {
	if f.values == nil {
		f.values = map[string]string{}
	}
	f.values[key] = value
}

func (f *BaseForm) Fields() FieldList {
	panic("should override this function")
}

func (f *BaseForm) Validations() ValidationMap {
	panic("should override this function")
}

// ==================================================================

// Build the form returning the generated HTML
func Build(f Form) string {
	results := []string{}
	for _, field := range f.Fields() {
		results = append(results, field.Build(f))
	}

	d := getFormData(f)
	return fmt.Sprintf(`
      <form class="form-horizontal" name="%s" novalidate ng-init="%s.val = false;"
        ng-submit="%s.$valid && %s()"><fieldset>%s</fieldset></form>
    `, d.Name, d.Name, d.Name, d.Submit, strings.Join(results, ""))
}

func getFormData(f Form) *FormData {
	d := f.Data()
	if d == nil {
		d = new(FormData)
	}
	if d.Name == "" {
		d.Name = "f"
	}
	if d.Submit == "" {
		d.Submit = "submit"
	}
	if d.TrySubmit == "" {
		d.TrySubmit = "trySubmit"
	}
	if d.ObjName == "" {
		d.ObjName = "data"
	}
	return d
}

// Validate the form.
// Returns a boolean indicating if the data was valid according
// to the validations defined on f. It returns an error too.
func Validate(r *http.Request, f Form) (bool, error) {
	// Copy the body to a buffer so we can use it twice
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return false, errors.New(err)
	}
	nbuf := ioutil.NopCloser(bytes.NewBuffer(buf.Bytes()))
	r.Body = nbuf

	m := make(map[string]interface{})
	if err := json.NewDecoder(buf).Decode(&m); err != nil {
		return false, errors.New(err)
	}

	fields := f.Fields()
	validations := f.Validations()
	for _, field := range fields {
		id := getId(field)
		if id == "" {
			continue
		}

		// Skip fields without validation constrainst
		if _, ok := validations[id]; !ok {
			continue
		}

		value := normalizeValue(id, m)
		for _, val := range validations[id] {
			if !val.Func(value) {
				return false, nil
			}
		}
		f.SetValue(id, value)
	}

	if err := json.NewDecoder(r.Body).Decode(f); err != nil {
		return false, errors.New(err)
	}

	return true, nil
}

// Extract the value or its str counterpart to validate it
func normalizeValue(id string, m map[string]interface{}) string {
	if v, ok := extractValue(id, m); ok {
		return v
	}
	if v, ok := extractValue("str"+id, m); ok {
		return v
	}

	return ""
}

// Extract a value from the body data, trimming it
func extractValue(id string, m map[string]interface{}) (string, bool) {
	value, ok := m[id]
	if ok {
		str, ok := value.(string)
		if ok {
			return strings.TrimSpace(str), true
		}
	}

	return "", false
}

func getId(f Field) string {
	input, ok := f.(*InputField)
	if ok {
		return input.Id
	}

	ta, ok := f.(*TextAreaField)
	if ok {
		return ta.Id
	}

	return ""
}
