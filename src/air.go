package main

import (
	"bufio"
	"os"
	"sort"
	"strconv"
	"strings"
)

type AnimFrame struct {
	Group, Number  int
	Xoff, Yoff     int
	Time           int
	HFlip, VFlip   bool
	XScale, YScale float64
	Angle          float64
}

type Animation struct {
	No     int
	Frames []AnimFrame
	idx    int
	tick   int
}

func parseInt(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
func parseFloat(s string, def float64) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}

func LoadAIR(path string) ([]*Animation, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []*Animation
	var cur *Animation
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.Index(line, ";"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		low := strings.ToLower(line)
		if strings.HasPrefix(low, "[begin action") {
			n := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(low, "[begin action"), "]"))
			cur = &Animation{No: parseInt(n, 0)}
			out = append(out, cur)
			continue
		}
		if strings.HasPrefix(line, "[") {
			cur = nil
			continue
		}
		if cur == nil {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 5 {
			continue
		}
		fr := AnimFrame{Group: parseInt(parts[0], -1), Number: parseInt(parts[1], -1), Xoff: parseInt(parts[2], 0), Yoff: parseInt(parts[3], 0), Time: parseInt(parts[4], 1), XScale: 1, YScale: 1}
		if fr.Time <= 0 {
			fr.Time = 1
		}
		if len(parts) >= 6 {
			flags := strings.ToLower(parts[5])
			fr.HFlip = strings.Contains(flags, "h")
			fr.VFlip = strings.Contains(flags, "v")
		}
		if len(parts) >= 8 {
			fr.XScale = parseFloat(parts[7], 1)
		}
		if len(parts) >= 9 {
			fr.YScale = parseFloat(parts[8], 1)
		}
		if len(parts) >= 10 {
			fr.Angle = parseFloat(parts[9], 0)
		}
		if fr.Group >= 0 && fr.Number >= 0 {
			cur.Frames = append(cur.Frames, fr)
		}
	}
	return out, sc.Err()
}

func (a *Animation) Reset() { a.idx, a.tick = 0, 0 }
func (a *Animation) CurrentFrame() AnimFrame {
	if len(a.Frames) == 0 {
		return AnimFrame{Time: 1, XScale: 1, YScale: 1}
	}
	return a.Frames[a.idx]
}
func (a *Animation) Step() {
	if len(a.Frames) == 0 {
		return
	}
	fr := a.Frames[a.idx]
	a.tick++
	if a.tick >= fr.Time {
		a.tick = 0
		a.idx = (a.idx + 1) % len(a.Frames)
	}
}
func DrawableAnimations(anims []*Animation, sff *SFF) []*Animation {
	var drawable []*Animation
	for _, a := range anims {
		for _, fr := range a.Frames {
			if sff.Sprites[[2]uint16{uint16(fr.Group), uint16(fr.Number)}] != nil {
				drawable = append(drawable, a)
				break
			}
		}
	}
	sort.Slice(drawable, func(i, j int) bool { return drawable[i].No < drawable[j].No })
	return drawable
}
