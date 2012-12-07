package ngforms

import (
	"fmt"
	"strings"
)

type Control struct {
	Id, Name string
	Help     string

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
		<div class="control-group" ng-class="f.%s.$dirty && (%s) && 'error'">
			<label class="control-label" for="%s">%s</label>
			<div class="controls">%%s%%s</div>
		</div>
	`, c.Id, errs, c.Id, c.Name)
}

// --------------------------------------------------------

type TextField struct {
	Control            *Control
	Class              []string
	Disabled, ReadOnly bool
	PlaceHolder        string

	Required       bool
	MinLen, MaxLen int
	Pattern        string

	RequiredMsg          string
	MinLenMsg, MaxLenMsg string
	PatternMsg           string
}

func (f *TextField) Build() string {
	// Initial arguments
	attrs := map[string]string{
		"type":        "text",
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
	errors := fmt.Sprintf(`{{f.email.$error}}<p class="help-block error" ng-show="f.%s.$dirty">`,
		f.Control.Id)
	if f.MinLen > 0 {
		attrs["ng-minlength"] = fmt.Sprintf("%d", f.MinLen)
		errors += fmt.Sprintf(`<span ng-show="f.%s.$error.minlength">%s</span>`,
			f.Control.Id, f.MinLenMsg)
		f.Control.errors = append(f.Control.errors, "minlength")
	}
	if f.MaxLen > 0 {
		attrs["ng-maxlength"] = fmt.Sprintf("%d", f.MaxLen)
		errors += fmt.Sprintf(`<span ng-show="f.%s.$error.maxlength">%s</span>`,
			f.Control.Id, f.MaxLenMsg)
		f.Control.errors = append(f.Control.errors, "maxlength")
	}
	if f.Pattern != "" {
		attrs["ng-pattern"] = f.Pattern
		errors += fmt.Sprintf(`<span ng-show="f.%s.$error.pattern">%s</span>`,
			f.Control.Id, f.PatternMsg)
		f.Control.errors = append(f.Control.errors, "pattern")
	}
	if f.Required {
		attrs["required"] = "required"
		errors += fmt.Sprintf(`<span ng-show="f.%s.$error.required">%s</span>`,
			f.Control.Id, f.RequiredMsg)
		f.Control.errors = append(f.Control.errors, "required")
	}
	errors += "</p>"

	// Build the control HTML
	ctrl := "<input"
	for k, v := range attrs {
		ctrl += fmt.Sprintf(" %s=\"%s\"", k, v)
	}
	ctrl += ">"

	return fmt.Sprintf(f.Control.Build(), ctrl, errors)
}

func (f *TextField) Validate(value string) string {
	return ""
}

// --------------------------------------------------------
/*
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
			<button type="submit" class="btn btn-primary">%s</button>
			%s
		</div>
	`, f.Label, cancel)
}

// --------------------------------------------------------

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
