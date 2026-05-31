package main

type TransType int

const (
	TransAlpha TransType = iota
	TransAdd
	TransSub
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
	Layer          int
	Priority       int
	SortIndex      int
}
