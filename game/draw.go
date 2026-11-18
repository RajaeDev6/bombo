package game

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Shapes are drawn from a white pixel and a pre-rendered circle so that
// Ebitengine can batch hundreds of them into a few GPU draw calls.

const circleTex = 256

var (
	whitePixel *ebiten.Image
	circleImg  *ebiten.Image
)

func initShapes() {
	if whitePixel != nil {
		return
	}
	base := ebiten.NewImage(3, 3)
	base.Fill(color.White)
	whitePixel = base.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

	circleImg = ebiten.NewImage(circleTex+2, circleTex+2)
	vector.FillCircle(circleImg, circleTex/2+1, circleTex/2+1, circleTex/2, color.White, true)
}

func scaleOf(c color.Color) ebiten.ColorScale {
	var cs ebiten.ColorScale
	cs.ScaleWithColor(c)
	return cs
}

func fillRect(dst *ebiten.Image, x, y, w, h float32, c color.Color) {
	initShapes()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w), float64(h))
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale = scaleOf(c)
	dst.DrawImage(whitePixel, op)
}

func fillCircle(dst *ebiten.Image, x, y, r float32, c color.Color) {
	if r <= 0 {
		return
	}
	initShapes()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	s := float64(r) * 2 / circleTex
	op.GeoM.Translate(-1, -1)
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(float64(x-r), float64(y-r))
	op.ColorScale = scaleOf(c)
	dst.DrawImage(circleImg, op)
}

// polys draws triangles given as flat x,y coordinates, three points each.
func polys(dst *ebiten.Image, pts []float32, c color.Color) {
	initShapes()
	r, g, b, a := c.RGBA()
	cr, cg, cb, ca := float32(r)/0xffff, float32(g)/0xffff, float32(b)/0xffff, float32(a)/0xffff
	vs := make([]ebiten.Vertex, len(pts)/2)
	is := make([]uint16, len(vs))
	for i := range vs {
		vs[i] = ebiten.Vertex{DstX: pts[i*2], DstY: pts[i*2+1], SrcX: 1.5, SrcY: 1.5, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca}
		is[i] = uint16(i)
	}
	op := &ebiten.DrawTrianglesOptions{ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha}
	dst.DrawTriangles(vs, is, whitePixel, op)
}

func fillTriangle(dst *ebiten.Image, n []float32, c color.Color) {
	polys(dst, n[:6], c)
}

func strokeLine(dst *ebiten.Image, x0, y0, x1, y1, w float32, c color.Color) {
	dx, dy := x1-x0, y1-y0
	l := float32(math.Hypot(float64(dx), float64(dy)))
	if l == 0 {
		return
	}
	nx, ny := -dy/l*w/2, dx/l*w/2
	polys(dst, []float32{
		x0 + nx, y0 + ny, x1 + nx, y1 + ny, x1 - nx, y1 - ny,
		x0 + nx, y0 + ny, x1 - nx, y1 - ny, x0 - nx, y0 - ny,
	}, c)
}

func strokeRect(dst *ebiten.Image, x, y, w, h, t float32, c color.Color) {
	fillRect(dst, x, y, w, t, c)
	fillRect(dst, x, y+h-t, w, t, c)
	fillRect(dst, x, y+t, t, h-2*t, c)
	fillRect(dst, x+w-t, y+t, t, h-2*t, c)
}

func strokeCircle(dst *ebiten.Image, x, y, r, t float32, c color.Color) {
	segs := int(math.Max(20, math.Min(96, float64(r))))
	ri, ro := r-t/2, r+t/2
	pts := make([]float32, 0, segs*12)
	for i := 0; i < segs; i++ {
		a0 := 2 * math.Pi * float64(i) / float64(segs)
		a1 := 2 * math.Pi * float64(i+1) / float64(segs)
		c0, s0 := float32(math.Cos(a0)), float32(math.Sin(a0))
		c1, s1 := float32(math.Cos(a1)), float32(math.Sin(a1))
		pts = append(pts,
			x+c0*ri, y+s0*ri, x+c0*ro, y+s0*ro, x+c1*ro, y+s1*ro,
			x+c0*ri, y+s0*ri, x+c1*ro, y+s1*ro, x+c1*ri, y+s1*ri,
		)
	}
	polys(dst, pts, c)
}
