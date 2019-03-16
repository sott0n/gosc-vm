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
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
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

//
// CPU/VM functions
//

// NewCPU returns a new CPU object.
func NewCPU() *CPU {
	x := &CPU{}
	x.Reset()
	return x
}

// Reset sets the CPU into a known-good state, by setting the IP to zero.
// and emptying all registers (i.e. setting them to zero too).
func (c *CPU) Reset() {
	for i := 0; i < 16; i++ {
		c.regs[i].SetInt(0)
	}
	c.ip = 0
	c.stack = NewStack()
}

// LoadFile loads the program from the named file into RAM.
func (c *CPU) LoadFile(path string) {
	// Ensure we reset our state.
	c.Reset()

	// Load the file.
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Failed to read file: %s - %s\n", path, err.Error())
		os.Exit(1)
	}

	if len(b) >= 0xFFFF {
		fmt.Printf("Program too large for RAM!\n")
		os.Exit(1)
	}

	// Copy contents of file to our memory region.
	for i := 0; i < len(b); i++ {
		c.mem[i] = b[i]
	}
}

// LoadBytes populates the given program into RAM.
func (c *CPU) LoadBytes(data []byte) {
	// Ensure we reset our state.
	c.Reset()

	if len(data) >= 0xFFFF {
		fmt.Printf("Program too large for RAM!\n")
		os.Exit(1)
	}

	// Copy contents of file to our memory region.
	for i := 0; i < len(data); i++ {
		c.mem[i] = data[i]
	}
}

// readString reads a string from the IP position
// Strings are prefixed by their length (two-bytes).
func (c *CPU) readString() string {
	// Read the length of the string we expect.
	len := c.read2Val()

	// Now build up the body of the string.
	s := ""
	for i := 0; i < len; i++ {
		s += string(c.mem[c.ip+i])
	}

	// Jump the IP over the length of the string.
	c.ip += (len)
	return s
}

// Read a two-byte number from the current IP.
// i.e. This reads two bytes and returns a 16-bit value to the caller,
// skipping over both bytes in the IP.
func (c *CPU) read2Val() int {
	l := int(c.mem[c.ip])
	c.ip++
	h := int(c.mem[c.ip])
	c.ip++

	val := l + h*256
	return (val)
}

// Run launches our interpreter.
func (c *CPU) Run() {
	run := true
	for run {
		instruction := c.mem[c.ip]
		debugPrintf("About to execute instruction %02X\n", instruction)

		switch instruction {
		case 0x00:
			debugPrintf("EXIT\n")
			run = false

		case 0x01:
			debugPrintf("INT_STORE\n")
			// register
			c.ip++
			reg := int(c.mem[c.ip])
			c.ip++
			val := c.read2Val()

			debugPrintf("\tSet register %02X to %04X\n", reg, val)
			c.regs[reg].SetInt(val)

		case 0x02:
			debugPrintf("INT_PRINT\n")
			// register
			c.ip++
			reg := c.mem[c.ip]

			val := c.regs[reg].GetInt()
			if val < 256 {
				fmt.Printf("%02X", val)
			} else {
				fmt.Printf("%04X", val)
			}
			c.ip++

		case 0x03:
			debugPrintf("INT_TOSTRING\n")
			// register
			c.ip++
			reg := c.mem[c.ip]

			// get value
			i := c.regs[reg].GetInt()

			// change from int to string
			c.regs[reg].SetString(fmt.Sprintf("%d", i))

			// next instruction
			c.ip++

		case 0x04:
			debugPrintf("INT_RANDOM\n")
			// register
			c.ip++
			reg := c.mem[c.ip]

			// New random source
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)

			// New random number
			c.regs[reg].SetInt(r1.Intn(0xffff))
			c.ip++

		case 0x10:
			debugPrintf("JUMP\n")
			c.ip++
			addr := c.read2Val()
			c.ip = addr

		case 0x11:
			debugPrintf("JUMP_Z\n")
			c.ip++
			addr := c.read2Val()
			if c.flags.z {
				c.ip = addr
			}

		case 0x12:
			debugPrintf("JUMP_NZ\n")
			c.ip++
			addr := c.read2Val()
			if !c.flags.z {
				c.ip = addr
			}

		case 0x13:
			debugPrintf("XOR\n")
			c.ip++
			reg := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[reg].SetInt(aVal ^ bVal)

		case 0x21:
			debugPrintf("ADD\n")
			c.ip++
			reg := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[reg].SetInt(aVal + bVal)

		case 0x22:
			debugPrintf("SUB\n")
			c.ip++
			reg := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[reg].SetInt(aVal - bVal)

			// set the zero-flag if the result was zero or less
			if c.regs[reg].i <= 0 {
				c.flags.z = true
			}

		case 0x23:
			debugPrintf("MUL\n")
			c.ip++
			reg := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[reg].SetInt(aVal * bVal)

		case 0x24:
			debugPrintf("DIV\n")
			c.ip++
			reg := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()

			if bVal == 0 {
				fmt.Printf("Attempting to divide by zero - denying\n")
				os.Exit(3)
			}
			c.regs[reg].SetInt(aVal / bVal)

		case 0x25:
			debugPrintf("INC\n")
			// register
			c.ip++
			reg := c.mem[c.ip]
			c.regs[reg].SetInt(c.regs[reg].GetInt() + 1)
			// bump past that
			c.ip++

		case 0x26:
			debugPrintf("DEC\n")
			// register
			c.ip++
			reg := c.mem[c.ip]
			c.regs[reg].SetInt(c.regs[reg].GetInt() - 1)
			// bump past that
			c.ip++

		case 0x27:
			debugPrintf("AND\n")
			c.ip++
			reg := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[reg].SetInt(aVal & bVal)

		case 0x28:
			debugPrintf("OR\n")
			c.ip++
			reg := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			// store result
			aVal := c.regs[a].GetInt()
			bVal := c.regs[b].GetInt()
			c.regs[reg].SetInt(aVal | bVal)

		case 0x30:
			debugPrintf("STORE_STRING\n")
			// register
			c.ip++
			reg := c.mem[c.ip]
			// bump past that to the length + string
			c.ip++
			// read it
			str := c.readString()
			debugPrintf("\tRead String: '%s'\n", str)
			// store the string
			c.regs[reg].SetString(str)

		case 0x31:
			debugPrintf("PRINT_STRING\n")
			// register
			c.ip++
			reg := c.mem[c.ip]
			fmt.Printf("%s", c.regs[reg].GetString())
			c.ip++

		case 0x32:
			debugPrintf("STRING_CONCAT\n")
			c.ip++
			reg := c.mem[c.ip]
			c.ip++
			a := c.mem[c.ip]
			c.ip++
			b := c.mem[c.ip]
			c.ip++

			// store result
			aVal := c.regs[a].GetString()
			bVal := c.regs[b].GetString()
			c.regs[reg].SetString(aVal + bVal)

		case 0x33:
			debugPrintf("SYSTEM\n")
			// register
			c.ip++
			reg := c.mem[c.ip]
			c.ip++

			// run the command
			toExec := splitCommand(c.regs[reg].GetString())
			cmd := exec.Command(toExec[0], toExec[1:]...)

			var out bytes.Buffer
			var err bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &err
			cmd.Run()

			// stdout
			fmt.Printf("%s", out.String())
			// stderr - if err is non-empty
			if len(err.String()) > 0 {
				fmt.Printf("%s", err.String())
			}

		case 0x34:
			debugPrintf("STRING_TOINT\n")
			// register
			c.ip++
			reg := c.mem[c.ip]

			// get value
			s := c.regs[reg].GetString()
			i, err := strconv.Atoi(s)
			if err == nil {
				c.regs[reg].SetInt(i)
			} else {
				fmt.Printf("Failed to convert '%s' to int: %s", s, err.Error())
				os.Exit(3)
			}

			// next instruction
			c.ip++

		case 0x40:
			debugPrintf("CMP_REG\n")
			c.ip++
			r1 := int(c.mem[c.ip])
			c.ip++
			r2 := int(c.mem[c.ip])
			c.ip++

			c.flags.z = false

			switch c.regs[r1].Type() {
			case "int":
				if c.regs[r1].GetInt() == c.regs[r2].GetInt() {
					c.flags.z = true
				}
			case "string":
				if c.regs[r1].GetString() == c.regs[r2].GetString() {
					c.flags.z = true
				}
			}

		case 0x41:
			debugPrintf("CMP_IMMEDIATE\n")
			c.ip++
			reg := int(c.mem[c.ip])
			c.ip++
			val := c.read2Val()

			if c.regs[reg].Type() == "int" && c.regs[reg].GetInt() == val {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case 0x42:
			debugPrintf("CMP_STR\n")
			c.ip++
			reg := int(c.mem[c.ip])
			c.ip++

			str := c.readString()

			if c.regs[reg].Type() == "string" && c.regs[reg].GetString() == str {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case 0x43:
			debugPrintf("IS_STRING\n")
			c.ip++
			reg := int(c.mem[c.ip])
			c.ip++

			if c.regs[reg].Type() == "string" {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case 0x44:
			debugPrintf("IS_INT\n")
			c.ip++
			reg := int(c.mem[c.ip])
			c.ip++

			if c.regs[reg].Type() == "int" {
				c.flags.z = true
			} else {
				c.flags.z = false
			}

		case 0x50:
			debugPrintf("NOP\n")
			c.ip++

		case 0x51:
			debugPrintf("STORE\n")
			c.ip++
			dst := int(c.mem[c.ip])
			c.ip++
			src := int(c.mem[c.ip])
			c.ip++

			c.regs[src] = c.regs[dst]

		case 0x60:
			debugPrintf("PEEK\n")
			c.ip++
			result := int(c.mem[c.ip])
			c.ip++
			src := int(c.mem[c.ip])

			// get the address from the src register contents.
			addr := c.regs[src].i

			// store the contents of the given address.
			c.regs[result].SetInt(int(c.mem[addr]))
			c.ip++

		case 0x61:
			debugPrintf("POKE\n")
			c.ip++
			src := int(c.mem[c.ip])
			c.ip++

			dst := int(c.mem[c.ip])
			c.ip++

			// So the destination will contain an address
			// put the contents of the source to that.
			addr := c.regs[dst].i
			val := c.regs[src].i

			debugPrintf("Writing %02X to %04X\n", val, addr)
			c.mem[addr] = byte(val)

		case 0x62:
			debugPrintf("MEMCPY\n")
			c.ip++
			dst := int(c.mem[c.ip])
			c.ip++

			src := int(c.mem[c.ip])
			c.ip++

			len := int(c.mem[c.ip])
			c.ip++

			// get the addresses from the registers
			src_addr := c.regs[src].GetInt()
			dst_addr := c.regs[dst].GetInt()
			length := c.regs[len].GetInt()

			i := 0
			for i < length {
				if dst_addr >= 0xFFFF {
					dst_addr = 0
				}
				if src_addr >= 0xFFFF {
					src_addr = 0
				}

				c.mem[dst_addr] = c.mem[src_addr]
				dst_addr += 1
				src_addr += 1
				i += 1
			}

		case 0x70:
			debugPrintf("PUSH\n")
			c.ip++
			reg := int(c.mem[c.ip])
			c.ip++

			// Store the value in the register on stack
			c.stack.Push(c.regs[reg].GetInt())

		case 0x71:
			debugPrintf("POP\n")
			c.ip++
			reg := int(c.mem[c.ip])
			c.ip++

			if c.stack.Empty() {
				fmt.Printf("Stack Underflow!\n")
				os.Exit(1)
			}
			// Store the value in the register on stack
			c.regs[reg].SetInt(c.stack.Pop())

		case 0x72:
			debugPrintf("RET\n")

			// Ensure our stack isn't empty
			if c.stack.Empty() {
				fmt.Printf("Stack Underflow!\n")
				os.Exit(1)
			}

			addr := c.stack.Pop()
			c.ip = addr

		case 0x73:
			debugPrintf("CALL\n")
			c.ip++

			addr := c.read2Val()

			c.stack.Push(c.ip)
			c.ip = addr

		default:
			fmt.Printf("Unrecognized/Unimplemented opcode %02X at IP %04X\n", instruction, c.ip)
			os.Exit(1)
		}

		// Ensure our instruction-pointer wraps around.
		if c.ip >= 0xFFFF {
			c.ip = 0
		}
	}
}
