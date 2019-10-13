package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/shabbyrobe/sortnet"
)

type gen struct {
	Input    Input
	Network  sortnet.Network
	Exported bool
	Forwards bool
}

func (g gen) Last() int {
	return g.Network.Size - 1
}

func (g gen) SliceName() string {
	return g.Input.Name(g.Input.isExported(), g.Network.Size, g.Forwards, "")
}

func (g gen) ArrayName() string {
	return g.Input.Name(g.Input.isExported(), g.Network.Size, g.Forwards, "Array")
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
{{ if .Input.Slice }}
func {{.SliceName}}(a []{{.Input.Type}}) {
	_ = a[{{.Last}}]
	{{ range .Network.Ops }}
	{{- cas $.Input $.Forwards . }}
	{{- end -}}
}
{{ end }}

{{ if .Input.Array }}
func {{.ArrayName}}(a *[{{.Network.Size}}]{{.Input.Type}}) {
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

type wrapperKey struct {
	Input    int
	Forwards bool
}

type wrapperGen struct {
	Input    Input
	Forwards bool
	Methods  map[int]string // template.Template visits these in order
}

func (w wrapperGen) Name() string {
	suffix := ""
	if !w.Forwards {
		suffix = "Reverse"
	}
	return fmt.Sprintf("NetworkSort%s%s", w.Input.TypeNamePart(), suffix)
}

var wrapperTpl = template.Must(template.New("").Parse(`
{{ if .Input.Wrap }}
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
