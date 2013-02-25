package ngforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ernestokarim/gaelib/errors"
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
	Data() *FormData
	Fields() FieldList
	Validations() ValidationMap
}

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
