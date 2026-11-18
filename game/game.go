// Package game is Bombo's 2D game library: a window, images, shapes, text,
// keyboard input, sounds and saving. Bombo code loads it with
//
//	tek "game"
//
// and starts it with game.run(update, draw).
package game

import (
	"errors"

	"github.com/RajaeDev6/bombo/interpreter"
	"github.com/hajimehoshi/ebiten/v2"
)

const tps = 60

func init() {
	interpreter.RegisterModule("game", load)
}

// lib is the state behind one `tek "game"`.
type lib struct {
	it *interpreter.Interpreter

	w, h  int
	title string

	screen     *ebiten.Image // only set while draw() runs
	camX, camY float64

	volume float64

	update, draw interpreter.Value
	running      bool
	ran          bool
	quit         bool
	err          error
}

func load(it *interpreter.Interpreter) (*interpreter.Map, error) {
	l := &lib{it: it, w: 640, h: 480, title: "Bombo", volume: 1}
	return l.module(), nil
}

// ebitenGame adapts lib to Ebitengine's game loop.
type ebitenGame struct{ l *lib }

func (g ebitenGame) Update() error {
	l := g.l
	if l.quit || l.err != nil {
		return ebiten.Termination
	}
	if l.update != nil {
		if _, err := l.it.Call(l.update, []interpreter.Value{1.0 / tps}, 0); err != nil {
			l.err = err
			return ebiten.Termination
		}
	}
	return nil
}

func (g ebitenGame) Draw(screen *ebiten.Image) {
	l := g.l
	if l.err != nil || l.draw == nil {
		return
	}
	l.screen = screen
	l.camX, l.camY = 0, 0
	if _, err := l.it.Call(l.draw, nil, 0); err != nil {
		l.err = err
	}
	l.screen = nil
}

func (g ebitenGame) Layout(_, _ int) (int, int) { return g.l.w, g.l.h }

// run(update, draw) opens the window and calls update(dt) 60 times a second
// and draw() every frame, until the window closes or game.quit() is called.
// Either function can be kakafaat.
func (l *lib) run(args []V, line int) (V, error) {
	if len(args) != 2 {
		return nil, interpreter.Errorf(line, "'game.run' want two functions: game.run(update, draw)")
	}
	for i, a := range args {
		switch a.(type) {
		case nil, *interpreter.Function, *interpreter.Builtin:
		default:
			return nil, interpreter.Errorf(line, "'game.run' want functions, but ting #%d a %s", i+1, interpreter.TypeName(a))
		}
	}
	if l.ran {
		return nil, interpreter.Errorf(line, "'game.run' can only run once per program")
	}
	l.update, l.draw = args[0], args[1]
	l.ran, l.running = true, true

	ebiten.SetWindowSize(l.w, l.h)
	ebiten.SetWindowTitle(l.title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(tps)
	err := ebiten.RunGame(ebitenGame{l})
	l.running = false

	if l.err != nil {
		return nil, l.err
	}
	if err != nil && !errors.Is(err, ebiten.Termination) {
		return nil, interpreter.Errorf(line, "Di game window bruk: %s", err)
	}
	return nil, nil
}
