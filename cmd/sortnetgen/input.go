package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

type Input struct {
	Forward     bool
	Reverse     bool
	Sizes       []int
	Package     string
	Type        string
	ForceExport bool

	// Generate a sorting network that operates on a slice, for example:
	// NetworkSort2xFloat64(a []float64)
	Slice bool

	// Generate a sorting network that operates on an array, for example:
	// NetworkSort2xFloat64Array(a *[2]float64)
	Array bool

	// Wrap the set of networks sorts produced for the different sizes of a given
	// slice into a method that dispatches to the correct network by length, for example:
	//
	//	func NetworkSortFloat64(a []float64) (ok bool) {
	//     switch len(a) {
	//     case 2:
	//         NetworkSort2xFloat64(a)
	//     case 3:
	//         NetworkSort3xFloat64(a)
	//     default:
	//         return false
	//     }
	//     return true
	//	}
	//
	Wrap bool

	// LessTemplate should presume the existence of the indexable 'a', which is the
	// slice/array being sorted.
	//
	// The `{{.From}}` and `{{.To}}` values are made available to the template, for
	// accessing the comparator's 'from index' and 'to index', respectively.
	//
	// The LessTemplate for standard comparable primitives looks like this:
	//
	//  if a[{{.From}}] < a[{{.To}}] {
	//		a[{{.From}}], a[{{.To}}] = a[{{.To}}], a[{{.From}}]
	//	}
	LessTemplate *template.Template

	// GreaterTemplate should presume the existence of the indexable 'a', which is the
	// slice/array being sorted.
	//
	// The `{{.From}}` and `{{.To}}` values are made available to the template, for
	// accessing the comparator's 'from index' and 'to index', respectively.
	//
	// The GreaterTemplate for standard comparable primitives looks like this:
	//
	//  if a[{{.From}}] > a[{{.To}}] {
	//		a[{{.From}}], a[{{.To}}] = a[{{.To}}], a[{{.From}}]
	//	}
	GreaterTemplate *template.Template
}

func (in *Input) name(exported bool, sz int, fwd bool, suffix string) (out string) {
	prefix := "NetworkSort"
	if !exported {
		prefix = "networkSort"
	}
	typ := ucfirst(in.Type)
	out = fmt.Sprintf("%s%dx%s%s", prefix, sz, typ, suffix)
	if !fwd {
		out += "Reverse"
	}
	return out
}

func (in *Input) isExported() bool {
	if in.ForceExport {
		return true
	}
	r, _ := utf8.DecodeRuneInString(in.Type)
	return in.isComparableBuiltin() || unicode.IsUpper(r)
}

func (in *Input) isComparableBuiltin() bool {
	return in.Package == "" && (in.Type == "string" ||
		in.Type == "int" ||
		in.Type == "int8" ||
		in.Type == "int16" ||
		in.Type == "int32" ||
		in.Type == "int64" ||
		in.Type == "uint" ||
		in.Type == "uint8" ||
		in.Type == "uint16" ||
		in.Type == "uint32" ||
		in.Type == "uint64" ||
		in.Type == "float32" ||
		in.Type == "float64")
}

func (in *Input) Validate() error {
	if len(in.Sizes) == 0 {
		return fmt.Errorf("no sizes to generate")
	}

	for _, sz := range in.Sizes {
		if sz <= 0 {
			return fmt.Errorf("sizes must be >= 1")
		}
	}

	if in.isComparableBuiltin() {
		if in.LessTemplate == nil {
			in.LessTemplate = defaultCASLessTpl
		}
		if in.GreaterTemplate == nil {
			in.GreaterTemplate = defaultCASGreaterTpl
		}

	} else {
		if (in.Reverse || in.Forward) && (in.LessTemplate == nil && in.GreaterTemplate == nil) {
			return fmt.Errorf("no -less or -greater provided for non-builtin input - only builtins can be compared using '<' or '>' so you have to provide a function")
		}
	}

	return nil
}

func BuildComparatorTemplate(raw string) (*template.Template, error) {
	tpl, err := template.New("").Parse(raw)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, map[string]string{"From": "1", "To": "2"}); err != nil {
		return nil, err
	}

	// If there are no substitutions, presume the input template contains a function
	// name only.
	if buf.String() == raw {
		tpl, err = template.New("").Parse(raw + "(&a[{{.From}}], &a[{{.To}}])")
		if err != nil {
			return nil, err
		}
	}

	return tpl, nil
}

const (
	inputPkg = 1
	inputTyp = 2
)

var inputPattern = regexp.MustCompile(`` +
	`^` +
	`(?:(?P<pkg>[\pL\d\/\.]+)\.)?` + // Greedy
	`(?P<typ>\pL[\pL\d]*)` +
	`$`,
)

func ParseInput(s string) (input Input, err error) {
	match := inputPattern.FindStringSubmatch(s)
	if len(match) == 0 {
		return input, fmt.Errorf("invalid input %q", s)
	}

	input.Package = match[inputPkg]
	input.Type = match[inputTyp]

	return input, nil
}

func ucfirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}
