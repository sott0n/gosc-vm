package main

import (
	"context"
	"flag"
	"fmt"
	"gosc-vm/cpu"

	"github.com/google/subcommands"
)

type executeCmd struct {
}

//
// Glue
//
func (*executeCmd) Name() string     { return "execute" }
func (*executeCmd) Synopsis() string { return "Executed a simple.vm program." }
func (*executeCmd) Usage() string {
	return `execute:
  Execute the bytecodes contained in the given input file.
`
}

//
// Flag setup: no flags
//
func (p *executeCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry point.
//
func (p *executeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	//
	// For each file on the command-line we can now parse and
	// enqueue the jobs
	//
	for _, file := range f.Args() {
		fmt.Printf("Loading file: %s\n", file)
		c := cpu.NewCPU()
		c.LoadFile(file)
		c.Run()
	}
	return subcommands.ExitSuccess
}
