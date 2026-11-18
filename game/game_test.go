package game_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	_ "github.com/RajaeDev6/bombo/game"
	"github.com/RajaeDev6/bombo/interpreter"
	"github.com/RajaeDev6/bombo/parser"
)

func run(t *testing.T, src string) (string, error) {
	t.Helper()
	prog, err := parser.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	it := interpreter.New(&out, bufio.NewReader(strings.NewReader("")))
	err = it.Run(prog)
	return out.String(), err
}

func TestModule(t *testing.T) {
	got, err := run(t, `tek "game"
game.window(320, 200, "Test")
seh game.width(), game.height(), game.rgb(1, 2, 3), game.held("space"), game.tapped("a")`)
	if err != nil {
		t.Fatal(err)
	}
	if want := "320 200 [1, 2, 3] nah mon nah mon\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestErrors(t *testing.T) {
	cases := map[string]string{
		`game.rect(0, 0, 10, 10, "red")`: "only work inside di draw function",
		`game.held("nope")`:              "Mi nuh know di key 'nope'",
		`game.sound("moo.wav")`:          "Mi cyaan find di sound 'moo.wav'",
		`game.image("nope.png")`:         "Mi cyaan find di picture 'nope.png'",
		`game.play(1)`:                   "want a sound",
		`game.draw(1, 2, 3)`:             "only work inside di draw function",
		`game.window(10, 10, "x")`:       "too likkle",
		`game.run(1, 2)`:                 "want functions",
		`game.save("x", [1])`:            "can only keep",
		`game.camera("a", 1)`:            "want a numba",
	}
	for src, want := range cases {
		_, err := run(t, "tek \"game\"\n"+src)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got error %v, want it to contain %q", src, err, want)
		}
	}
}
