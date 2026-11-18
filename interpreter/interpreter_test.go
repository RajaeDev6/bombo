package interpreter_test

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RajaeDev6/bombo/interpreter"
	"github.com/RajaeDev6/bombo/parser"
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
		{"dot access", "mek p = {x: 1, pos: {y: 2}}\np.x += 4\nbig up p.pos.y\np.z = p.x + p.pos.y\nseh p.x, p.pos.y, p.z, p.nope", "5 3 8 kakafaat\n"},
		{"block inside call", "wuk twice(f) { f()\n f() }\ntwice(wuk() {\n mek a = 1\n seh a\n})", "1\n1\n"},
		{"multiline map in call", "seh howmuch({\n a: 1,\n b: [\n 2,\n 3\n ]\n})", "2\n"},
		{"dot call", "mek m = {f: wuk(a) { sen back a * 2 }}\nseh m.f(21)", "42\n"},
		{"float after dot", "mek m = {a: 1.5}\nseh m.a + 0.5", "2\n"},
		{"math", "seh min(3, 1, 2), max([4, 9, 2]), round(sin(pi / 2)), atan2(0, 1)", "1 9 1 0\n"},
		{"tek not a phrase", "tek a chance { bruk 1 } if it bruk { seh \"ok\" }", "ok\n"},
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

func TestImportFile(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("shapes.bombo", "mek sides = 4\nwuk area(w, h) { sen back w * h }\nseh \"loaded\"")
	write("main.bombo", "tek \"shapes\"\ntek \"shapes.bombo\" as s\nseh shapes.sides, s.area(2, 3), keys(shapes)")
	write("loop_a.bombo", "tek \"loop_b\"")
	write("loop_b.bombo", "tek \"loop_a\"")
	write("broken.bombo", "mek a = 1\nseh a / 0")

	runDir := func(src string) (string, error) {
		prog, err := parser.Parse(src)
		if err != nil {
			return "", err
		}
		var out bytes.Buffer
		it := interpreter.New(&out, bufio.NewReader(strings.NewReader("")))
		it.Dir = dir
		err = it.Run(prog)
		return out.String(), err
	}

	got, err := runDir(`tek "main"`)
	if err != nil {
		t.Fatal(err)
	}
	if want := "loaded\n4 6 [\"area\", \"sides\"]\n"; got != want {
		t.Fatalf("got %q, want %q (a module should only run once)", got, want)
	}

	for src, want := range map[string]string{
		`tek "loop_a"`:  "in a circle",
		`tek "nothing"`: "Mi cyaan find a module name 'nothing'",
		`tek "broken"`:  "Inna 'broken' line 2: Yuh cyaan divide by zero",
	} {
		if _, err := runDir(src); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got error %v, want it to contain %q", src, err, want)
		}
	}
}

func TestGoLibrary(t *testing.T) {
	interpreter.RegisterModule("testlib", func(it *interpreter.Interpreter) (*interpreter.Map, error) {
		m := interpreter.NewMap()
		m.Set("double", interpreter.NewBuiltin("double", func(it *interpreter.Interpreter, args []interpreter.Value, line int) (interpreter.Value, error) {
			n, ok := args[0].(float64)
			if !ok {
				return nil, interpreter.Errorf(line, "double want a numba")
			}
			return n * 2, nil
		}))
		return m, nil
	})
	got, err := run(t, "tek \"testlib\"\nseh testlib.double(21)")
	if err != nil || got != "42\n" {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := run(t, "tek \"testlib\"\ntestlib.double(\"x\")"); err == nil || !strings.Contains(err.Error(), "Line 2: double want a numba") {
		t.Fatalf("got error %v", err)
	}
}
