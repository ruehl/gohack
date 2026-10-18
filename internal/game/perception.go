package game

import (
	"gohack/internal/fov"
	"gohack/internal/geom"
	"gohack/internal/world"
)

// FOVRadius is how far an entity can see, in tiles.
const FOVRadius = 32

// bitgrid is a fixed-size bit set over level coordinates.
type bitgrid struct {
	w, h int
	bits []uint64
}

func newBitgrid(w, h int) *bitgrid {
	return &bitgrid{w: w, h: h, bits: make([]uint64, (w*h+63)/64)}
}

func (b *bitgrid) Set(x, y int) {
	if x < 0 || y < 0 || x >= b.w || y >= b.h {
		return
	}
	i := y*b.w + x
	b.bits[i>>6] |= 1 << (i & 63)
}

func (b *bitgrid) Has(x, y int) bool {
	if x < 0 || y < 0 || x >= b.w || y >= b.h {
		return false
	}
	i := y*b.w + x
	return b.bits[i>>6]&(1<<(i&63)) != 0
}

func (b *bitgrid) Clear() { clear(b.bits) }

// perception is what one player currently sees and remembers.
type perception struct {
	level   world.LevelID
	visible *bitgrid
	seen    map[world.LevelID]*bitgrid
}

// refresh recomputes line of sight from pos on level l.
func (p *perception) refresh(l *world.Level, pos geom.Point) {
	if p.visible == nil || p.level != l.ID {
		p.visible = newBitgrid(l.W, l.H)
		p.level = l.ID
	} else {
		p.visible.Clear()
	}
	seen := p.seen[l.ID]
	if seen == nil {
		seen = newBitgrid(l.W, l.H)
		p.seen[l.ID] = seen
	}
	fov.Compute(pos.X, pos.Y, FOVRadius,
		func(x, y int) bool { return !l.Transparent(geom.Point{X: x, Y: y}) },
		func(x, y int) {
			p.visible.Set(x, y)
			seen.Set(x, y)
		})
}
