package main

import "sort"

type DrawList struct {
	items []RenderParams
	seq   int
}

func (dl *DrawList) Clear() {
	dl.items = dl.items[:0]
	dl.seq = 0
}

func (dl *DrawList) Add(rp RenderParams) {
	rp.SortIndex = dl.seq
	dl.seq++
	dl.items = append(dl.items, rp)
}

func (dl *DrawList) DrawLayer(layer int, r *GLRenderer) error {
	var list []RenderParams
	for _, rp := range dl.items {
		if rp.Layer == layer {
			list = append(list, rp)
		}
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Priority != list[j].Priority {
			return list[i].Priority < list[j].Priority
		}
		return list[i].SortIndex < list[j].SortIndex
	})
	for _, rp := range list {
		if err := r.RenderSprite(rp); err != nil {
			return err
		}
	}
	return nil
}
