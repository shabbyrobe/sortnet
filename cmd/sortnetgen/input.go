package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

type Input struct {
	Forward bool
	Reverse bool
	Sizes   []int
	Package string
	Type    string

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

func (in *Input) typeNamePart() string {
	typ := in.Type
	if in.isComparableBuiltin() {
		typ = strings.ToUpper(typ[:1]) + typ[1:]
	}
	return typ
}

func (in *Input) name(exported bool, sz int, fwd bool, suffix string) (out string) {
	prefix := "NetworkSort"
	if !exported {
		prefix = "networkSort"
	}
	out = fmt.Sprintf("%s%dx%s%s", prefix, sz, in.typeNamePart(), suffix)
	if !fwd {
		out += "Reverse"
	}
	return out
}

func (in *Input) isExported() bool {
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

func (in *Input) ensureTemplates() error {
	if in.Reverse && in.LessTemplate == nil {
		if in.isComparableBuiltin() {
			in.LessTemplate = defaultCASLessTpl
		} else {
			return fmt.Errorf("no -less template provided for non-builtin input")
		}
	}

	if in.Forward && in.GreaterTemplate == nil {
		if in.isComparableBuiltin() {
			in.GreaterTemplate = defaultCASGreaterTpl
		} else {
			return fmt.Errorf("no -greater template provided for non-builtin input")
		}
	}

	return nil
}

const (
	inputPkg   = 1
	inputTyp   = 2
	inputSizes = 3
)

var inputPattern = regexp.MustCompile(`` +
	`^` +
	`(?:(?P<pkg>.*)\.)?` + // Greedy
	`(?P<typ>.*?)` +
	`:` +
	`(?P<sizes>.*?)` +
	`$`,
)

const (
	sizeSize = 1
	sizeTo   = 2
)

var sizePattern = regexp.MustCompile(`` +
	`^` +
	`(?P<size>[0-9]+)` +
	`(?:-(?P<to>[0-9]+))?` +
	`$`,
)

func ParseInput(s string, idx int) (input Input, err error) {
	match := inputPattern.FindStringSubmatch(s)
	if len(match) == 0 {
		return input, fmt.Errorf("invalid input %q", s)
	}

	input.Package = match[inputPkg]
	input.Type = match[inputTyp]

	sizeSet := make(map[int]bool)
	for _, sizePart := range strings.Split(match[inputSizes], ",") {
		sizeMatch := sizePattern.FindStringSubmatch(sizePart)
		if len(sizeMatch) == 0 {
			return input, fmt.Errorf("invalid size %q", sizePart)
		}

		from, err := strconv.ParseInt(sizeMatch[sizeSize], 10, 0)
		if err != nil {
			return input, fmt.Errorf("size was not numeric in %q", sizePart)
		}
		to := from
		if sizeMatch[sizeTo] != "" {
			to, err = strconv.ParseInt(sizeMatch[sizeTo], 10, 0)
			if err != nil {
				return input, fmt.Errorf("size range end was not numeric in %q", sizePart)
			}
		}
		if to < from {
			return input, fmt.Errorf("size range end was before start in %q", sizePart)
		}

		for i := from; i <= to; i++ {
			if i <= 0 {
				return input, fmt.Errorf("sort size must be >= 1, found %d", i)
			}
			sizeSet[int(i)] = true
		}
	}

	for sz := range sizeSet {
		input.Sizes = append(input.Sizes, sz)
	}

	sort.Ints(input.Sizes)

	if len(input.Sizes) == 0 {
		return input, fmt.Errorf("input %q contained no sizes", s)
	}

	return input, nil
}
