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
	Raw             string
	Index           int
	Forward         bool
	Reverse         bool
	Sizes           []int
	Package         string
	Type            string
	LessTemplate    *template.Template
	GreaterTemplate *template.Template
}

func (in *Input) TypeNamePart() string {
	typ := in.Type
	if in.IsComparableBuiltin() {
		typ = strings.ToUpper(typ[:1]) + typ[1:]
	}
	return typ
}

func (in *Input) Name(exported bool, sz int, fwd bool, suffix string) (out string) {
	prefix := "NetworkSort"
	if !exported {
		prefix = "networkSort"
	}
	out = fmt.Sprintf("%s%s%d%s", prefix, in.TypeNamePart(), sz, suffix)
	if !fwd {
		out += "Reverse"
	}
	return out
}

func (in *Input) IsExported() bool {
	r, _ := utf8.DecodeRuneInString(in.Type)
	return in.IsComparableBuiltin() || unicode.IsUpper(r)
}

func (in *Input) IsComparableBuiltin() bool {
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
		if in.IsComparableBuiltin() {
			in.LessTemplate = defaultCASLessTpl
		} else {
			return fmt.Errorf("no -less template provided for non-builtin input %q at index %d", in.Raw, in.Index+1)
		}
	}

	if in.Forward && in.GreaterTemplate == nil {
		if in.IsComparableBuiltin() {
			in.GreaterTemplate = defaultCASGreaterTpl
		} else {
			return fmt.Errorf("no -greater template provided for non-builtin input %q at index %d", in.Raw, in.Index+1)
		}
	}

	return nil
}

const (
	inputDir   = 1
	inputPkg   = 2
	inputTyp   = 3
	inputSizes = 4
)

var inputPattern = regexp.MustCompile(`` +
	`^` +
	`(?P<dir>[+\-]{0,2})?` +
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

	for _, c := range match[inputDir] {
		if c == '+' {
			input.Forward = true
		} else if c == '-' {
			input.Reverse = true
		}
	}
	if !input.Forward && !input.Reverse {
		input.Forward = true
	}

	input.Raw = s
	input.Index = idx
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
				return input, fmt.Errorf("sort size must be >= 1, found %d at input %d", i, input.Index+1)
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
