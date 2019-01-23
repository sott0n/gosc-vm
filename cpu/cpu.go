//
// This is a simple port of the virtual machine interpreter to golang.
//
// For example the loop script could be compiled to bytecode like this:
//
//    ./compiler examples/loop.in
//
// Once that has been done it can be executed:
//    go run main.go examples/loop.raw
//
//

package cpu

// Flags holds the CPU flags - of which we only have one.
type Flags struct {
	// Zero-flag
	z bool
}

// Register holds the contents of a single register.
// This is horrid because we don't use an enum for the type.
type Register struct {
	// Integer contents of register if t == "int"
	i int
	// String contents of register if t == "string"
	s string
	// Register type: "int" vs. "string"
	t string
}

// Stack holds return-addresses when the `call` operation is being completed.
// It can also be used for storing ints.
type Stack struct {
	// The entries on our stack
	entries []int
}

// CPU is our virtual machine state.
type CPU struct {
	// Registers
	regs [16]Register
	// Flags
	flags Flags
	// Our RAM - where the program is loaded
	mem [0xFFFF]byte
	// Instruction-pointer
	ip int
	// stack
	stack *Stack
}
