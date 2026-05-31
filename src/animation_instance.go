package main

import (
	"math"
	"strconv"
)

type AnimationInstance struct {
	Anim            *Animation
	SFF             *SFF
	X, Y            float64
	ScaleX, ScaleY  float64
	Facing          float64
	Layer           int
	Priority        int
	BlendMode       TransType
	Alpha           [2]int32
	Tint            [4]float32
	PalRemap        PalRemap
	PaletteOverride int // -1 means use default/remap

	DebugClip      bool
	DebugMask      int32
	DebugRotCenter bool
	DebugAngle     float64
}

func NewAnimationInstance(sff *SFF, anim *Animation) *AnimationInstance {
	palOverride := -1
	if idx, ok := sff.PaletteByKey[[2]uint16{1, 1}]; ok {
		palOverride = idx
	}
	return &AnimationInstance{
		SFF:             sff,
		Anim:            anim,
		X:               320,
		Y:               360,
		ScaleX:          1,
		ScaleY:          1,
		Facing:          1,
		BlendMode:       TransAlpha,
		Alpha:           [2]int32{255, 0},
		Tint:            [4]float32{1, 1, 1, 1},
		PaletteOverride: palOverride,
	}
}

func (ai *AnimationInstance) SetAnim(anim *Animation) {
	ai.Anim = anim
	if ai.Anim != nil {
		ai.Anim.Reset()
	}
}

func (ai *AnimationInstance) Update() {
	if ai.Anim != nil {
		ai.Anim.Step()
	}
}

func (ai *AnimationInstance) CurrentSprite() *Sprite {
	if ai == nil || ai.Anim == nil {
		return nil
	}
	fr := ai.Anim.CurrentFrame()
	return ai.SFF.Sprites[[2]uint16{uint16(fr.Group), uint16(fr.Number)}]
}

func (ai *AnimationInstance) ClearPalette() {
	ai.PaletteOverride = -1
	ai.PalRemap = nil
}

func (ai *AnimationInstance) CyclePaletteOverride(delta int) int {
	if len(ai.SFF.Palettes) == 0 {
		return -1
	}
	if ai.PaletteOverride < 0 {
		sp := ai.CurrentSprite()
		if sp != nil {
			ai.PaletteOverride = sp.PalIndex
		} else {
			ai.PaletteOverride = 0
		}
	}
	ai.PaletteOverride = (ai.PaletteOverride + delta + len(ai.SFF.Palettes)) % len(ai.SFF.Palettes)
	ai.PalRemap = nil
	return ai.PaletteOverride
}

func (ai *AnimationInstance) RemapCurrentDefault(delta int) (int, int) {
	sp := ai.CurrentSprite()
	if sp == nil || len(ai.SFF.Palettes) == 0 {
		return -1, -1
	}
	srcIdx := sp.PalIndex
	if srcIdx < 0 || srcIdx >= len(ai.SFF.Palettes) {
		srcIdx = 0
	}
	dstIdx := (srcIdx + delta + len(ai.SFF.Palettes)) % len(ai.SFF.Palettes)
	src := ai.SFF.Palettes[srcIdx]
	dst := ai.SFF.Palettes[dstIdx]
	if ai.PalRemap == nil {
		ai.PalRemap = PalRemap{}
	}
	ai.PalRemap[[2]uint16{src.Group, src.Item}] = [2]uint16{dst.Group, dst.Item}
	ai.PaletteOverride = -1
	return srcIdx, dstIdx
}

func (ai *AnimationInstance) PaletteName() string {
	if ai == nil || ai.SFF == nil {
		return "none"
	}
	sp := ai.CurrentSprite()
	pal := ai.SFF.ResolvePalette(sp, ai.PaletteOverride, ai.PalRemap)
	if pal == nil {
		return "none"
	}
	return fmtPal(pal)
}

func fmtPal(p *PaletteEntry) string {
	if p == nil {
		return "none"
	}
	return itoa2(p.Group) + "," + itoa2(p.Item)
}

func itoa2(v uint16) string {
	return strconv.Itoa(int(v))
}

func (ai *AnimationInstance) Draw(dl *DrawList) {
	if ai == nil || ai.Anim == nil || len(ai.Anim.Frames) == 0 {
		return
	}

	fr := ai.Anim.CurrentFrame()
	sp := ai.SFF.Sprites[[2]uint16{uint16(fr.Group), uint16(fr.Number)}]
	if sp == nil {
		return
	}

	sx := ai.ScaleX * fr.XScale
	sy := ai.ScaleY * fr.YScale
	if sx == 0 {
		sx = 1
	}
	if sy == 0 {
		sy = 1
	}
	if ai.Facing < 0 {
		sx = -sx
	}

	dw := math.Abs(float64(sp.W) * sx)
	dh := math.Abs(float64(sp.H) * sy)
	drawX := ai.X - float64(sp.XOff) + float64(fr.Xoff) - dw/2
	drawY := ai.Y - float64(sp.YOff) + float64(fr.Yoff) - dh

	angle := fr.Angle + ai.DebugAngle

	var window *[4]int32
	if ai.DebugClip {
		clip := [4]int32{160, 120, 320, 240}
		window = &clip
	}

	rotCenter := [2]float64{}
	if ai.DebugRotCenter {
		rotCenter = [2]float64{0, float64(sp.H)}
	}

	pal := ai.SFF.ResolvePalette(sp, ai.PaletteOverride, ai.PalRemap)
	dl.Add(RenderParams{
		Sprite:    sp,
		Palette:   pal,
		X:         drawX,
		Y:         drawY,
		ScaleX:    sx,
		ScaleY:    sy,
		Angle:     angle,
		FlipX:     fr.HFlip,
		FlipY:     fr.VFlip,
		Tint:      ai.Tint,
		BlendMode: ai.BlendMode,
		Alpha:     ai.Alpha,
		Mask:      ai.DebugMask,
		Window:    window,
		RotCenter: rotCenter,
		Layer:     ai.Layer,
		Priority:  ai.Priority,
	})
}
