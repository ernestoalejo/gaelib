package ngforms

import (
	"fmt"
	"strings"
)

func BuildControl(form Form, id, name, help string) (map[string]string, string) {
	var errs, messages string
	attrs := map[string]string{}

	validationMap := form.Validations()
	validations, ok := validationMap[id]
	if !ok {
		panic("control without validations: " + id)
	}

	d := getFormData(form)
	messages = fmt.Sprintf(`
        <p class="help-block error" ng-show="%s.val && %s.%s.$invalid">
	`, d.Name, d.Name, id)

	for _, val := range validations {
		update(attrs, val.Attrs)
		errs += fmt.Sprintf("%s.%s.$error.%s || ", d.Name, id, val.Error)
		messages += fmt.Sprintf(`
			<span ng-show="%s.%s.$error.%s">%s</span>
		`, d.Name, id, val.Error, val.Message)
	}

	messages += `</p>`
	errs = errs[:len(errs)-4]

	if name == "" {
		return attrs, fmt.Sprintf(`
	        <div class="control-group" ng-class="%s.val && (%s) && 'error'">
	          %%s%s
	        </div>
		`, d.Name, errs, messages)
	}

	return attrs, fmt.Sprintf(`
      <div class="control-group" ng-class="%s.val && (%s) && 'error'">
        <label class="control-label" for="%s">%s</label>
        <div class="controls">%%s%s</div>
      </div>
	`, d.Name, errs, id, name, messages)
}

// ==================================================================

type InputField struct {
	Id, Name    string
	Help        string
	Type        string
	Class       []string
	PlaceHolder string

	Attrs map[string]string
}

func (f *InputField) Build(form Form) string {
	if f.Type == "" {
		panic("input type should not be empty: " + f.Id)
	}

	attrs := map[string]string{
		"type":        f.Type,
		"id":          f.Id,
		"name":        f.Id,
		"placeholder": f.PlaceHolder,
		"class":       strings.Join(f.Class, " "),
		"ng-model":    "data." + f.Id,
	}
	update(attrs, f.Attrs)

	controlAttrs, control := BuildControl(form, f.Id, f.Name, f.Help)
	update(attrs, controlAttrs)

	ctrl := "<input"
	for k, v := range attrs {
		ctrl += fmt.Sprintf(` %s="%s"`, k, v)
	}
	ctrl += ">"

	return fmt.Sprintf(control, ctrl)
}

// ==================================================================

type SubmitField struct {
	Label       string
	CancelUrl   string
	CancelLabel string
}

func (f *SubmitField) Build(form Form) string {
	cancel := ""
	if f.CancelLabel != "" && f.CancelUrl != "" {
		cancel = fmt.Sprintf(`&nbsp;&nbsp;&nbsp;<a href="%s" class="btn">%s</a>`,
			f.CancelUrl, f.CancelLabel)
	}

	d := getFormData(form)
	return fmt.Sprintf(`
		<div class="form-actions">
			<button ng-click="%s(); %s.val = true;" class="btn btn-primary"
				ng-disabled="%s.val && !%s.$valid">%s</button>
			%s
		</div>
	`, d.TrySubmit, d.Name, d.Name, d.Name, f.Label, cancel)
}

// ==================================================================

type TextAreaField struct {
	Id, Name    string
	Help        string
	Class       []string
	Rows        int
	PlaceHolder string
}

func (f *TextAreaField) Build(form Form) string {
	attrs := map[string]string{
		"id":          f.Id,
		"name":        f.Id,
		"placeholder": f.PlaceHolder,
		"class":       strings.Join(f.Class, " "),
		"ng-model":    "data." + f.Id,
		"rows":        fmt.Sprintf("%d", f.Rows),
	}

	controlAttrs, control := BuildControl(form, f.Id, f.Name, f.Help)
	update(attrs, controlAttrs)

	ctrl := "<textarea"
	for k, v := range attrs {
		ctrl += fmt.Sprintf(` %s="%s"`, k, v)
	}
	ctrl += "></textarea>"

	return fmt.Sprintf(control, ctrl)
}

/*
// ==================================================================

type SelectField struct {
	Control        *Control
	Class          []string
	Labels, Values []string
}

func (f *SelectField) Build() string {
	// The select tag attributes
	attrs := map[string]string{
		"id":       f.Control.Id,
		"name":     f.Control.Id,
		"ng-model": "data." + f.Control.Id,
	}

	// The CSS classes
	if f.Class != nil {
		attrs["class"] = strings.Join(f.Class, " ")
	}

	// Add the validators
	errors := fmt.Sprintf(`<p class="help-block error" ng-show="val && f.%s.$invalid">`,
		f.Control.Id)
	for _, v := range f.Control.Validations {
		// Fail early if it's not a valid one
		if v.Error != "required" && v.Error != "select" {
			panic("validator not allowed in select " + f.Control.Id + ": " + v.Error)
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

	// Build the tag
	ctrl := "<select"
	for k, v := range attrs {
		ctrl += fmt.Sprintf(" %s=\"%s\"", k, v)
	}
	ctrl += ">"

	// Assert the same length precondition, because the error is not
	// very descriptive. Then add the option tags to the select field.
	if len(f.Labels) != len(f.Values) {
		panic("labels and values should have the same size")
	}
	for i, label := range f.Labels {
		ctrl += fmt.Sprintf(`<option value="%s">%s</option>`, f.Values[i], label)
	}

	// Finish the control build
	ctrl += "</select>"

	return fmt.Sprintf(f.Control.Build(), ctrl, errors)
}

func (f *SelectField) Validate(value string) bool {
	return f.Control.Validate(value)
}*/

// ==================================================================

// Update the contents of m with the s items
func update(m map[string]string, s map[string]string) {
	for k, v := range s {
		m[k] = v
	}
}
