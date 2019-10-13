package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/shabbyrobe/sortnet"
	"golang.org/x/tools/imports"
)

const Usage = `sortnetgen: generates a sorting network of specific sizes

Usage: sortnetgen [options] <input>...

Flags may appear at any point; they set the value for the remaining inputs.
This will generate forward and reverse sorters for string(2), and forward-only
sorters for int(2) and uint(2):
	-fwd -rev string:2 -rev=false int:2 uint:2

Each <input> contains a fully qualified type name and a 'sizespec', where 'sizespec'
is a comma-separated list of 
interpretation of the below grammar:

Generate forward and reverse sorting network of sizes 3-5 for example.com/yep.Foo
    -fwd rev example.com/yep.Foo:3-5

Generate forward sorting network of sizes 3-5 for example.com/yep.Foo
    -fwd example.com/yep.Foo:3-5

Generate reverse sorting network of sizes 3, 4, 5 and 9 for string:
    -rev string:3-5,9

The type will be the basis for the comparison. If <input> is a builtin primitive, '<' is
used for comparisons, otherwise -castpl is used to determine how to compare and swap.

If the type contains a package, it is imported. It is not validated.

If the leading direction specifier is not present, '+' is inferred, otherwise if '+' is
found, a forward sort method is generated and if '-' is found, a reverse sort method is
generated.
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
	reverse         bool
}

func (i *inputFlags) parseAgain(args []string) ([]string, error) {
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
	flags.BoolVar(&i.wrap, "wrap", i.wrap, "Generate wrapper sorter that chooses the right sort based on len(a) and returns false if none present")
	flags.StringVar(&i.greaterTemplate, "greater", "", ""+
		"'compare-and-swap' template that evaluates to true if the first value is greater "+
		"than the second. Used if the sort values are structs.")
	flags.StringVar(&i.lessTemplate, "less", "", ""+
		"Like -greater, except used for reverse sorting")
}

func (i *inputFlags) BuildLessTemplate() (*template.Template, error) {
	var err error
	var lessTemplate *template.Template
	if i.lessTemplate != "" {
		lessTemplate, err = template.New("").Parse(i.lessTemplate)
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
		greaterTemplate, err = template.New("").Parse(i.greaterTemplate)
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
	flags.StringVar(&cmd.out, "out", "sortnet_gen.go", "output file name ('-' for stdout)")
	flags.BoolVar(&cmd.format, "format", true, "run gofmt on result")

	cmd.inputFlags.slice = true
	cmd.inputFlags.wrap = true
	cmd.inputFlags.Flags(flags)
}

func (cmd *Command) Synopsis() string { return "Generate enum-ish helpers from a bag of constants" }

func (cmd *Command) Usage() string { return Usage }

func (cmd *Command) Run(args ...string) (err error) {
	if cmd.out == "" {
		return usageError("-out not set")
	}
	if cmd.pkg == "" {
		return usageError("-pkg not set")
	}

	var curArgs = cmd.inputFlags
	if !curArgs.forward && !curArgs.reverse {
		curArgs.forward = true
	}

	var inputs = make([]Input, 0, len(args))
	var idx = 0

	for {
		args, err = curArgs.parseAgain(args)
		if err != nil {
			return err
		}
		if len(args) == 0 {
			break
		}

		var arg string
		arg, args = args[0], args[1:]

		input, err := ParseInput(arg, idx)
		if err != nil {
			return err
		}
		input.Slice = curArgs.slice
		input.Array = curArgs.array
		input.Wrap = curArgs.wrap
		input.Forward = curArgs.forward
		input.Reverse = curArgs.reverse

		input.LessTemplate, err = curArgs.BuildLessTemplate()
		if err != nil {
			return err
		}

		input.GreaterTemplate, err = curArgs.BuildGreaterTemplate()
		if err != nil {
			return err
		}

		if err := input.ensureTemplates(); err != nil {
			return err
		}
		inputs = append(inputs, input)
		idx++
	}

	var buf bytes.Buffer

	{ // File init
		const preamble = "// Code generated by 'github.com/shabbyrobe/go-enumgen'. DO NOT EDIT."
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
	var wrappers = map[wrapperKey]*wrapperGen{}

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
					if wg := wrappers[wrapperKey{inputIndex, fwd}]; wg == nil {
						wrappers[wrapperKey{inputIndex, fwd}] = &wrapperGen{
							Input:    input,
							Forwards: fwd,
							Methods:  map[int]string{},
						}
					}
					wrappers[wrapperKey{inputIndex, fwd}].Methods[sz] = g.SliceName()
				}
			}
		}

		for _, gen := range gens {
			if err := genTpl.Execute(&genBuf, gen); err != nil {
				return err
			}
		}
	}

	{ // Wrappers
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
