package forms

import (
	"strings"
)

type Validator struct {
	Name string
	Args []interface{}
	Func func(string) string
}

func NotEmpty(message string) *Validator {
	return &Validator{
		Name: "not-empty",
		Func: func(v string) string {
			if v == "" {
				return message
			}

			return ""
		},
	}
}

func LargerThan(n int, message string) *Validator {
	return &Validator{
		Name: "larger-than",
		Args: []interface{}{n},
		Func: func(v string) string {
			if len(v) < n {
				return message
			}

			return ""
		},
	}
}

func Email(message string) *Validator {
	return &Validator{
		Name: "email",
		Func: func(v string) string {
			if !strings.Contains(v, "@") || !strings.Contains(v, ".") {
				return message
			}

			return ""
		},
	}
}

func EqualsTo(f *Form, id, message string) *Validator {
	return &Validator{
		Name: "equals-to",
		Args: []interface{}{id},
		Func: func(v string) string {
			input := f.GetControl(id)
			if input != nil {
				if input.Value != v {
					return message
				}

				return ""
			}

			panic("should not reach here: " + id)
		},
	}
}

func SelectValue(f *Form, id, message string) *Validator {
	return &Validator{
		Name: "select-value",
		Func: func(v string) string {
			for _, field := range f.Fields {
				sel, ok := field.(*SelectField)
				if ok && sel.Control.Id == id {
					for _, value := range sel.Values {
						if v == value {
							return ""
						}
					}
				}
			}

			return message

			panic("should not reach here: " + id)
		},
	}
}
