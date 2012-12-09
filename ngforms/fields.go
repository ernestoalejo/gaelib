package ngforms

import (
	"fmt"
	"strings"
)

// Allowed validators for this kind of input
var allowedValidators = map[string]map[string]bool{
	"text": map[string]bool{
		"required":  true,
		"minlength": true,
		"maxlength": true,
		"pattern":   true,
	},
	"email": map[string]bool{
		"required": true,
		"email":    true,
	},
	"password": map[string]bool{
		"required":  true,
		"minlength": true,
	},
}

// Some validators are always required by the input type
// (they're checked in the client anyway).
var neededValidators = map[string][]string{
	"text":     []string{},
	"email":    []string{"email"},
	"password": []string{},
}

// ==================================================================

type Control struct {
	Id, Name    string
	Help        string
	Validations []*Validator

	errors []string
}

func (c *Control) Build() string {
	var errs string
	for _, err := range c.errors {
		errs += "f." + c.Id + ".$error." + err + " || "
	}
	if len(errs) > 0 {
		errs = errs[:len(errs)-4]
	}

	return fmt.Sprintf(`
		<div class="control-group" ng-class="val && (%s) && 'error'">
			<label class="control-label" for="%s">%s</label>
			<div class="controls">%%s%%s</div>
		</div>
	`, errs, c.Id, c.Name)
}

func (f *Control) Validate(value string) bool {
	for _, v := range f.Validations {
		if !v.Func(value) {
			return false
		}
	}
	return true
}

// ==================================================================

type InputField struct {
	Type               string
	Control            *Control
	Class              []string
	Disabled, ReadOnly bool
	PlaceHolder        string
}

func (f *InputField) Build() string {
	// Initial arguments
	attrs := map[string]string{
		"type":        f.Type,
		"id":          f.Control.Id,
		"name":        f.Control.Id,
		"placeholder": f.PlaceHolder,
		"class":       strings.Join(f.Class, " "),
		"ng-model":    "data." + f.Control.Id,
	}

	// Flags
	if f.Disabled {
		attrs["disabled"] = "disabled"
	}
	if f.ReadOnly {
		attrs["readonly"] = "readonly"
	}

	// Validation attrs
	errors := fmt.Sprintf(`<p class="help-block error" ng-show="val">`)
	for _, v := range f.Control.Validations {
		// Check if it's an accepted validator
		allowed, ok := allowedValidators[f.Type]
		if !ok {
			panic("input type not supported")
		}
		if _, ok := allowed[v.Error]; !ok {
			panic("validator not allowed in " + f.Control.Id + ": " + v.Error)
		}

		// Add the attributes and errors
		for k, v := range v.Attrs {
			attrs[k] = v
		}
		errors += fmt.Sprintf(`<span ng-show="f.%s.$error.%s">%s</span>`, f.Control.Id,
			v.Error, v.Message)
		f.Control.errors = append(f.Control.errors, v.Error)
	}
	errors += "</p>"

	// Check that the needed validators for this input type are present
	for _, needed := range neededValidators[f.Type] {
		found := false
		for _, v := range f.Control.Validations {
			if v.Error == needed {
				found = true
				break
			}
		}

		if !found {
			panic("required validator for field " + f.Control.Id + ": " + needed)
		}
	}

	// Build the control HTML
	ctrl := "<input"
	for k, v := range attrs {
		ctrl += fmt.Sprintf(" %s=\"%s\"", k, v)
	}
	ctrl += ">"

	return fmt.Sprintf(f.Control.Build(), ctrl, errors)
}

func (f *InputField) Validate(value string) bool {
	return f.Control.Validate(value)
}

// ==================================================================

type SubmitField struct {
	Label                  string
	CancelUrl, CancelLabel string
}

func (f *SubmitField) Build() string {
	// Build the cancel button if present
	cancel := ""
	if f.CancelLabel != "" && f.CancelUrl != "" {
		cancel = fmt.Sprintf(`&nbsp;&nbsp;&nbsp;<a href="%s" class="btn">%s</a>`,
			f.CancelUrl, f.CancelLabel)
	}

	// Build the control
	return fmt.Sprintf(`
		<div class="form-actions">
			<button ng-click="val = true;" class="btn btn-primary">%s</button>
			%s
		</div>
	`, f.Label, cancel)
}

func (f *SubmitField) Validate(value string) bool {
	return true
}

/*

type SelectField struct {
	Control        *Control
	Class          []string
	Labels, Values []string
}

func (f *SelectField) Build() string {
	// The select tag attributes
	attrs := map[string]string{
		"id":   f.Control.Id,
		"name": f.Control.Id,
	}

	// The CSS classes
	if f.Class != nil {
		attrs["class"] = strings.Join(f.Class, " ")
	}

	ctrl := "<select"
	for k, v := range attrs {
		ctrl += fmt.Sprintf(" %s=\"%s\"", k, v)
	}
	ctrl += ">"

	// Assert the same length precondition, because the error is not
	// very descriptive
	if len(f.Labels) != len(f.Values) {
		panic("labels and values should have the same size")
	}

	for i, label := range f.Labels {
		// Option tag attributes
		attrs := map[string]string{}

		if f.Values[i] == "" {
			// Hide the option if it's the default blank one
			attrs["style"] = "display: none;"
		} else {
			// If it's the currently select one, select it again
			if f.Control.Value == f.Values[i] {
				attrs["selected"] = "selected"
			}

			// Set the value
			attrs["value"] = f.Values[i]
		}

		// Build the HTML of the option tag
		ctrl += "<option"
		for k, v := range attrs {
			ctrl += fmt.Sprintf(" %s=\"%s\"", k, v)
		}
		ctrl += ">" + label + "</option>"
	}

	// Finish the control build
	ctrl += "</select>"

	return fmt.Sprintf(f.Control.Build(), ctrl)
}

// --------------------------------------------------------

type TextAreaField struct {
	Control     *Control
	Class       []string
	Rows        int
	PlaceHolder string
}

func (f *TextAreaField) Build() string {
	// Tag attributes
	attrs := map[string]string{
		"rows":        strconv.FormatInt(int64(f.Rows), 10),
		"id":          f.Control.Id,
		"name":        f.Control.Id,
		"placeholder": f.PlaceHolder,
	}

	// The CSS classes
	if f.Class != nil {
		attrs["class"] = strings.Join(f.Class, " ")
	}

	// Build the control HTML
	ctrl := "<textarea"
	for k, v := range attrs {
		ctrl += fmt.Sprintf(" %s=\"%s\"", k, v)
	}
	ctrl += ">" + template.HTMLEscapeString(f.Control.Value) + "</textarea>"

	return fmt.Sprintf(f.Control.Build(), ctrl)
}

// --------------------------------------------------------

type HiddenField struct {
	Name, Value string
}

func (f *HiddenField) Build() string {
	return fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, f.Name,
		template.HTMLEscapeString(f.Value))
}
*/
