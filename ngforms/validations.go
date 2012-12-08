package ngforms

import (
	"fmt"
)

// A validator func it's one that receive a value as a param
// and returns true if the input it's correct.
type ValidatorFunc func(string) bool

type Validator struct {
	Attrs   map[string]string
	Message string
	Error   string
	Func    ValidatorFunc
}

func Required(msg string) *Validator {
	return &Validator{
		Attrs:   map[string]string{"required": ""},
		Message: msg,
		Error:   "required",
		Func:    func(v string) bool { return v != "" },
	}
}

func LargerThan(value int, msg string) *Validator {
	return &Validator{
		Attrs:   map[string]string{"ng-minlength": fmt.Sprintf("%d", value)},
		Message: msg,
		Error:   "minlength",
		Func:    func(v string) bool { return len(v) > value },
	}
}

func ShorterThan(value int, msg string) *Validator {
	return &Validator{
		Attrs:   map[string]string{"ng-maxlength": fmt.Sprintf("%d", value)},
		Message: msg,
		Error:   "maxlength",
		Func:    func(v string) bool { return len(v) < value },
	}
}

func Pattern(pattern, msg string) *Validator {
	return &Validator{
		Attrs:   map[string]string{"pattern": pattern},
		Message: msg,
		Error:   "pattern",
		Func:    func(v string) bool { return true },
	}
}

/*package forms

import (
	"strings"
)

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
			sel, ok := f.Fields[id].(*SelectField)
			if ok {
				for _, value := range sel.Values {
					if v == value {
						return ""
					}
				}

				return message
			}

			panic("should not reach here: " + id)
		},
	}
}
*/
