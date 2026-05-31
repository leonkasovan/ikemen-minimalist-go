package main

import (
	"fmt"
	"io"
	"sort"
)

type MissingAIRRef struct {
	Group  int
	Number int
}

type RenderDiagnostics struct {
	SFFVersion          string
	SpriteHeaderCount   int
	DecodedSpriteCount  int
	PaletteCount        int
	LinkedSpriteCount   int
	AIRActionCount      int
	DrawableActionCount int
	MissingAIRRefs      []MissingAIRRef
	DrawableActions     map[int]bool
}

// BuildRenderDiagnostics verifies that the loaded SFF sprite table and AIR
// animation references agree before the renderer starts. This keeps loader
// parity checks separate from palette remapping and the indexed OpenGL path.
func BuildRenderDiagnostics(sff *SFF, animations []*Animation) RenderDiagnostics {
	d := RenderDiagnostics{
		SFFVersion:         sff.Version,
		SpriteHeaderCount:  sff.SpriteHeaderCount,
		DecodedSpriteCount: len(sff.Sprites),
		PaletteCount:       len(sff.Palettes),
		LinkedSpriteCount:  sff.LinkedSpriteCount,
		AIRActionCount:     len(animations),
		DrawableActions:    map[int]bool{},
	}

	missing := map[[2]int]bool{}
	for _, anim := range animations {
		drawable := false
		for _, frame := range anim.Frames {
			if frame.Group < 0 || frame.Number < 0 {
				continue
			}
			key := [2]uint16{uint16(frame.Group), uint16(frame.Number)}
			if sff.Sprites[key] != nil {
				drawable = true
			} else {
				missing[[2]int{frame.Group, frame.Number}] = true
			}
		}
		if drawable {
			d.DrawableActions[anim.No] = true
		}
	}
	d.DrawableActionCount = len(d.DrawableActions)

	for key := range missing {
		d.MissingAIRRefs = append(d.MissingAIRRefs, MissingAIRRef{Group: key[0], Number: key[1]})
	}
	sort.Slice(d.MissingAIRRefs, func(i, j int) bool {
		if d.MissingAIRRefs[i].Group != d.MissingAIRRefs[j].Group {
			return d.MissingAIRRefs[i].Group < d.MissingAIRRefs[j].Group
		}
		return d.MissingAIRRefs[i].Number < d.MissingAIRRefs[j].Number
	})

	return d
}

func (d RenderDiagnostics) Print(w io.Writer) {
	fmt.Fprintln(w, "Ikemen render parity diagnostics:")
	fmt.Fprintf(w, "  SFF version: %s\n", d.SFFVersion)
	fmt.Fprintf(w, "  sprite header count: %d\n", d.SpriteHeaderCount)
	fmt.Fprintf(w, "  decoded sprite count: %d\n", d.DecodedSpriteCount)
	fmt.Fprintf(w, "  linked sprite count: %d\n", d.LinkedSpriteCount)
	fmt.Fprintf(w, "  palette count: %d\n", d.PaletteCount)
	fmt.Fprintf(w, "  AIR action count: %d\n", d.AIRActionCount)
	fmt.Fprintf(w, "  drawable action count: %d\n", d.DrawableActionCount)
	fmt.Fprintf(w, "  action 5100 drawable: %v\n", d.DrawableActions[5100])
	fmt.Fprintf(w, "  action 5170 drawable: %v\n", d.DrawableActions[5170])

	if len(d.MissingAIRRefs) == 0 {
		fmt.Fprintln(w, "  missing AIR refs: none")
		return
	}

	fmt.Fprintf(w, "  missing AIR refs: %d\n", len(d.MissingAIRRefs))
	limit := len(d.MissingAIRRefs)
	if limit > 20 {
		limit = 20
	}
	for i := 0; i < limit; i++ {
		ref := d.MissingAIRRefs[i]
		fmt.Fprintf(w, "    %d,%d\n", ref.Group, ref.Number)
	}
	if len(d.MissingAIRRefs) > limit {
		fmt.Fprintf(w, "    ... %d more\n", len(d.MissingAIRRefs)-limit)
	}
}

// ValidateBundledKFM enforces the explicit regression target for the bundled
// kfm.sff/kfm.air pair. It is intentionally called only for that pair so other
// characters can still be inspected without KFM-specific assumptions.
func (d RenderDiagnostics) ValidateBundledKFM() error {
	if d.DrawableActionCount != 117 {
		return fmt.Errorf("bundled KFM drawable action count = %d, want 117", d.DrawableActionCount)
	}
	if !d.DrawableActions[5100] {
		return fmt.Errorf("bundled KFM action 5100 is not drawable")
	}
	if !d.DrawableActions[5170] {
		return fmt.Errorf("bundled KFM action 5170 is not drawable")
	}
	if len(d.MissingAIRRefs) != 0 {
		return fmt.Errorf("bundled KFM has %d missing AIR sprite refs", len(d.MissingAIRRefs))
	}
	return nil
}
