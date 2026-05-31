package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

type Sprite struct {
	Group, Number uint16
	W, H          int
	XOff, YOff    int16
	Indexed       []byte
	IsIndexed     bool
	PalIndex      int
	RGBA          []byte
}

type SFF struct {
	Sprites      map[[2]uint16]*Sprite
	Palettes     []PaletteEntry
	PaletteByKey map[[2]uint16]int

	// PaletteIndexByHeader maps SFF v2 palette-header indices to Palettes entries.
	// Sprite headers reference palette header indices directly, so palette headers
	// are preserved even when their group/item keys are duplicates.
	PaletteIndexByHeader []int

	// Diagnostics captured while loading. These are intentionally simple
	// counters so the viewer can verify SFF/AIR parity without touching
	// the renderer, palette remap, or OpenGL indexed texture path.
	Version            string
	SpriteHeaderCount  int
	DecodedSpriteCount int
	LinkedSpriteCount  int

	DuplicateSpriteKeys   int
	DuplicatePaletteKeys  int
	InvalidLinkedSprites  int
	InvalidLinkedPalettes int
	DecodeErrorCount      int
	DecodeErrors          []string
}

type decodedSprite struct {
	indexed   []byte
	rgba      []byte
	isIndexed bool
	w, h      int
}

func u16(b []byte) uint16 { return binary.LittleEndian.Uint16(b) }
func u32(b []byte) uint32 { return binary.LittleEndian.Uint32(b) }

func newSFF() *SFF { return &SFF{Sprites: map[[2]uint16]*Sprite{}, PaletteByKey: map[[2]uint16]int{}} }
func (s *SFF) addPalette(group, item uint16, rgba []byte) int {
	// Preserve SFF v2 palette-header index semantics: sprite headers store a palette
	// header index, so every palette header must append one PaletteEntry. Duplicate
	// keys only affect key-based remap lookup, where original Ikemen keeps the
	// first key mapping.
	idx := len(s.Palettes)
	s.Palettes = append(s.Palettes, PaletteEntry{
		Group: group,
		Item:  item,
		RGBA:  ensurePalette(rgba),
	})

	key := [2]uint16{group, item}
	if _, ok := s.PaletteByKey[key]; ok {
		s.DuplicatePaletteKeys++
		return idx
	}

	s.PaletteByKey[key] = idx
	return idx
}

func (s *SFF) addSprite(sp *Sprite) bool {
	if sp == nil {
		return false
	}
	key := [2]uint16{sp.Group, sp.Number}
	if s.Sprites[key] != nil {
		// Match original Ikemen: keep the first sprite for duplicate keys.
		s.DuplicateSpriteKeys++
		return false
	}
	s.Sprites[key] = sp
	return true
}

func (s *SFF) recordDecodeError(format string, args ...any) {
	s.DecodeErrorCount++
	if len(s.DecodeErrors) < 20 {
		s.DecodeErrors = append(s.DecodeErrors, fmt.Sprintf(format, args...))
	}
}

func (s *SFF) paletteIndexFromHeader(headerIndex int) int {
	if headerIndex >= 0 && headerIndex < len(s.PaletteIndexByHeader) {
		idx := s.PaletteIndexByHeader[headerIndex]
		if idx >= 0 && idx < len(s.Palettes) {
			return idx
		}
	}
	if headerIndex >= 0 && headerIndex < len(s.Palettes) {
		return headerIndex
	}
	return 0
}

func (s *SFF) ResolvePalette(sp *Sprite, override int, remap PalRemap) *PaletteEntry {
	if sp == nil || !sp.IsIndexed || len(s.Palettes) == 0 {
		return nil
	}

	idx := sp.PalIndex
	if idx < 0 || idx >= len(s.Palettes) {
		idx = 0
	}
	pal := &s.Palettes[idx]

	if remap != nil {
		if dst, ok := remap[[2]uint16{pal.Group, pal.Item}]; ok {
			if dstIdx, ok := s.PaletteByKey[dst]; ok && dstIdx >= 0 && dstIdx < len(s.Palettes) {
				pal = &s.Palettes[dstIdx]
			}
		}
	}

	if override >= 0 && override < len(s.Palettes) {
		pal = &s.Palettes[override]
	}

	return pal
}

func cloneLinkedSprite(src *Sprite, group, number uint16, xoff, yoff int16, palIndex int) *Sprite {
	if src == nil {
		return nil
	}
	sp := &Sprite{
		Group:     group,
		Number:    number,
		W:         src.W,
		H:         src.H,
		XOff:      xoff,
		YOff:      yoff,
		IsIndexed: src.IsIndexed,
		PalIndex:  src.PalIndex,
	}
	if palIndex >= 0 {
		sp.PalIndex = palIndex
	}
	if src.Indexed != nil {
		sp.Indexed = append([]byte(nil), src.Indexed...)
	}
	if src.RGBA != nil {
		sp.RGBA = append([]byte(nil), src.RGBA...)
	}
	return sp
}

func LoadSFF(path string) (*SFF, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 64 || string(data[:12]) != "ElecbyteSpr\x00" {
		return nil, fmt.Errorf("not an SFF file")
	}
	s := newSFF()
	switch data[15] {
	case 1:
		return loadSFFv1(data, s)
	case 2:
		return loadSFFv2(data, s)
	default:
		return nil, fmt.Errorf("unsupported SFF version %d", data[15])
	}
}

func loadSFFv1(data []byte, s *SFF) (*SFF, error) {
	count, ofs := int(u32(data[20:24])), int(u32(data[24:28]))
	s.Version = "v1"
	s.SpriteHeaderCount = count

	var sharedPal []byte
	sharedPalIdx := -1
	spritesByIndex := make([]*Sprite, 0, count)

	for i := 0; i < count && ofs+32 <= len(data); i++ {
		next, size := int(u32(data[ofs:ofs+4])), int(u32(data[ofs+4:ofs+8]))
		xo, yo := int16(u16(data[ofs+8:ofs+10])), int16(u16(data[ofs+10:ofs+12]))
		grp, num := u16(data[ofs+12:ofs+14]), u16(data[ofs+14:ofs+16])
		link := int(u16(data[ofs+16 : ofs+18]))
		samePalette := ofs+18 < len(data) && data[ofs+18] != 0

		// SFF v1 linked sprites have their own group/number and axis but reuse
		// pixel data from a previous subfile. AIR can reference these group/number
		// pairs directly, so they must be present in SFF.Sprites.
		if size == 0 {
			if link >= 0 && link < len(spritesByIndex) && spritesByIndex[link] != nil {
				sp := cloneLinkedSprite(spritesByIndex[link], grp, num, xo, yo, -1)
				s.addSprite(sp)
				s.LinkedSpriteCount++
				spritesByIndex = append(spritesByIndex, sp)
			} else {
				s.InvalidLinkedSprites++
				spritesByIndex = append(spritesByIndex, nil)
			}
			if next == 0 {
				break
			}
			ofs = next
			continue
		}

		start := ofs + 32
		end := start + size
		// The next-subheader pointer is the safest block boundary for SFF v1.
		// Some files include padding or slightly inconsistent data sizes.
		if next > start && next <= len(data) {
			end = next
		}
		if end > len(data) {
			end = len(data)
		}

		var sp *Sprite
		if start < end {
			fallback := []byte(nil)
			if samePalette {
				fallback = sharedPal
			}
			dec, pal, err := decodePCXIndexed(data[start:end], fallback, !samePalette)
			if err != nil {
				s.recordDecodeError("SFF v1 sprite %d,%d index %d: %v", grp, num, i, err)
			} else if dec != nil && dec.indexed != nil {
				palIdx := sharedPalIdx
				if !samePalette || palIdx < 0 {
					sharedPal = pal
					sharedPalIdx = s.addPalette(1, uint16(len(s.Palettes)+1), pal)
					palIdx = sharedPalIdx
				}
				if palIdx < 0 {
					palIdx = s.addPalette(1, uint16(len(s.Palettes)+1), ensurePalette(pal))
				}
				sp = &Sprite{Group: grp, Number: num, W: dec.w, H: dec.h, XOff: xo, YOff: yo, Indexed: dec.indexed, IsIndexed: true, PalIndex: palIdx}
				s.addSprite(sp)
			} else {
				s.recordDecodeError("SFF v1 sprite %d,%d index %d: empty decode", grp, num, i)
			}
		}

		spritesByIndex = append(spritesByIndex, sp)
		if next == 0 {
			break
		}
		ofs = next
	}
	s.DecodedSpriteCount = len(s.Sprites)
	return s, nil
}

func loadSFFv2(data []byte, s *SFF) (*SFF, error) {
	spriteOfs, count := int(u32(data[36:40])), int(u32(data[40:44]))
	palOfs, palCount := int(u32(data[44:48])), int(u32(data[48:52]))
	lofs, tofs := int(u32(data[52:56])), int(u32(data[60:64]))
	s.Version = "v2"
	s.SpriteHeaderCount = count

	loadSFFv2Palettes(data, s, palOfs, palCount, lofs, data[13])
	spritesByIndex := make([]*Sprite, 0, count)

	for i := 0; i < count; i++ {
		h := spriteOfs + i*28
		if h+28 > len(data) {
			s.recordDecodeError("SFF v2 sprite header %d outside file", i)
			break
		}
		grp, num := u16(data[h:h+2]), u16(data[h+2:h+4])
		w, hh := int(u16(data[h+4:h+6])), int(u16(data[h+6:h+8]))
		xo, yo := int16(u16(data[h+8:h+10])), int16(u16(data[h+10:h+12]))
		link := int(u16(data[h+12 : h+14]))
		format, depth := int(data[h+14]), int(data[h+15])
		dofs, size := int(u32(data[h+16:h+20])), int(u32(data[h+20:h+24]))
		palHeaderIdx := int(u16(data[h+24 : h+26]))
		flags := u16(data[h+26 : h+28])
		palidx := s.paletteIndexFromHeader(palHeaderIdx)

		// SFF v2 linked sprites have size 0 and point at a previous sprite by
		// subfile index. They still need a separate map entry because AIR refers
		// to the linked sprite's own group/number.
		if size == 0 {
			if link >= 0 && link < len(spritesByIndex) && spritesByIndex[link] != nil {
				sp := cloneLinkedSprite(spritesByIndex[link], grp, num, xo, yo, palidx)
				s.addSprite(sp)
				s.LinkedSpriteCount++
				spritesByIndex = append(spritesByIndex, sp)
			} else {
				s.InvalidLinkedSprites++
				spritesByIndex = append(spritesByIndex, nil)
			}
			continue
		}

		if flags&1 == 0 {
			dofs += lofs
		} else {
			dofs += tofs
		}
		if dofs < 0 || dofs >= len(data) || size < 0 {
			s.recordDecodeError("SFF v2 sprite %d,%d index %d: invalid data offset/size offset=%d size=%d", grp, num, i, dofs, size)
			spritesByIndex = append(spritesByIndex, nil)
			continue
		}
		end := dofs + size
		if end > len(data) {
			s.recordDecodeError("SFF v2 sprite %d,%d index %d: data block truncated offset=%d size=%d", grp, num, i, dofs, size)
			end = len(data)
		}
		dec, err := decodeSFFv2Sprite(data[dofs:end], w, hh, format, depth)
		var sp *Sprite
		if err != nil {
			s.recordDecodeError("SFF v2 sprite %d,%d index %d format=%d depth=%d: %v", grp, num, i, format, depth, err)
		} else if dec != nil {
			sp = &Sprite{Group: grp, Number: num, W: dec.w, H: dec.h, XOff: xo, YOff: yo, Indexed: dec.indexed, RGBA: dec.rgba, IsIndexed: dec.isIndexed, PalIndex: palidx}
			s.addSprite(sp)
		} else {
			s.recordDecodeError("SFF v2 sprite %d,%d index %d format=%d depth=%d: empty decode", grp, num, i, format, depth)
		}
		spritesByIndex = append(spritesByIndex, sp)
	}
	s.DecodedSpriteCount = len(s.Sprites)
	return s, nil
}

func loadSFFv2Palettes(data []byte, s *SFF, palOfs, palCount, lofs int, alphaVersionByte byte) {
	// Palette headers must remain addressable by their original header index because
	// SFF v2 sprite headers store palette-header indices directly.
	s.PaletteIndexByHeader = make([]int, palCount)
	for i := range s.PaletteIndexByHeader {
		s.PaletteIndexByHeader[i] = -1
	}

	for i := 0; i < palCount; i++ {
		h := palOfs + i*16
		if h+16 > len(data) {
			s.recordDecodeError("SFF v2 palette header %d outside file", i)
			break
		}
		grp, item := u16(data[h:h+2]), u16(data[h+2:h+4])
		link := int(u16(data[h+6 : h+8]))
		ofs, size := int(u32(data[h+8:h+12]))+lofs, int(u32(data[h+12:h+16]))

		if size == 0 {
			if link >= 0 && link < len(s.PaletteIndexByHeader) {
				linkedIdx := s.PaletteIndexByHeader[link]
				if linkedIdx >= 0 && linkedIdx < len(s.Palettes) {
					s.PaletteIndexByHeader[i] = s.addPalette(grp, item, s.Palettes[linkedIdx].RGBA)
					continue
				}
			}
			s.InvalidLinkedPalettes++
			continue
		}

		if ofs < 0 || ofs+size > len(data) || size <= 0 {
			s.recordDecodeError("SFF v2 palette %d,%d header %d: invalid data offset/size offset=%d size=%d", grp, item, i, ofs, size)
			continue
		}
		s.PaletteIndexByHeader[i] = s.addPalette(grp, item, decodePaletteRGBA(data, ofs, size, alphaVersionByte))
	}
}

func decodePaletteRGBA(data []byte, ofs, size int, alphaVersionByte byte) []byte {
	pal := make([]byte, 256*4)

	n := size / 4
	if n > 256 {
		n = 256
	}

	for i := 0; i < n; i++ {
		r := data[ofs+i*4+0]
		g := data[ofs+i*4+1]
		b := data[ofs+i*4+2]
		a := data[ofs+i*4+3]

		// Match original Ikemen alpha handling for SFF v2.0.0.0:
		// index 0 forced transparent, all others forced opaque.
		if alphaVersionByte == 0 {
			if i == 0 {
				a = 0
			} else {
				a = 255
			}
		}

		pal[i*4+0] = r
		pal[i*4+1] = g
		pal[i*4+2] = b
		pal[i*4+3] = a
	}

	return pal
}

func decodeSFFv2Sprite(buf []byte, w, h, format, depth int) (*decodedSprite, error) {
	if len(buf) == 0 || w <= 0 || h <= 0 {
		return nil, nil
	}
	payload := buf
	if len(buf) > 4 && (format >= 2 && format <= 12) {
		payload = buf[4:]
	}
	switch format {
	case 0:
		if depth == 8 {
			return &decodedSprite{indexed: clampIndexed(payload, w*h), isIndexed: true, w: w, h: h}, nil
		}
		if depth == 24 || depth == 32 {
			return &decodedSprite{rgba: rawToRGBA(payload, w, h, depth), w: w, h: h}, nil
		}
	case 2:
		return &decodedSprite{indexed: rle8Decode(payload, w*h), isIndexed: true, w: w, h: h}, nil
	case 3:
		return &decodedSprite{indexed: rle5Decode(payload, w*h), isIndexed: true, w: w, h: h}, nil
	case 4:
		return &decodedSprite{indexed: lz5Decode(payload, w*h), isIndexed: true, w: w, h: h}, nil
	case 10, 11, 12:
		img, err := png.Decode(bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		if pi, ok := img.(*image.Paletted); ok && format == 10 {
			return &decodedSprite{indexed: palettedPixTight(pi), isIndexed: true, w: pi.Bounds().Dx(), h: pi.Bounds().Dy()}, nil
		}
		rgba, ww, hh := imageToRGBABytes(img)
		return &decodedSprite{rgba: rgba, w: ww, h: hh}, nil
	}
	return nil, fmt.Errorf("unsupported SFFv2 sprite format=%d depth=%d", format, depth)
}

func decodePCXIndexed(buf []byte, fallback []byte, readPalette bool) (*decodedSprite, []byte, error) {
	if len(buf) < 128 {
		return nil, nil, fmt.Errorf("short pcx")
	}

	xmin, ymin := int(u16(buf[4:6])), int(u16(buf[6:8]))
	xmax, ymax := int(u16(buf[8:10])), int(u16(buf[10:12]))
	w, h := xmax-xmin+1, ymax-ymin+1
	if w <= 0 || h <= 0 {
		return nil, nil, fmt.Errorf("invalid pcx size %dx%d", w, h)
	}

	bpl := int(u16(buf[66:68]))
	if bpl <= 0 {
		bpl = w
	}

	pal := fallback
	end := len(buf)

	// Match original Ikemen's SFF v1 rule: only look for an embedded PCX
	// palette when the SFF subheader says this sprite owns a new palette.
	// Same-palette subfiles often contain only the RLE stream. In those files,
	// bytes equal to 0x0c can appear inside normal RLE data; treating them as a
	// palette marker cuts the stream early and makes lower rows transparent.
	if readPalette {
		paletteOffset := -1
		if len(buf) >= 769 {
			for pos := len(buf) - 769; pos >= 128; pos-- {
				if buf[pos] == 0x0c {
					paletteOffset = pos
					break
				}
			}
			if paletteOffset < 0 {
				paletteOffset = len(buf) - 769
			}
		}

		if paletteOffset >= 0 && paletteOffset+769 <= len(buf) {
			pal = make([]byte, 256*4)
			p := buf[paletteOffset+1 : paletteOffset+769]
			for i := 0; i < 256; i++ {
				a := byte(255)
				if i == 0 {
					a = 0
				}
				pal[i*4+0] = p[i*3+0]
				pal[i*4+1] = p[i*3+1]
				pal[i*4+2] = p[i*3+2]
				pal[i*4+3] = a
			}
			end = paletteOffset
		}
	}

	pix := make([]byte, bpl*h)
	src := buf[128:end]
	j := 0
	for i := 0; i < len(src) && j < len(pix); i++ {
		c := src[i]
		n := 1
		if c >= 0xC0 && i+1 < len(src) {
			n = int(c & 0x3f)
			i++
			c = src[i]
		}
		for ; n > 0 && j < len(pix); n-- {
			pix[j] = c
			j++
		}
	}

	tight := make([]byte, w*h)
	for y := 0; y < h; y++ {
		copy(tight[y*w:(y+1)*w], pix[y*bpl:y*bpl+w])
	}
	return &decodedSprite{indexed: tight, isIndexed: true, w: w, h: h}, ensurePalette(pal), nil
}

func ensurePalette(pal []byte) []byte {
	if len(pal) >= 1024 {
		return pal[:1024]
	}
	out := make([]byte, 1024)
	copy(out, pal)
	if len(pal) == 0 {
		for i := 0; i < 256; i++ {
			out[i*4], out[i*4+1], out[i*4+2] = byte(i), byte(i), byte(i)
			if i == 0 {
				out[i*4+3] = 0
			} else {
				out[i*4+3] = 255
			}
		}
	}
	return out
}

func clampIndexed(src []byte, n int) []byte {
	out := make([]byte, n)
	copy(out, src)
	return out
}

func rawToRGBA(payload []byte, w, h, depth int) []byte {
	out := make([]byte, w*h*4)
	n := 0
	for i := 0; i < w*h; i++ {
		if n+2 >= len(payload) {
			break
		}
		b, g, r := payload[n], payload[n+1], payload[n+2]
		a := byte(255)
		n += 3
		if depth == 32 {
			if n >= len(payload) {
				break
			}
			a = payload[n]
			n++
		}
		out[i*4], out[i*4+1], out[i*4+2], out[i*4+3] = r, g, b, a
	}
	return out
}

func imageToRGBABytes(img image.Image) ([]byte, int, int) {
	b := img.Bounds()
	r := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(r, r.Bounds(), img, b.Min, draw.Src)
	return r.Pix, r.Bounds().Dx(), r.Bounds().Dy()
}

func palettedPixTight(pi *image.Paletted) []byte {
	b := pi.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([]byte, w*h)
	for y := 0; y < h; y++ {
		src := (y+b.Min.Y-pi.Rect.Min.Y)*pi.Stride + (b.Min.X - pi.Rect.Min.X)
		copy(out[y*w:(y+1)*w], pi.Pix[src:src+w])
	}
	return out
}

func paletteFromGoPalette(gp color.Palette) []byte {
	out := make([]byte, 1024)
	for i := 0; i < len(gp) && i < 256; i++ {
		r, g, b, a := gp[i].RGBA()
		out[i*4], out[i*4+1], out[i*4+2], out[i*4+3] = byte(r>>8), byte(g>>8), byte(b>>8), byte(a>>8)
	}
	return ensurePalette(out)
}

func rle8Decode(rle []byte, outSize int) []byte {
	if len(rle) == 0 {
		return rle
	}

	p := make([]byte, outSize)
	i, j := 0, 0

	for j < len(p) {
		n, d := 1, rle[i]
		if i < len(rle)-1 {
			i++
		}

		if d&0xc0 == 0x40 {
			n = int(d & 0x3f)
			d = rle[i]
			if i < len(rle)-1 {
				i++
			}
		}

		for ; n > 0; n-- {
			if j < len(p) {
				p[j] = d
				j++
			}
		}
	}

	return p
}

func rle5Decode(rle []byte, outSize int) []byte {
	if len(rle) == 0 {
		return rle
	}

	p := make([]byte, outSize)
	i, j := 0, 0

	for j < len(p) {
		rl := int(rle[i])
		if i < len(rle)-1 {
			i++
		}

		dl := int(rle[i] & 0x7f)
		c := byte(0)

		if rle[i]>>7 != 0 {
			if i < len(rle)-1 {
				i++
			}
			c = rle[i]
		}

		if i < len(rle)-1 {
			i++
		}

		for {
			if j < len(p) {
				p[j] = c
				j++
			}

			rl--
			if rl < 0 {
				dl--
				if dl < 0 {
					break
				}

				c = rle[i] & 0x1f
				rl = int(rle[i] >> 5)

				if i < len(rle)-1 {
					i++
				}
			}
		}
	}

	return p
}

func lz5Decode(rle []byte, outSize int) []byte {
	if len(rle) == 0 {
		return rle
	}

	p := make([]byte, outSize)
	i, j, n := 0, 0, 0

	ct, cts, rb, rbc := rle[i], uint(0), byte(0), uint(0)
	if i < len(rle)-1 {
		i++
	}

	for j < len(p) {
		d := int(rle[i])
		if i < len(rle)-1 {
			i++
		}

		if ct&byte(1<<cts) != 0 {
			if d&0x3f == 0 {
				d = (d<<2 | int(rle[i])) + 1
				if i < len(rle)-1 {
					i++
				}

				n = int(rle[i]) + 2
				if i < len(rle)-1 {
					i++
				}
			} else {
				rb |= byte(d&0xc0) >> rbc
				rbc += 2

				n = int(d & 0x3f)
				if rbc < 8 {
					d = int(rle[i]) + 1
					if i < len(rle)-1 {
						i++
					}
				} else {
					d = int(rb) + 1
					rb, rbc = 0, 0
				}
			}

			for {
				if j < len(p) {
					p[j] = p[j-d]
					j++
				}

				n--
				if n < 0 {
					break
				}
			}
		} else {
			if d&0xe0 == 0 {
				n = int(rle[i]) + 8
				if i < len(rle)-1 {
					i++
				}
			} else {
				n = d >> 5
				d &= 0x1f
			}

			for ; n > 0; n-- {
				if j < len(p) {
					p[j] = byte(d)
					j++
				}
			}
		}

		cts++
		if cts >= 8 {
			ct, cts = rle[i], 0
			if i < len(rle)-1 {
				i++
			}
		}
	}

	return p
}
