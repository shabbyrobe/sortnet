package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/shabbyrobe/sortnet"
)

type Gen struct {
	Input    Input
	Network  sortnet.Network
	Exported bool
	Forwards bool
	Slice    bool
	Array    bool
}

func (g Gen) Last() int {
	return g.Network.Size - 1
}

func (g Gen) SliceName() string {
	return g.Input.Name(g.Input.IsExported(), g.Network.Size, g.Forwards, "")
}

func (g Gen) ArrayName() string {
	return g.Input.Name(g.Input.IsExported(), g.Network.Size, g.Forwards, "Array")
}

var genFuncs = template.FuncMap{
	"cas": func(input Input, fwd bool, op sortnet.CompareAndSwap) string {
		var buf bytes.Buffer
		var tpl *template.Template

		if fwd {
			tpl = input.GreaterTemplate
		} else {
			tpl = input.LessTemplate
		}

		if err := tpl.Execute(&buf, op); err != nil {
			panic(err)
		}
		return strings.TrimSpace(buf.String()) + "\n"
	},
}

var genTpl = template.Must(template.New("").Funcs(genFuncs).Parse(`
{{ if .Slice }}
func {{.SliceName}}(a []{{.Input.Type}}) {
	_ = a[{{.Last}}]
	{{ range .Network.Ops }}
	{{- cas $.Input $.Forwards . }}
	{{- end -}}
}
{{ end }}

{{ if .Array }}
func {{.ArrayName}}(a [{{.Network.Size}}]{{.Input.Type}}) {
	{{ range .Network.Ops }}
	{{- cas $.Input $.Forwards . }}
	{{- end -}}
}
{{ end }}
`))

var defaultCASGreaterTpl = template.Must(template.New("").Parse(`
if a[{{.From}}] > a[{{.To}}] {
	a[{{.From}}], a[{{.To}}] = a[{{.To}}], a[{{.From}}]
}
`))

var defaultCASLessTpl = template.Must(template.New("").Parse(`
if a[{{.From}}] < a[{{.To}}] {
	a[{{.From}}], a[{{.To}}] = a[{{.To}}], a[{{.From}}]
}
`))

type WrapperKey struct {
	Input    int
	Forwards bool
}

type WrapperGen struct {
	Input    Input
	Forwards bool
	Methods  map[int]string // template.Template visits these in order
	Wrap     bool
}

func (w WrapperGen) Name() string {
	suffix := ""
	if !w.Forwards {
		suffix = "Reverse"
	}
	return fmt.Sprintf("NetworkSort%s%s", w.Input.TypeNamePart(), suffix)
}

var wrapperTpl = template.Must(template.New("").Parse(`
{{ if .Wrap }}
// {{.Name}} sorts the input according to its length using a sorting network
// if one is available. If the sort was applied, 'ok' is true, otherwise it
// is false to allow you to perform your own sort as a fallback.
//
func {{.Name}}(a []{{.Input.Type}}) (ok bool) {
	switch len(a) {
	{{- range $sz, $name := .Methods }}
	case {{$sz}}:
		{{$name}}(a)
	{{- end }}
	default:
		return false
	}
	return true
}
{{ end }}
`))
