package fov

import (
	"math/rand/v2"
	"strings"
	"testing"
)

// grid is a test map: '#' is opaque, anything else transparent; outside is opaque.
type grid struct {
	w, h int
	op   []bool
}

func parse(s string) (*grid, int, int) {
	rows := strings.Split(strings.Trim(s, "\n"), "\n")
	g := &grid{w: len(rows[0]), h: len(rows), op: make([]bool, len(rows[0])*len(rows))}
	ox, oy := -1, -1
	for y, r := range rows {
		for x, c := range r {
			g.op[y*g.w+x] = c == '#'
			if c == '@' {
				ox, oy = x, y
			}
		}
	}
	return g, ox, oy
}

func (g *grid) opaque(x, y int) bool {
	return x < 0 || y < 0 || x >= g.w || y >= g.h || g.op[y*g.w+x]
}

func (g *grid) visible(ox, oy, radius int) map[[2]int]bool {
	vis := map[[2]int]bool{}
	Compute(ox, oy, radius, g.opaque, func(x, y int) { vis[[2]int{x, y}] = true })
	return vis
}

func TestPillarCastsShadow(t *testing.T) {
	g, ox, oy := parse(`
###########
#.........#
#.@.#.....#
#.........#
###########
`)
	vis := g.visible(ox, oy, 20)
	if !vis[[2]int{4, 2}] {
		t.Error("pillar itself must be visible")
	}
	if vis[[2]int{6, 2}] || vis[[2]int{7, 2}] {
		t.Error("cells right behind the pillar must be hidden")
	}
	if !vis[[2]int{6, 1}] || !vis[[2]int{6, 3}] {
		t.Error("cells beside the shadow must be visible")
	}
	if !vis[[2]int{0, 0}] && !vis[[2]int{1, 0}] {
		t.Error("walls bordering the room must be visible")
	}
}

func TestClosedDoorBlocks(t *testing.T) {
	closed, ox, oy := parse(`
#######
#@.#..#
#######
`)
	if v := closed.visible(ox, oy, 20); !v[[2]int{3, 1}] || v[[2]int{4, 1}] || v[[2]int{5, 1}] {
		t.Error("opaque cell visible, cells behind it hidden")
	}
	open, ox, oy := parse(`
#######
#@.'..#
#######
`)
	if v := open.visible(ox, oy, 20); !v[[2]int{5, 1}] {
		t.Error("open door must let sight through")
	}
}

func TestRadius(t *testing.T) {
	g, ox, oy := parse(`
...........
.....@.....
...........
`)
	// everything outside the grid is opaque, so only the interior is lit
	v := g.visible(ox, oy, 3)
	if !v[[2]int{8, 1}] || v[[2]int{9, 1}] {
		t.Error("radius 3 should reach x=8 but not x=9")
	}
}

func TestSymmetry(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	const n = 40
	for iter := 0; iter < 30; iter++ {
		g := &grid{w: n, h: n, op: make([]bool, n*n)}
		for i := range g.op {
			g.op[i] = rng.IntN(100) < 20
		}
		var floors [][2]int
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				if !g.opaque(x, y) {
					floors = append(floors, [2]int{x, y})
				}
			}
		}
		a := floors[rng.IntN(len(floors))]
		va := g.visible(a[0], a[1], 100)
		for _, b := range floors {
			if va[b] != g.visible(b[0], b[1], 100)[a] {
				t.Fatalf("asymmetric visibility between %v and %v", a, b)
			}
		}
	}
}

func BenchmarkOpenField(b *testing.B) {
	const n = 1024
	op := make([]bool, n*n)
	rng := rand.New(rand.NewPCG(3, 4))
	for i := range op {
		op[i] = rng.IntN(100) < 10
	}
	isOpaque := func(x, y int) bool { return x < 0 || y < 0 || x >= n || y >= n || op[y*n+x] }
	count := 0
	mark := func(x, y int) { count++ }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compute(512, 512, 40, isOpaque, mark)
	}
}
