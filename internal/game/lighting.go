package game

import (
	"hash/fnv"
	"math"
	"time"

	"gohack/internal/fov"
	"gohack/internal/geom"
	"gohack/internal/world"
)

// Face bits. A transparent tile has only faceAll; an opaque tile (wall, closed
// door) has one bit per side that borders a transparent tile.
const (
	faceN uint8 = 1 << iota
	faceE
	faceS
	faceW
	faceAll
)

var faceDirs = [...]struct {
	bit    uint8
	dx, dy int
}{
	{faceN, 0, -1}, {faceE, 1, 0}, {faceS, 0, 1}, {faceW, -1, 0},
}

// lightCell is how one tile is lit by one light: how strongly, and on which faces.
type lightCell struct {
	intensity float32 // 0 (unlit) to 1
	faces     uint8
}

// lightField is the area one light reaches, computed from line of sight from
// the light. It does not depend on whether the light is on, so toggling a
// light never invalidates it; only tile changes (a door opening) do.
type lightField struct {
	rev    uint64 // level revision it was built at
	origin geom.Point
	size   int // side length: 2*radius+1
	cells  []lightCell
}

func (f *lightField) at(p geom.Point) lightCell {
	x, y := p.X-f.origin.X, p.Y-f.origin.Y
	if x < 0 || y < 0 || x >= f.size || y >= f.size {
		return lightCell{}
	}
	return f.cells[y*f.size+x]
}

// falloff is the brightness at distance d from a light of the given radius.
func falloff(d float64, radius int) float32 {
	return float32(max(0, 1-d/float64(radius+1)))
}

// exposed reports whether the face of opaque tile p towards (dx, dy) borders
// a transparent tile.
func exposed(l *world.Level, p geom.Point, dx, dy int) bool {
	return l.Transparent(geom.Point{X: p.X + dx, Y: p.Y + dy})
}

// buildField computes the area lit by li on level l.
func buildField(l *world.Level, li *world.Light) *lightField {
	r := li.Def.Radius
	f := &lightField{
		rev:    l.Rev(),
		origin: geom.Point{X: li.Pos.X - r, Y: li.Pos.Y - r},
		size:   2*r + 1,
	}
	f.cells = make([]lightCell, f.size*f.size)
	reached := newBitgrid(f.size, f.size)
	fov.Compute(li.Pos.X, li.Pos.Y, r,
		func(x, y int) bool { return !l.Transparent(geom.Point{X: x, Y: y}) },
		func(x, y int) { reached.Set(x-f.origin.X, y-f.origin.Y) })

	for ly := 0; ly < f.size; ly++ {
		for lx := 0; lx < f.size; lx++ {
			if !reached.Has(lx, ly) {
				continue
			}
			p := geom.Point{X: f.origin.X + lx, Y: f.origin.Y + ly}
			d := math.Hypot(float64(p.X-li.Pos.X), float64(p.Y-li.Pos.Y))
			c := lightCell{intensity: falloff(d, r)}
			if l.Transparent(p) {
				c.faces = faceAll
			} else {
				// A side is lit if the light is on that side and sees the
				// tile next to it.
				for _, fd := range faceDirs {
					if (li.Pos.X-p.X)*fd.dx+(li.Pos.Y-p.Y)*fd.dy > 0 &&
						exposed(l, p, fd.dx, fd.dy) &&
						reached.Has(lx+fd.dx, ly+fd.dy) {
						c.faces |= fd.bit
					}
				}
			}
			f.cells[ly*f.size+lx] = c
		}
	}
	return f
}

// field returns the cached light field of li, rebuilding it if the level's
// tiles changed since.
func (g *Game) field(l *world.Level, li *world.Light) *lightField {
	if f := g.fields[li]; f != nil && f.rev == l.Rev() {
		return f
	}
	f := buildField(l, li)
	g.fields[li] = f
	return f
}

// viewerFaces returns the faces of tile p that the viewer at pos can see.
func viewerFaces(l *world.Level, visible *bitgrid, pos, p geom.Point) uint8 {
	if l.Transparent(p) {
		return faceAll
	}
	var faces uint8
	for _, fd := range faceDirs {
		if (pos.X-p.X)*fd.dx+(pos.Y-p.Y)*fd.dy > 0 &&
			exposed(l, p, fd.dx, fd.dy) &&
			visible.Has(p.X+fd.dx, p.Y+fd.dy) {
			faces |= fd.bit
		}
	}
	return faces
}

// addLight adds the light of every lit light source on level l to the cells of
// v. A tile is lit by a light only if the light reaches a face of it that the
// viewer at pos can see.
func (g *Game) addLight(v *View, l *world.Level, p *perception, pos, origin geom.Point, at time.Duration) {
	for _, li := range l.Lights {
		if !li.On {
			continue
		}
		f := g.field(l, li)
		col := lightColor(l.ID, li, at)
		x0, y0 := max(f.origin.X, origin.X), max(f.origin.Y, origin.Y)
		x1 := min(f.origin.X+f.size, origin.X+v.W)
		y1 := min(f.origin.Y+f.size, origin.Y+v.H)
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				c := &v.Cells[(y-origin.Y)*v.W+(x-origin.X)]
				if c.Vis != Visible {
					continue
				}
				mp := geom.Point{X: x, Y: y}
				lc := f.at(mp)
				if lc.intensity == 0 || lc.faces&viewerFaces(l, p.visible, pos, mp) == 0 {
					continue
				}
				c.Light = world.RGB{
					R: satAdd(c.Light.R, col.R, lc.intensity),
					G: satAdd(c.Light.G, col.G, lc.intensity),
					B: satAdd(c.Light.B, col.B, lc.intensity),
				}
			}
		}
	}
}

// satAdd returns a + b*k, capped at 255.
func satAdd(a, b uint8, k float32) uint8 {
	return uint8(min(255, int(a)+int(float32(b)*k)))
}

const (
	flickerStep = 100 * time.Millisecond
	flickerMin  = 0.75 // dimmest flicker, as a fraction of the light's color
)

// lightColor is the color of li at animation time at.
func lightColor(level world.LevelID, li *world.Light, at time.Duration) world.RGB {
	c := li.Def.Color
	if li.Def.Anim != world.AnimFlicker {
		return c
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(level))
	seed := h.Sum64() ^ uint64(li.Pos.X)*0x9e3779b97f4a7c15 ^ uint64(li.Pos.Y)*0xc2b2ae3d27d4eb4f
	step := int64(at / flickerStep)
	frac := float64(at%flickerStep) / float64(flickerStep)
	frac = frac * frac * (3 - 2*frac) // smoothstep
	n := noise(seed, step)*(1-frac) + noise(seed, step+1)*frac
	k := flickerMin + (1-flickerMin)*n
	return world.RGB{R: uint8(float64(c.R) * k), G: uint8(float64(c.G) * k), B: uint8(float64(c.B) * k)}
}

// noise is a deterministic hash of (seed, step) in [0, 1). The animation
// deliberately avoids rules.RNG so it can never disturb game rules.
func noise(seed uint64, step int64) float64 {
	x := seed + uint64(step)*0x9e3779b97f4a7c15
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return float64(x>>11) / (1 << 53)
}
