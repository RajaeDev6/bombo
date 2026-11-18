package game

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/RajaeDev6/bombo/interpreter"
	"github.com/hajimehoshi/ebiten/v2"
)

// Image is a picture loaded with game.image(path).
type Image struct {
	path string
	img  *ebiten.Image
	w, h int
}

func (i *Image) TypeName() string { return "image" }
func (i *Image) String() string   { return "<image " + i.path + ">" }
func (i *Image) Field(name string) V {
	switch name {
	case "width":
		return float64(i.w)
	case "height":
		return float64(i.h)
	case "path":
		return i.path
	}
	return nil
}

// image(path) loads a .png or .jpg file.
func (l *lib) loadImage(args []V, line int) (V, error) {
	if len(args) != 1 {
		return nil, errorf(line, "'game.image' want a file name: game.image(\"ship.png\")")
	}
	path := interpreter.ToString(args[0])
	full := path
	if !filepath.IsAbs(full) {
		full = filepath.Join(l.it.Dir, full)
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, errorf(line, "Mi cyaan find di picture '%s'", path)
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, errorf(line, "Di picture '%s' look broken (mi can read .png an .jpg)", path)
	}
	b := src.Bounds()
	return &Image{path: path, img: ebiten.NewImageFromImage(src), w: b.Dx(), h: b.Dy()}, nil
}

// draw(img, x, y) draws an image with its top-left corner at x, y.
// draw(img, x, y, options) also takes a map with any of:
//
//	center  yeh mon to put the middle of the image at x, y
//	scale   how big to draw it (1 is normal size)
//	angle   rotation in radians, around the image's middle
//	alpha   see-through-ness, from 0 (invisible) to 1
//	tint    a colour to multiply the image by
//	flip    yeh mon to mirror it left to right
//	smooth  yeh mon for smooth scaling instead of crisp pixels
func (l *lib) drawImage(args []V, line int) (V, error) {
	if err := l.needScreen("draw", line); err != nil {
		return nil, err
	}
	if len(args) != 3 && len(args) != 4 {
		return nil, errorf(line, "'game.draw' want an image, x, y an maybe options: game.draw(ship, 10, 20, {scale: 2})")
	}
	im, ok := args[0].(*Image)
	if !ok {
		return nil, errorf(line, "'game.draw' want an image from game.image(...), not a %s", interpreter.TypeName(args[0]))
	}
	pos, err := nums("draw", args[1:], 2, line)
	if err != nil {
		return nil, err
	}

	center, flip, smooth := false, false, false
	scale, angle, alpha := 1.0, 0.0, 1.0
	var tint V
	if len(args) == 4 {
		opts, ok := args[3].(*interpreter.Map)
		if !ok {
			return nil, errorf(line, "'game.draw' options fi be a map, like {scale: 2, center: yeh mon}")
		}
		for _, k := range opts.Keys {
			v := opts.Vals[k]
			switch k {
			case "center", "flip", "smooth":
				b, ok := v.(bool)
				if !ok {
					return nil, errorf(line, "Di '%s' option fi be yeh mon or nah mon", k)
				}
				switch k {
				case "center":
					center = b
				case "flip":
					flip = b
				default:
					smooth = b
				}
			case "scale", "angle", "alpha":
				n, ok := v.(float64)
				if !ok {
					return nil, errorf(line, "Di '%s' option fi be a numba", k)
				}
				switch k {
				case "scale":
					scale = n
				case "angle":
					angle = n
				default:
					alpha = math.Max(0, math.Min(1, n))
				}
			case "tint":
				tint = v
			default:
				return nil, errorf(line, "'game.draw' nuh know di option '%s'. Use center, scale, angle, alpha, tint, flip or smooth", k)
			}
		}
	}

	op := &ebiten.DrawImageOptions{}
	if smooth {
		op.Filter = ebiten.FilterLinear
	}
	w, h := float64(im.w), float64(im.h)
	op.GeoM.Translate(-w/2, -h/2)
	if flip {
		op.GeoM.Scale(-1, 1)
	}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Rotate(angle)
	x, y := pos[0]-l.camX, pos[1]-l.camY
	if !center {
		x += w * scale / 2
		y += h * scale / 2
	}
	op.GeoM.Translate(x, y)
	if tint != nil {
		c, err := parseColor(tint, line)
		if err != nil {
			return nil, err
		}
		op.ColorScale.ScaleWithColor(c)
	}
	op.ColorScale.ScaleAlpha(float32(alpha))
	l.screen.DrawImage(im.img, op)
	return nil, nil
}
