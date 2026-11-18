package game

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/RajaeDev6/bombo/interpreter"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gomonobold"
)

type V = interpreter.Value

type fn func(args []V, line int) (V, error)

var errorf = interpreter.Errorf

// module builds the map that `tek "game"` gives to Bombo code.
func (l *lib) module() *interpreter.Map {
	m := interpreter.NewMap()
	add := func(name string, f fn) {
		m.Set(name, interpreter.NewBuiltin("game."+name, func(_ *interpreter.Interpreter, args []V, line int) (V, error) {
			return f(args, line)
		}))
	}

	add("window", l.window)
	add("run", l.run)
	add("quit", func([]V, int) (V, error) { l.quit = true; return nil, nil })
	add("width", func([]V, int) (V, error) { return float64(l.w), nil })
	add("height", func([]V, int) (V, error) { return float64(l.h), nil })

	add("clear", l.clear)
	add("camera", l.camera)
	add("rect", l.shape("rect", 5, func(s *ebiten.Image, n []float32, c color.Color) { fillRect(s, n[0], n[1], n[2], n[3], c) }))
	add("box", l.shape("box", 6, func(s *ebiten.Image, n []float32, c color.Color) { strokeRect(s, n[0], n[1], n[2], n[3], n[4], c) }))
	add("circle", l.shape("circle", 4, func(s *ebiten.Image, n []float32, c color.Color) { fillCircle(s, n[0], n[1], n[2], c) }))
	add("ring", l.shape("ring", 5, func(s *ebiten.Image, n []float32, c color.Color) { strokeCircle(s, n[0], n[1], n[2], n[3], c) }))
	add("line", l.shape("line", 6, func(s *ebiten.Image, n []float32, c color.Color) { strokeLine(s, n[0], n[1], n[2], n[3], n[4], c) }))
	add("triangle", l.shape("triangle", 7, func(s *ebiten.Image, n []float32, c color.Color) { fillTriangle(s, n, c) }))
	add("image", l.loadImage)
	add("draw", l.drawImage)
	add("text", l.text)
	add("rgb", rgbFn)
	add("screenshot", l.screenshot)

	add("held", l.key("held", ebiten.IsKeyPressed))
	add("tapped", l.key("tapped", inpututil.IsKeyJustPressed))

	add("sound", l.loadSound)
	add("play", l.play)
	add("volume", l.setVolume)

	add("save", l.save)
	add("load", l.loadValue)
	return m
}

func nums(name string, args []V, want int, line int) ([]float64, error) {
	if len(args) < want {
		return nil, errorf(line, "'game.%s' want %d numba dem but yuh gi it %d ting", name, want, len(args))
	}
	out := make([]float64, want)
	for i := 0; i < want; i++ {
		f, ok := args[i].(float64)
		if !ok {
			return nil, errorf(line, "'game.%s' want a numba fi ting #%d, but yuh gi it a %s", name, i+1, interpreter.TypeName(args[i]))
		}
		out[i] = f
	}
	return out, nil
}

func (l *lib) needScreen(name string, line int) error {
	if l.screen == nil {
		return errorf(line, "'game.%s' only work inside di draw function yuh give game.run", name)
	}
	return nil
}

// Which arguments of each shape are x positions; each is followed by its y.
var cameraShift = map[string][]int{
	"rect": {0}, "box": {0}, "circle": {0}, "ring": {0},
	"line": {0, 2}, "triangle": {0, 2, 4},
}

// shape makes a drawing function that takes n-1 numbers and then a colour.
func (l *lib) shape(name string, n int, draw func(*ebiten.Image, []float32, color.Color)) fn {
	return func(args []V, line int) (V, error) {
		if err := l.needScreen(name, line); err != nil {
			return nil, err
		}
		if len(args) != n {
			return nil, errorf(line, "'game.%s' want %d ting (numba dem, then a colour) but yuh gi it %d", name, n, len(args))
		}
		f, err := nums(name, args, n-1, line)
		if err != nil {
			return nil, err
		}
		c, err := parseColor(args[n-1], line)
		if err != nil {
			return nil, err
		}
		pts := make([]float32, len(f))
		for i, v := range f {
			pts[i] = float32(v)
		}
		for _, i := range cameraShift[name] {
			pts[i] -= float32(l.camX)
			pts[i+1] -= float32(l.camY)
		}
		draw(l.screen, pts, c)
		return nil, nil
	}
}

// window(w, h, title) sets the window size and title.
func (l *lib) window(args []V, line int) (V, error) {
	if len(args) != 3 {
		return nil, errorf(line, "'game.window' want width, height an a title: game.window(480, 720, \"Mi Game\")")
	}
	f, err := nums("window", args, 2, line)
	if err != nil {
		return nil, err
	}
	if f[0] < 50 || f[1] < 50 {
		return nil, errorf(line, "Di window too likkle, bredda")
	}
	l.w, l.h = int(f[0]), int(f[1])
	l.title = interpreter.ToString(args[2])
	if l.running {
		ebiten.SetWindowSize(l.w, l.h)
		ebiten.SetWindowTitle(l.title)
	}
	return nil, nil
}

func (l *lib) clear(args []V, line int) (V, error) {
	if err := l.needScreen("clear", line); err != nil {
		return nil, err
	}
	if len(args) != 1 {
		return nil, errorf(line, "'game.clear' want one colour")
	}
	c, err := parseColor(args[0], line)
	if err != nil {
		return nil, err
	}
	l.screen.Fill(c)
	return nil, nil
}

// camera(x, y) shifts everything drawn after it. It resets every frame.
func (l *lib) camera(args []V, line int) (V, error) {
	f, err := nums("camera", args, 2, line)
	if err != nil {
		return nil, err
	}
	l.camX, l.camY = f[0], f[1]
	return nil, nil
}

// text(words, x, y, size, colour) or text(words, x, y, size, colour, "center")
func (l *lib) text(args []V, line int) (V, error) {
	if err := l.needScreen("text", line); err != nil {
		return nil, err
	}
	if len(args) != 5 && len(args) != 6 {
		return nil, errorf(line, "'game.text' want: game.text(words, x, y, size, colour) an maybe \"center\" or \"right\"")
	}
	f, err := nums("text", args[1:], 3, line)
	if err != nil {
		return nil, err
	}
	c, err := parseColor(args[4], line)
	if err != nil {
		return nil, err
	}
	align := "left"
	if len(args) == 6 {
		align = interpreter.ToString(args[5])
	}
	drawText(l.screen, interpreter.ToString(args[0]), f[0]-l.camX, f[1]-l.camY, f[2], c, align)
	return nil, nil
}

var (
	fontSource *text.GoTextFaceSource
	faces      = map[float64]*text.GoTextFace{}
)

func drawText(dst *ebiten.Image, s string, x, y, size float64, c color.Color, align string) {
	if fontSource == nil {
		var err error
		if fontSource, err = text.NewGoTextFaceSource(bytes.NewReader(gomonobold.TTF)); err != nil {
			panic(err)
		}
	}
	face, ok := faces[size]
	if !ok {
		face = &text.GoTextFace{Source: fontSource, Size: size}
		faces[size] = face
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	op.LineSpacing = size * 1.3
	switch align {
	case "center", "middle":
		op.PrimaryAlign = text.AlignCenter
	case "right":
		op.PrimaryAlign = text.AlignEnd
	}
	text.Draw(dst, s, face, op)
}

// screenshot(path) saves the screen as it looks right now to a PNG file.
func (l *lib) screenshot(args []V, line int) (V, error) {
	if err := l.needScreen("screenshot", line); err != nil {
		return nil, err
	}
	if len(args) != 1 {
		return nil, errorf(line, "'game.screenshot' want a file name: game.screenshot(\"shot.png\")")
	}
	path := interpreter.ToString(args[0])
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.it.Dir, path)
	}
	img := image.NewRGBA(l.screen.Bounds())
	l.screen.ReadPixels(img.Pix)
	f, err := os.Create(path)
	if err == nil {
		err = png.Encode(f, img)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}
	if err != nil {
		return nil, errorf(line, "Mi cyaan save di screenshot: %s", err)
	}
	return nil, nil
}

func (l *lib) key(name string, check func(ebiten.Key) bool) fn {
	return func(args []V, line int) (V, error) {
		if len(args) != 1 {
			return nil, errorf(line, "'game.%s' want one key name, like game.%s(\"space\")", name, name)
		}
		s := interpreter.ToString(args[0])
		k, ok := keyByName(s)
		if !ok {
			return nil, errorf(line, "Mi nuh know di key '%s'", s)
		}
		return l.running && check(k), nil
	}
}

var namedKeys = map[string]ebiten.Key{
	"left": ebiten.KeyArrowLeft, "right": ebiten.KeyArrowRight,
	"up": ebiten.KeyArrowUp, "down": ebiten.KeyArrowDown,
	"space": ebiten.KeySpace, "enter": ebiten.KeyEnter, "escape": ebiten.KeyEscape,
	"esc": ebiten.KeyEscape, "shift": ebiten.KeyShiftLeft, "tab": ebiten.KeyTab,
	"backspace": ebiten.KeyBackspace, "ctrl": ebiten.KeyControlLeft,
}

func keyByName(s string) (ebiten.Key, bool) {
	s = strings.ToLower(s)
	if k, ok := namedKeys[s]; ok {
		return k, true
	}
	if len(s) == 1 {
		switch c := s[0]; {
		case c >= 'a' && c <= 'z':
			return ebiten.KeyA + ebiten.Key(c-'a'), true
		case c >= '0' && c <= '9':
			return ebiten.KeyDigit0 + ebiten.Key(c-'0'), true
		}
	}
	return 0, false
}

// volume(v) sets how loud all sounds are, from 0 (silent) to 1.
func (l *lib) setVolume(args []V, line int) (V, error) {
	f, err := nums("volume", args, 1, line)
	if err != nil {
		return nil, err
	}
	l.volume = math.Max(0, math.Min(1, f[0]))
	return nil, nil
}

// ---- colours ----

var namedColors = map[string]color.RGBA{
	"black": {0, 0, 0, 255}, "white": {255, 255, 255, 255},
	"red": {230, 40, 50, 255}, "green": {0, 155, 58, 255}, "gold": {254, 209, 0, 255},
	"yellow": {255, 230, 60, 255}, "blue": {50, 120, 255, 255}, "cyan": {60, 220, 255, 255},
	"orange": {255, 140, 30, 255}, "purple": {150, 70, 220, 255}, "pink": {255, 90, 170, 255},
	"grey": {128, 128, 128, 255}, "gray": {128, 128, 128, 255}, "clear": {0, 0, 0, 0},
}

// parseColor accepts a name ("gold"), hex ("#fed100", "#fed10080") or a
// list from rgb() like [254, 209, 0, 128].
func parseColor(v V, line int) (color.Color, error) {
	switch c := v.(type) {
	case string:
		if nc, ok := namedColors[strings.ToLower(c)]; ok {
			return nc, nil
		}
		if h := strings.TrimPrefix(c, "#"); h != c {
			if len(h) == 3 {
				h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
			}
			if len(h) == 6 {
				h += "ff"
			}
			if n, err := strconv.ParseUint(h, 16, 32); err == nil && len(h) == 8 {
				return premul(float64(n>>24), float64(n>>16&255), float64(n>>8&255), float64(n&255)), nil
			}
		}
	case *interpreter.List:
		if len(c.Elems) == 3 || len(c.Elems) == 4 {
			f := []float64{0, 0, 0, 255}
			ok := true
			for i, e := range c.Elems {
				n, isNum := e.(float64)
				ok = ok && isNum
				f[i] = n
			}
			if ok {
				return premul(f[0], f[1], f[2], f[3]), nil
			}
		}
	}
	return nil, errorf(line, "Dat a nuh colour: %s. Use a name like \"gold\", hex like \"#fed100\", or game.rgb(254, 209, 0)", interpreter.Repr(v))
}

func premul(r, g, b, a float64) color.RGBA {
	cl := func(x float64) float64 { return math.Max(0, math.Min(255, x)) }
	r, g, b, a = cl(r), cl(g), cl(b), cl(a)
	k := a / 255
	return color.RGBA{uint8(r * k), uint8(g * k), uint8(b * k), uint8(a)}
}

// rgb(r, g, b) or rgb(r, g, b, alpha), each 0-255. Returns a colour list.
func rgbFn(args []V, line int) (V, error) {
	if len(args) != 3 && len(args) != 4 {
		return nil, errorf(line, "'game.rgb' want three or four numba dem (0 tuh 255)")
	}
	f, err := nums("rgb", args, len(args), line)
	if err != nil {
		return nil, err
	}
	l := &interpreter.List{}
	for _, n := range f {
		l.Elems = append(l.Elems, n)
	}
	return l, nil
}

// ---- save / load ----

// savePath gives each game its own save file, named after its window title.
func (l *lib) savePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	var slug strings.Builder
	for _, r := range strings.ToLower(l.title) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			slug.WriteRune(r)
		case slug.Len() > 0 && !strings.HasSuffix(slug.String(), "-"):
			slug.WriteByte('-')
		}
	}
	name := strings.Trim(slug.String(), "-")
	if name == "" {
		name = "game"
	}
	return filepath.Join(dir, "bombo", name+".json")
}

func (l *lib) readSave() map[string]any {
	m := map[string]any{}
	if b, err := os.ReadFile(l.savePath()); err == nil {
		_ = json.Unmarshal(b, &m)
	}
	return m
}

// save(key, value) keeps a number, word or truth between runs.
func (l *lib) save(args []V, line int) (V, error) {
	if len(args) != 2 {
		return nil, errorf(line, "'game.save' want a key an a value: game.save(\"hiscore\", 100)")
	}
	switch args[1].(type) {
	case float64, string, bool, nil:
	default:
		return nil, errorf(line, "'game.save' can only keep a numba, wud or truth, not a %s", interpreter.TypeName(args[1]))
	}
	m := l.readSave()
	m[interpreter.ToString(args[0])] = args[1]
	b, _ := json.MarshalIndent(m, "", "  ")
	p := l.savePath()
	err := os.MkdirAll(filepath.Dir(p), 0o755)
	if err == nil {
		err = os.WriteFile(p, b, 0o644)
	}
	if err != nil {
		return nil, errorf(line, "Mi cyaan save: %s", err)
	}
	return nil, nil
}

// load(key) gives back what save kept, or kakafaat.
func (l *lib) loadValue(args []V, line int) (V, error) {
	if len(args) != 1 {
		return nil, errorf(line, "'game.load' want one key: game.load(\"hiscore\")")
	}
	switch v := l.readSave()[interpreter.ToString(args[0])].(type) {
	case float64, string, bool:
		return v, nil
	}
	return nil, nil
}
