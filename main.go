// Command bombo runs Bombo programs.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/RajaeDev6/bombo/interpreter"
	"github.com/RajaeDev6/bombo/parser"
)

const version = "1.0.0"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		repl()
		return
	}
	switch args[0] {
	case "help", "-h", "--help":
		fmt.Print(helpText)
	case "version", "-v", "--version":
		fmt.Println("bombo", version)
	default:
		os.Exit(runFile(args[0]))
	}
}

func report(err error) {
	fmt.Fprintf(os.Stderr, "BOMBOCLAAT! %s\n", err)
}

func runFile(path string) int {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "BOMBOCLAAT! Mi cyaan find or open '%s'\n", path)
		return 1
	}
	prog, err := parser.Parse(string(src))
	if err != nil {
		report(err)
		return 1
	}
	it := interpreter.New(os.Stdout, bufio.NewReader(os.Stdin))
	if err := it.Run(prog); err != nil && !errors.Is(err, interpreter.ErrExit) {
		report(err)
		return 1
	}
	return 0
}

func repl() {
	in := bufio.NewReader(os.Stdin)
	it := interpreter.New(os.Stdout, in)
	fmt.Printf("Bombo %s. Wah gwaan! Type 'likkle more' fi leave, 'help' fi di keywords.\n", version)

	var buf strings.Builder
	for {
		if buf.Len() == 0 {
			fmt.Print("bombo> ")
		} else {
			fmt.Print("  ...> ")
		}
		line, err := in.ReadString('\n')
		if err != nil && line == "" {
			if err == io.EOF {
				fmt.Println("\nLikkle more!")
			}
			return
		}
		if buf.Len() == 0 && strings.TrimSpace(line) == "help" {
			fmt.Print(helpText)
			continue
		}
		buf.WriteString(line)
		if openDepth(buf.String()) > 0 {
			continue
		}
		src := buf.String()
		buf.Reset()
		if strings.TrimSpace(src) == "" {
			continue
		}
		prog, err := parser.Parse(src)
		if err != nil {
			report(err)
			continue
		}
		v, err := it.RunREPL(prog)
		if errors.Is(err, interpreter.ErrExit) {
			fmt.Println("Likkle more!")
			return
		}
		if err != nil {
			report(err)
			continue
		}
		if v != nil {
			fmt.Println(interpreter.Repr(v))
		}
	}
}

// openDepth counts unclosed brackets so the REPL can accept multi-line blocks.
func openDepth(src string) int {
	depth := 0
	var quote rune
	escaped := false
	for _, c := range src {
		switch {
		case quote != 0:
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '{' || c == '(' || c == '[':
			depth++
		case c == '}' || c == ')' || c == ']':
			depth--
		}
	}
	return depth
}

const helpText = `Bombo ` + version + ` - a small interpreted programming language

USAGE
    bombo <file>          Run a Bombo program
    bombo                 Start the interactive REPL
    bombo help            Show this help
    bombo version         Show the version

VARIABLES
    mek x = 5             Declare a variable
    x = 6                 Assign
    x += 1    x -= 1      Compound assignment
    big up x              Increment (x = x + 1)
    bun x                 Decrement (x = x - 1)

OUTPUT
    seh "hello", x        Print values separated by spaces

CONTROL FLOW
    bombo cond { ... }              if
    bomboclaat cond { ... }         else if
    bloodclaat { ... }              else
    gwaan cond { ... }              while
    fi i from 1 tuh 10 { ... }      Numeric loop (inclusive range)
    fi item inna list { ... }       Loop over a list, string or map keys
    rassclaat                       break
    gallang                         continue
    likkle more                     Exit the program

FUNCTIONS
    wuk add(a, b) {                 Define a function
        sen back a + b              Return a value
    }
    add(1, 2)                       Call a function

ERRORS
    tek a chance { ... }            try
    if it bruk e { ... }            catch (e holds the error)
    bruk "message"                  Raise an error

TYPES AND LITERALS
    42   3.14                       Number
    "text"   'text'                 String
    yeh mon   nah mon               Boolean (true / false)
    kakafaat                        nil
    [1, 2, 3]                       List
    {name: "Ann", age: 30}          Map

OPERATORS
    +  -  *  /  %                   Arithmetic (+ also joins strings and lists)
    ==  !=  <  <=  >  >=            Comparison
    an   or   nuh                   Logical and / or / not
    xs[0]   m["key"]                Indexing

COMMENTS
    # comment
    watch yah comment

BUILT-IN FUNCTIONS
    howmuch(x)        len         Length of a string, list or map
    pon(list, x...)   push        Append values to a list
    tekoff(list)      pop         Remove and return the last element
    ax(prompt)        input       Read a line from standard input
    wahkind(x)        type        Name of a value's type
    numba(x)          num         Convert to a number
    wud(x)            str         Convert to a string
    luck(lo, hi)      random      Random integer between lo and hi
    dash(c, key)      remove      Remove a map key or list index; returns the value
    cool(ms)          sleep       Pause for ms milliseconds
    keys(map)                     List of a map's keys
    has(c, x)                     Whether a map, list or string contains x
    sort(list)                    Sorted copy of a list
    split(s, sep)                 Split a string into a list
    join(list, sep)               Join a list into a string
    upper(s)   lower(s)           Change string case
    floor  ceil  round  abs  sqrt Math functions
    time()                        Current time in seconds
`
