package display

import (
	//	"log"
	"math"

	"github.com/bluepeppers/allegro"
	"github.com/go-gl/gl"

	"github.com/bluepeppers/danckelmann/resources"
)

var ISOMETRIC_ROTATION = float32(3 * math.Pi / 8)

type Viewport struct {
	X, Y, W, H   int
	XZoom, YZoom float64
}

func (v *Viewport) Move(dx, dy int) {
	v.x += dx
	v.y += dy
}

func (v *Viewport) OnScreen(x, y, w, h int) bool {
	return !(
	// Off left side
	x+w < v.x ||
		// Off right side
		x > v.x+int(float64(v.w)*v.xZoom) ||
		// Off top
		y+h < v.y ||
		// Off bottom
		y > v.y+int(float64(v.h)*v.yZoom))
}

func (v *Viewport) TileCoordinatesToScreen(tx, ty float64, config DisplayConfig) (float64, float64) {
	var trans allegro.Transform
	trans.Build(float32(-v.x), float32(-v.y), float32(v.xZoom), float32(v.yZoom),
		ISOMETRIC_ROTATION)
	x, y := trans.Apply(float32(tx), float32(ty))
	return float64(x), float64(y)
}

func (v *Viewport) ScreenCoordinatesToTile(sx, sy int, config DisplayConfig) (float64, float64) {
	// I have never been more ashamed of code I have written
	w, h := float64(config.TileW), float64(config.TileH)
	fx, fy := float32(sx), float32(sy)
	var trans allegro.Transform
	trans.Identity()
	// Builds the viewport alignment matrix
	trans.Build(float32(-v.x), float32(-v.y), float32(v.xZoom), float32(v.yZoom),
		0)
	// Invert it to get back to pixel coordinates
	trans.Invert()
	// We need to translate back half a width to get to the pivot of the tiles
	trans.Translate(-float32(w/2), 0)

	x, y := trans.Apply(fx, fy)
	// Then we manually rotate it (because I'm bad at maths I guess)
	tx := float64(float64(y)*w-float64(x)*h) / (w * h)
	ty := float64(float64(y)*w+float64(x)*h) / (w * h)
	return tx, ty
}
