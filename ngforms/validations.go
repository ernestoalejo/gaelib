package ngforms

import (
	"fmt"
	"regexp"
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
	re := regexp.MustCompile(pattern)

	return &Validator{
		Attrs:   map[string]string{"pattern": pattern},
		Message: msg,
		Error:   "pattern",
		Func:    func(v string) bool { return re.MatchString(v) },
	}
}

func Email(msg string) *Validator {
	re := regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,4}$`)

	return &Validator{
		Attrs:   map[string]string{},
		Message: msg,
		Error:   "email",
		Func:    func(v string) bool { return re.MatchString(v) },
	}
}

func Match(f *Form, field, msg string) *Validator {
	return &Validator{
		Attrs:   map[string]string{"match": field},
		Message: msg,
		Error:   "match",
		Func: func(v string) bool {
			ctrl := f.GetControl(field)
			if ctrl != nil {
				return ctrl.value == v
			}
			panic("should not reach here: " + field)
		},
	}
}

func Select(f *Form, field, msg string) *Validator {
	return &Validator{
		Attrs:   map[string]string{},
		Message: msg,
		Error:   "select",
		Func: func(v string) bool {
			sel, ok := f.Fields[field].(*SelectField)
			if ok {
				for _, value := range sel.Values {
					if v == value {
						return true
					}
				}

				return false
			}

			panic("select field not found in form (or not a select): " + field)
		},
	}
}
