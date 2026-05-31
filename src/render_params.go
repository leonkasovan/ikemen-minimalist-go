package main

type TransType int

const (
	TransAlpha TransType = iota
	TransAdd
	TransSub
	TransNone
)

type PaletteEntry struct {
	Group uint16
	Item  uint16
	RGBA  []byte // 256*4 RGBA bytes
}

type PalRemap map[[2]uint16][2]uint16

type RenderParams struct {
	Sprite         *Sprite
	Palette        *PaletteEntry
	X, Y           float64
	ScaleX, ScaleY float64
	Angle          float64
	FlipX, FlipY   bool
	Tint           [4]float32
	BlendMode      TransType
	Alpha          [2]int32

	// Mask is reserved for Ikemen-style sprite masking. A value of 0 keeps
	// current behavior unchanged; positive values discard that palette index
	// for indexed sprites.
	Mask int32

	// Window is an optional scissor rectangle in top-left screen coordinates:
	// [x, y, width, height]. Nil disables clipping.
	Window *[4]int32

	// RotCenter is an optional rotation center relative to the sprite top-left.
	// The zero value keeps the previous center-of-quad rotation behavior.
	RotCenter [2]float64

	Layer     int
	Priority  int
	SortIndex int
}
