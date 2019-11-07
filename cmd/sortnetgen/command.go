package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/shabbyrobe/sortnet"
	"golang.org/x/tools/imports"
)

const Usage = `sortnetgen: generates a sorting network of specific sizes

Usage: sortnetgen [options] <input>...

Flags may appear at any point; they set the value for the remaining inputs.
This will generate forward and reverse sorters for string(2), and forward-only
sorters for int(2) and uint(2):
    -fwd -rev -size 2 string -rev=false int uint

Each '<input>' contains a fully qualified type name and a 'sizespec', where 'sizespec'
is a comma-separated list of ints or hyphen-separated int ranges.

Generate forward and reverse sorting network of sizes 3-5 for float64:
    -fwd -rev -size 3-5 float64

Generate forward sorting network of sizes 3-5 for int64
    -fwd -size 3-5 int64

Generate reverse sorting network of sizes 3, 4, 5 and 9 for string:
    -rev size 3-5, string

The type will be the basis for the comparison. If <input> is a builtin primitive, '<' is
used for comparisons, otherwise -greater and -less are used to determine how to compare
and swap for -fwd and -rev sorts respectively.

Generate forward sorting network of size 2 for example.com/foo.Yep, providing -greater:
    -size 2 -greater 'foo.YepCASGreater(&a[{{.From}}], &a[{{.To}}])' example.com/foo.Yep

If neither '{{.From}}' nor '{{.To}}' are present in the template, it is presumed to
be a function. The following is equivalent to the previous example:
    -size 2 -greater 'foo.YepCASGreater' example.com/foo.Yep

Only one of -less or -greater needs to be provided, regardless of whether -fwd and/or
-rev are passed. If -less is passed but only -fwd is used, the generator knows how to
call the function with the correct arguments.
`

type usageError string

func (u usageError) Error() string { return string(u) }

func IsUsageError(err error) bool {
	_, ok := err.(usageError)
	return ok
}

type inputFlags struct {
	lessTemplate    string
	greaterTemplate string
	array           bool
	slice           bool
	wrap            bool
	forward         bool
	forceExport     bool
	reverse         bool
	sizes           sizeSpec
}

func (i *inputFlags) parseFlagsAgain(args []string) ([]string, error) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	i.Flags(fs)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	args = fs.Args()
	return args, nil
}

func (i *inputFlags) Flags(flags *flag.FlagSet) {
	flags.BoolVar(&i.forward, "fwd", i.forward, "Generate ascending-order sorters (defaults to true if neither -fwd nor -rev passed)")
	flags.BoolVar(&i.reverse, "rev", i.reverse, "Generate descending-order sorters")
	flags.BoolVar(&i.array, "array", i.array, "Generate fixed-length array sorters")
	flags.BoolVar(&i.slice, "slice", i.slice, "Generate slice sorters")
	flags.BoolVar(&i.wrap, "wrap", i.wrap, "Generate wrapper sorter that chooses the right sort based on len(a)")
	flags.BoolVar(&i.forceExport, "export", false, "Always export the following sorters, even if the type itself is not exported (needed for builtins)")
	flags.StringVar(&i.greaterTemplate, "greater", i.greaterTemplate, "Template for 'compare-and-swap' function")
	flags.StringVar(&i.lessTemplate, "less", i.lessTemplate, "Like -greater, except used for reverse sorting")
	flags.Var(&i.sizes, "size", "Size set; comma separated list of individual sizes or ranges")
}

func (i *inputFlags) BuildLessTemplate() (*template.Template, error) {
	var err error
	var lessTemplate *template.Template
	if i.lessTemplate != "" {
		lessTemplate, err = BuildComparatorTemplate(i.lessTemplate)
		if err != nil {
			return nil, err
		}
	}
	return lessTemplate, nil
}

func (i *inputFlags) BuildGreaterTemplate() (*template.Template, error) {
	var err error
	var greaterTemplate *template.Template
	if i.greaterTemplate != "" {
		greaterTemplate, err = BuildComparatorTemplate(i.greaterTemplate)
		if err != nil {
			return nil, err
		}
	}
	return greaterTemplate, nil
}

type Command struct {
	inputFlags
	pkg    string
	prefix string
	format bool
	out    string
}

func (cmd *Command) Flags(flags *flag.FlagSet) {
	flags.StringVar(&cmd.pkg, "pkg", os.Getenv("GOPACKAGE"), "package name")
	flags.StringVar(&cmd.out, "o", "sortnet_gen.go", "output file name ('-' for stdout)")
	flags.BoolVar(&cmd.format, "format", true, "run gofmt on result")

	cmd.inputFlags.slice = true
	cmd.inputFlags.wrap = true
	cmd.inputFlags.Flags(flags)
}

func (cmd *Command) Synopsis() string { return "Generate enum-ish helpers from a bag of constants" }

func (cmd *Command) Usage() string { return strings.Replace(Usage, "\t", "    ", -1) }

func (cmd *Command) readInputs(args []string) ([]Input, error) {
	var curArgs = cmd.inputFlags
	if !curArgs.forward && !curArgs.reverse {
		curArgs.forward = true
	}

	var err error
	var inputs = make([]Input, 0, len(args))
	var idx = 0

	for {
		args, err = curArgs.parseFlagsAgain(args)
		if err != nil {
			return nil, err
		}
		if len(args) == 0 {
			break
		}

		var arg string
		arg, args = args[0], args[1:]
		idx++

		input, err := ParseInput(arg)
		if err != nil {
			return nil, err
		}
		input.Slice = curArgs.slice
		input.Array = curArgs.array
		input.Wrap = curArgs.wrap
		input.Forward = curArgs.forward
		input.Reverse = curArgs.reverse
		input.Sizes = curArgs.sizes.items
		input.ForceExport = curArgs.forceExport

		input.LessTemplate, err = curArgs.BuildLessTemplate()
		if err != nil {
			return nil, err
		}

		input.GreaterTemplate, err = curArgs.BuildGreaterTemplate()
		if err != nil {
			return nil, err
		}
		if err := input.Validate(); err != nil {
			return nil, fmt.Errorf("input %q failed at index %d: %w", arg, idx, err)
		}

		inputs = append(inputs, input)
		idx++
	}

	return inputs, nil
}

func (cmd *Command) Run(args ...string) (err error) {
	if cmd.out == "" {
		return usageError("-out not set")
	}
	if cmd.pkg == "" {
		return usageError("-pkg not set")
	}

	inputs, err := cmd.readInputs(args)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	{ // File init
		const preamble = "// Code generated by 'github.com/shabbyrobe/sortnet/cmd/sortnetgen'. DO NOT EDIT."
		buf.WriteString(preamble)
		buf.WriteString("\n\n")
		buf.WriteString(fmt.Sprintf("package %s\n\n", cmd.pkg))

		for _, input := range inputs {
			if input.Package != "" {
				buf.WriteString(fmt.Sprintf("import %q\n", input.Package))
			}
		}
	}

	var genBuf bytes.Buffer
	var wrappersIndex = map[wrapperKey]*wrapperGen{}
	var wrappers []*wrapperGen

	{ // Individual networks
		var gens []gen
		for inputIndex, input := range inputs {
			for _, sz := range input.Sizes {
				net := sortnet.New(sz)
				g := gen{
					Input:    input,
					Exported: input.isExported(),
					Network:  net,
				}

				fwds := []bool{}
				if input.Forward {
					fwds = append(fwds, true)
				}
				if input.Reverse {
					fwds = append(fwds, false)
				}

				for _, fwd := range fwds {
					g.Forwards = fwd
					gens = append(gens, g)
					wg := wrappersIndex[wrapperKey{inputIndex, fwd}]
					if wg == nil {
						wg = &wrapperGen{
							Input:    input,
							Forwards: fwd,
							Exported: input.isExported(),
							Methods:  map[int]string{},
						}
						wrappersIndex[wrapperKey{inputIndex, fwd}] = wg
						wrappers = append(wrappers, wg)
					}
					wg.Methods[sz] = g.SliceName()
				}
			}
		}

		sort.Slice(gens, func(i, j int) bool { return gens[i].SortKey() < gens[j].SortKey() })

		for _, gen := range gens {
			if err := genTpl.Execute(&genBuf, gen); err != nil {
				return err
			}
		}
	}

	{ // Wrappers
		sort.Slice(wrappers, func(i, j int) bool { return wrappers[i].SortKey() < wrappers[j].SortKey() })

		for _, wrapper := range wrappers {
			if err := wrapperTpl.Execute(&buf, wrapper); err != nil {
				return err
			}
		}
	}

	// Wrappers should go above individual functions:
	buf.Write(genBuf.Bytes())

	var out = buf.Bytes()

	{ // Gofmt
		if cmd.format {
			var err error
			out, err = imports.Process(cmd.out, out, nil)
			if err != nil {
				return err
			}
		}
	}

	{ // Write output
		if cmd.out == "-" {
			os.Stdout.Write(out)
		} else {
			var write bool
			existing, err := ioutil.ReadFile(cmd.out)
			if os.IsNotExist(err) || err == nil {
				write = true
			} else if err != nil {
				return err
			} else if !bytes.Equal(out, existing) {
				write = true
			}

			if write {
				return ioutil.WriteFile(cmd.out, out, 0644)
			}
		}
	}

	return nil
}
