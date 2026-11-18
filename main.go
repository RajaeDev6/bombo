// Command bombo runs Bombo programs.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"bombo/interpreter"
	"bombo/parser"
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

const helpText = `
BOMBO: di Jamaican programming language

  usage: bombo <file.bombo>     run a program
         bombo                  start di REPL

  KEYWORDS
    mek x = 5                   declare a variable
    x = 6   x += 1   x -= 1     change it
    big up x  /  bun x          x + 1  /  x - 1
    seh "hello", x              print
    bombo cond { }              if
    bomboclaat cond { }         else if
    bloodclaat { }              else
    gwaan cond { }              while loop
    fi i from 1 tuh 10 { }      counting loop (inclusive)
    fi x inna list { }          for-each loop
    rassclaat                   break
    gallang                     continue
    wuk name(a, b) { }          define a function
    sen back x                  return
    tek a chance { }            try
    if it bruk e { }            catch
    bruk "message"              throw an error
    likkle more                 exit di program
    wah gwaan                   optional greeting (does nothing)
    watch yah ...  /  # ...     comment

  VALUES
    yeh mon / nah mon           true / false
    kakafaat                    nil / nothing
    an  or  nuh                 and  or  not
    [1, 2, 3]   {name: "Bob"}   lists and maps

  BUILT-IN FUNCTIONS
    howmuch(x)  pon(list, x)  tekoff(list)  ax("question? ")
    wahkind(x)  numba(x)  wud(x)  luck(lo, hi)  keys(map)
    has(c, x)  dash(c, key)  sort(list)  split(s, sep)  join(list, sep)
    upper(s)  lower(s)  floor ceil round abs sqrt  cool(ms)  time()
`
