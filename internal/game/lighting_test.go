package game

import (
	"testing"
	"time"

	"gohack/internal/entity"
	"gohack/internal/geom"
	"gohack/internal/world"
)

// cellAt returns the view cell showing map position mp, for a view built around p.
func cellAt(v *View, p *entity.Entity, mp geom.Point) Cell {
	return v.At(mp.X-p.Pos.X+v.W/2, mp.Y-p.Pos.Y+v.H/2)
}

func lightGame(t *testing.T, src string) (*Game, *entity.Entity) {
	t.Helper()
	return newGame(t, &seq{}, map[string]string{"a": "id: a\n---\n" + src})
}

func TestLightFalloff(t *testing.T) {
	g, p := lightGame(t, `#############
#@..*.......#
#############
`)
	v := g.BuildView(p.ID, 21, 9, 0)
	near := cellAt(v, p, geom.Point{X: 3, Y: 1}).Light.R
	mid := cellAt(v, p, geom.Point{X: 6, Y: 1}).Light.R
	far := cellAt(v, p, geom.Point{X: 10, Y: 1}).Light.R
	if near <= mid || mid <= far || far == 0 {
		t.Fatalf("light should fade with distance: near=%d mid=%d far=%d", near, mid, far)
	}
	if out := cellAt(v, p, geom.Point{X: 11, Y: 1}); out.Vis != Visible || out.Light != (world.RGB{}) {
		t.Fatalf("tile beyond the radius should be visible but unlit: %+v", out)
	}
	if self := cellAt(v, p, geom.Point{X: 4, Y: 1}).Light; self.R <= near {
		t.Fatalf("the light's own tile should be the brightest: %v", self)
	}
}

func TestLightOnlyReachesWhatItSees(t *testing.T) {
	g, p := lightGame(t, `#######
#....@#
#.....#
#*.#..#
#######
`)
	v := g.BuildView(p.ID, 21, 9, 0)
	shadowed := cellAt(v, p, geom.Point{X: 4, Y: 3})
	if shadowed.Vis != Visible {
		t.Fatal("test setup: the player should see the tile behind the pillar")
	}
	if shadowed.Light != (world.RGB{}) {
		t.Fatalf("tile hidden from the light must stay unlit, got %v", shadowed.Light)
	}
	if cellAt(v, p, geom.Point{X: 2, Y: 3}).Light == (world.RGB{}) {
		t.Fatal("tile next to the light should be lit")
	}
}

func TestWallLitOnlyOnSideFacingPlayerAndLight(t *testing.T) {
	north, pn := lightGame(t, "#######\n#@.*..#\n#..#..#\n#.....#\n#.....#\n#######\n")
	wall := geom.Point{X: 3, Y: 2}
	c := cellAt(north.BuildView(pn.ID, 21, 9, 0), pn, wall)
	if c.Vis != Visible || c.Light == (world.RGB{}) {
		t.Fatalf("wall seen from the light's side should be lit: %+v", c)
	}

	south, ps := lightGame(t, "#######\n#..*..#\n#..#..#\n#.....#\n#..@..#\n#######\n")
	c = cellAt(south.BuildView(ps.ID, 21, 9, 0), ps, wall)
	if c.Vis != Visible || c.Light != (world.RGB{}) {
		t.Fatalf("wall seen from behind must not look lit: %+v", c)
	}
}

func TestDoorOpeningInvalidatesLightCache(t *testing.T) {
	g, p := lightGame(t, `##########
#@.+.*...#
##########
`)
	behind := geom.Point{X: 1, Y: 1}
	g.Submit(p.ID, MoveCmd{geom.E})
	if c := cellAt(g.BuildView(p.ID, 21, 9, 0), p, behind); c.Vis != Visible || c.Light != (world.RGB{}) {
		t.Fatalf("closed door should keep the light out: %+v", c)
	}
	g.Submit(p.ID, MoveCmd{geom.E}) // opens the door
	if c := cellAt(g.BuildView(p.ID, 21, 9, 0), p, behind); c.Light == (world.RGB{}) {
		t.Fatal("light should reach through the opened door")
	}
}

func TestToggleLight(t *testing.T) {
	g, p := lightGame(t, `#####
#@*.#
#####
`)
	last := func() string {
		m := g.BuildView(p.ID, 9, 5, 0).Messages
		return m[len(m)-1]
	}
	lit := func() bool {
		return cellAt(g.BuildView(p.ID, 9, 5, 0), p, geom.Point{X: 3, Y: 1}).Light != world.RGB{}
	}
	if !lit() {
		t.Fatal("torch should start lit")
	}
	g.Submit(p.ID, ToggleLightCmd{geom.E})
	if g.Turn != 1 || last() != "You put out the torch." || lit() {
		t.Fatalf("turn=%d msg=%q lit=%v", g.Turn, last(), lit())
	}
	g.Submit(p.ID, ToggleLightCmd{geom.E})
	if g.Turn != 2 || last() != "You light the torch." || !lit() {
		t.Fatalf("turn=%d msg=%q lit=%v", g.Turn, last(), lit())
	}
	g.Submit(p.ID, ToggleLightCmd{geom.W}) // wall
	if g.Turn != 2 || last() != "There is no light there." {
		t.Fatalf("turn=%d msg=%q", g.Turn, last())
	}
}

func TestFlicker(t *testing.T) {
	def := world.LightTemplates["torch"]
	li := &world.Light{Def: &def, Pos: geom.Point{X: 3, Y: 4}, On: true}
	base := def.Color
	varied := false
	first := lightColor("a", li, 0)
	for at := time.Duration(0); at < 3*time.Second; at += 13 * time.Millisecond {
		c := lightColor("a", li, at)
		if c != lightColor("a", li, at) {
			t.Fatalf("flicker must be deterministic at %v", at)
		}
		if c.R > base.R || float64(c.R) < 0.74*float64(base.R) {
			t.Fatalf("flicker out of range at %v: %v", at, c)
		}
		varied = varied || c != first
	}
	if !varied {
		t.Fatal("flicker never changed")
	}
	steady := &world.Light{Def: &world.LightDef{Color: base}, On: true}
	if lightColor("a", steady, 777*time.Millisecond) != base {
		t.Fatal("light without animation must keep its color")
	}
}

func TestCrystalIsSteadyAndBlue(t *testing.T) {
	def := world.LightTemplates["crystal"]
	li := &world.Light{Def: &def, Pos: geom.Point{X: 2, Y: 2}, On: true}
	for _, at := range []time.Duration{0, 37 * time.Millisecond, time.Second, 90 * time.Second} {
		if c := lightColor("a", li, at); c != def.Color {
			t.Fatalf("crystal color changed at %v: %v", at, c)
		}
	}

	g, p := lightGame(t, "#####\n#@.&#\n#####\n")
	v := g.BuildView(p.ID, 9, 5, 0)
	if c := cellAt(v, p, geom.Point{X: 2, Y: 1}).Light; c.B <= c.R || c.B <= c.G {
		t.Fatalf("floor next to a crystal should be tinted blue, got %v", c)
	}
	if v2 := g.BuildView(p.ID, 9, 5, 5*time.Second); cellAt(v2, p, geom.Point{X: 2, Y: 1}).Light != cellAt(v, p, geom.Point{X: 2, Y: 1}).Light {
		t.Fatal("crystal light must not animate")
	}
}

func TestCellSourceStaysWhenRemembered(t *testing.T) {
	g, p := lightGame(t, "#######\n#&.@..#\n###+###\n  #.#\n  ###\n")
	src := geom.Point{X: 1, Y: 1}
	if s := cellAt(g.BuildView(p.ID, 9, 5, 0), p, src).Source; s == nil || s.Glyph != '&' {
		t.Fatalf("visible light source should carry its definition, got %+v", s)
	}
	g.Submit(p.ID, MoveCmd{geom.S}) // opens the door
	g.Submit(p.ID, MoveCmd{geom.S})
	g.Submit(p.ID, MoveCmd{geom.S}) // out of sight of the crystal
	c := cellAt(g.BuildView(p.ID, 9, 5, 0), p, src)
	if c.Vis != Remembered || c.Source == nil || c.Source.Glyph != '&' {
		t.Fatalf("remembered light source should keep its definition: %+v", c)
	}
}
