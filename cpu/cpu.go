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

import (
	"fmt"
	"os"
	"regexp"
)

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

//
// Global functions
//

// debugPrintf outputs some debugging details when `$DEBUG=1`.
func debugPrintf(fmtDeb string, args ...interface{}) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	prefix := fmt.Sprintf("%s", fmtDeb)
	fmt.Printf(prefix, args...)
}

// Split a line of text into tokens, but keep anything "quoted" together.
// So this input:
//
// /bin/sh -c "ls /etc"
//
// Would give output of the form:
//   /bin/sh
//   -c
//   ls /etc
//
func splitCommand(input string) []string {
	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'(|^']*)`)
	res := r.FindAllString(input, -1)

	// However the resulting pieces might be quoted.
	// So we have to remove them, if present.
	var result []string
	for _, e := range res {
		result = append(result, trimQuotes(e, '"'))
	}
	return (result)
}

// Remove balanced characters around a string.
func trimQuotes(in string, c byte) string {
	if len(in) >= 2 {
		if in[0] == c && in[len(in)-1] == c {
			return in[1 : len(in)-1]
		}
	}
	return in
}

//
// Register functions
//

// GetInt retrieves the integer content of the given register.
// If the register contains a string that is a fatal error.
func (r *Register) GetInt() int {
	if r.t != "int" {
		fmt.Printf("Error: Attempting to call GetInt on a register holding a non-integer value.\n")
		os.Exit(3)
	}
	return r.i
}

// SetInt stores the given integer in the register.
func (r *Register) SetInt(v int) {
	r.i = v
	r.t = "int"
}

// GetString retrieves the string content of the given register.
// If the register contains a integer that is a fatal error.
func (r *Register) GetString() string {
	if r.t != "string" {
		fmt.Printf("Error: Attempting to call GetString on a register holding a non-string value.\n")
		os.Exit(3)
	}
	return r.s
}

// SetString stores the given string in the register.
func (r *Register) SetString(v string) {
	r.s = v
	r.t = "string"
}

// Type returns the type of a registers contents `int` vs. `string`.
func (r *Register) Type() string {
	return (r.t)
}

//
// Stack functions
//

// NewStack creates a new stack object.
func NewStack() *Stack {
	return &Stack{}
}

// Empty returns boolean, is the stack empty?
func (s *Stack) Empty() bool {
	return (len(s.entries) <= 0)
}

// Push add a value to the stack.
func (s *Stack) Push(value int) {
	s.entries = append(s.entries, value)
}

// Pop removes a value from the stack.
func (s *Stack) Pop() int {
	result := s.entries[0]
	s.entries = append(s.entries[:0], s.entries[1:]...)
	return (result)
}
