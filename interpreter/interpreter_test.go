package interpreter_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"bombo/interpreter"
	"bombo/parser"
)

func run(t *testing.T, src string) (string, error) {
	t.Helper()
	prog, err := parser.Parse(src)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	it := interpreter.New(&out, bufio.NewReader(strings.NewReader("")))
	err = it.Run(prog)
	return out.String(), err
}

func TestPrograms(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"print", `seh "wah gwaan", 1 + 2`, "wah gwaan 3\n"},
		{"vars", "mek x = 2\nx += 3\nbig up x\nbun x\nbun x\nseh x", "4\n"},
		{"if chain", "mek n = 5\nbombo n > 10 { seh \"big\" } bomboclaat n > 3 { seh \"medium\" } bloodclaat { seh \"small\" }", "medium\n"},
		{"else on new line", "bombo nah mon {\n seh 1\n}\nbloodclaat {\n seh 2\n}", "2\n"},
		{"else bombo", "bombo nah mon { seh 1 } bloodclaat bombo yeh mon { seh 2 }", "2\n"},
		{"while break continue", "mek i = 0\ngwaan i < 10 {\n big up i\n bombo i % 2 == 0 { gallang }\n bombo i > 6 { rassclaat }\n seh i\n}", "1\n3\n5\n"},
		{"range down", "fi i from 3 tuh 1 { seh i }", "3\n2\n1\n"},
		{"for in string", `fi c inna "abc" { seh c }`, "a\nb\nc\n"},
		{"recursion", "wuk f(n) { bombo n <= 1 { sen back 1 }\n sen back n * f(n - 1) }\nseh f(5)", "120\n"},
		{"closure", "wuk mk() { mek c = 0\n sen back wuk() { big up c\n sen back c } }\nmek n = mk()\nn()\nseh n()", "2\n"},
		{"bools nil", "seh yeh mon, nah mon, kakafaat, nuh kakafaat", "yeh mon nah mon kakafaat yeh mon\n"},
		{"logic", `seh kakafaat or "x", 1 an 2, 0 an 2`, "x 2 0\n"},
		{"lists", "mek a = [1, 2]\npon(a, 3)\na[0] = 9\nseh a, howmuch(a), a[-1], tekoff(a), a", "[9, 2, 3] 3 3 3 [9, 2]\n"},
		{"maps", "mek m = {a: 1, \"b c\": 2}\nm[\"d\"] = 3\nseh m, m[\"zz\"], has(m, \"a\")", "{\"a\": 1, \"b c\": 2, \"d\": 3} kakafaat yeh mon\n"},
		{"multiline list", "mek a = [\n 1,\n 2,\n]\nseh a", "[1, 2]\n"},
		{"try catch", "tek a chance { bruk \"oops\" } if it bruk e { seh \"caught\", e }", "caught oops\n"},
		{"runtime error caught", "tek a chance { seh 1 / 0 } if it bruk { seh \"div\" }", "div\n"},
		{"comments", "watch yah dis a comment\nseh 1 # another\nseh 2", "1\n2\n"},
		{"floats", "seh 7 / 2, 0.1 + 0.2 == 0.3, 10 % 3", "3.5 nah mon 1\n"},
		{"strings", `seh "ha" * 3, "a" + 1, upper("yow"), join(split("a-b", "-"), "+")`, "hahaha a1 YOW a+b\n"},
		{"exit", "seh 1\nlikkle more\nseh 2", "1\n"},
		{"identifier near phrase", "mek big = 3\nmek watch = 4\nseh big + watch", "7\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := run(t, c.src)
			if err != nil && err != interpreter.ErrExit {
				t.Fatalf("error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"undefined", "seh x", "Mi cyaan find"},
		{"assign undefined", "x = 1", "Mi cyaan find 'x'. Use"},
		{"div zero", "seh 1 / 0", "divide by zero"},
		{"arity", "wuk f(a) { }\nf(1, 2)", "want 1 ting but yuh gi it 2"},
		{"break outside", "rassclaat", "only work inside a loop"},
		{"unclosed block", "bombo yeh mon {", "never close it"},
		{"stray else", "bloodclaat { }", "need a 'bombo' first"},
		{"bad char", "seh 1 @ 2", "Mi nuh know dis character"},
		{"index", "mek a = []\nseh a[0]", "outta bounds"},
		{"call non fn", "mek a = 1\na()", "not a function"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := run(t, c.src)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got error %v, want it to contain %q", err, c.want)
			}
		})
	}
}
