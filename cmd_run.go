package main

import (
	"context"
	"flag"
	"fmt"
	"gosc-vm/compiler"
	"gosc-vm/cpu"
	"gosc-vm/lexer"
	"io/ioutil"

	"github.com/google/subcommands"
)

type runCmd struct {
}

//
// Glue
//
func (*runCmd) Name() string     { return "run" }
func (*runCmd) Synopsis() string { return "Run the given source program." }
func (*runCmd) Usage() string {
	return `run :
  The run sub-command compiles the given source program, and then executes
  it immediately.
`
}

//
// Flag setup: no flags
//
func (p *runCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *runCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	for _, file := range f.Args() {
		fmt.Printf("Parsing file: %s\n", file)

		// Read the file.
		input, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading %s - %s\n", file, err.Error())
			return subcommands.ExitFailure
		}

		// Lex it
		l := lexer.New(string(input))

		// Compile it
		e := compiler.New(l)
		e.Compile()

		// Now create a machine to run the compiled program in
		c := cpu.NewCPU()

		// Load the program
		c.LoadBytes(e.Output())

		// Run the machine
		c.Run()
	}
	return subcommands.ExitSuccess
}
