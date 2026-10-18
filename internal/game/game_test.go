package game

import (
	"strings"
	"testing"
	"testing/fstest"

	"gohack/internal/entity"
	"gohack/internal/geom"
	"gohack/internal/world"
)

// seq is an RNG returning fixed values, then zeros.
type seq []int

func (s *seq) IntN(int) int {
	if len(*s) == 0 {
		return 0
	}
	v := (*s)[0]
	*s = (*s)[1:]
	return v
}

func newGame(t *testing.T, rng *seq, levels map[string]string) (*Game, *entity.Entity) {
	t.Helper()
	fsys := fstest.MapFS{}
	for name, src := range levels {
		fsys["l/"+name+".txt"] = &fstest.MapFile{Data: []byte(src)}
	}
	w, err := world.Load(fsys, "l")
	if err != nil {
		t.Fatal(err)
	}
	g, err := New(w, rng)
	if err != nil {
		t.Fatal(err)
	}
	return g, g.AddPlayer("Hero", entity.Human, entity.Knight)
}

const arena = `id: a
name: Arena
portal: 6,1 -> b:1,1
---
#######
#@..+.>
#######
`

const second = `id: b
name: Cellar
---
####
#>.#
####
`

func TestMoveBlockedAndDoor(t *testing.T) {
	g, p := newGame(t, &seq{}, map[string]string{"a": arena, "b": second})
	g.Submit(p.ID, MoveCmd{geom.N}) // wall: no turn
	if p.Pos != (geom.Point{X: 1, Y: 1}) || g.Turn != 0 {
		t.Fatalf("wall must block without spending a turn: pos=%v turn=%d", p.Pos, g.Turn)
	}
	g.Submit(p.ID, MoveCmd{geom.E})
	g.Submit(p.ID, MoveCmd{geom.E})
	if p.Pos != (geom.Point{X: 3, Y: 1}) || g.Turn != 2 {
		t.Fatalf("pos=%v turn=%d", p.Pos, g.Turn)
	}
	l := g.World.Level("a")
	g.Submit(p.ID, MoveCmd{geom.E}) // opens the door, stays put
	if l.At(geom.Point{X: 4, Y: 1}) != world.DoorOpen || p.Pos.X != 3 || g.Turn != 3 {
		t.Fatalf("door should open and player stay: pos=%v turn=%d", p.Pos, g.Turn)
	}
	g.Submit(p.ID, MoveCmd{geom.E})
	if p.Pos.X != 4 {
		t.Fatalf("player should walk through open door, pos=%v", p.Pos)
	}
}

func TestDoorBlocksSight(t *testing.T) {
	g, p := newGame(t, &seq{}, map[string]string{"a": arena, "b": second})
	g.Submit(p.ID, MoveCmd{geom.E})
	g.Submit(p.ID, MoveCmd{geom.E})
	v := g.BuildView(p.ID, 21, 9)
	// player at view center (10,4); door at +1; stairs behind the door at +3
	if v.At(11, 4).Vis != Visible || v.At(13, 4).Vis != Unseen {
		t.Fatalf("door visible=%v stairs=%v", v.At(11, 4).Vis, v.At(13, 4).Vis)
	}
	g.Submit(p.ID, MoveCmd{geom.E}) // open
	v = g.BuildView(p.ID, 21, 9)
	if v.At(13, 4).Vis != Visible {
		t.Fatal("stairs should be visible through open door")
	}
}

func TestStairsBothWays(t *testing.T) {
	g, p := newGame(t, &seq{}, map[string]string{"a": arena, "b": second})
	g.Submit(p.ID, UseStairsCmd{})
	if p.Level != "a" || g.Turn != 0 {
		t.Fatal("no stairs here: nothing should happen")
	}
	p.Pos = geom.Point{X: 6, Y: 1}
	g.moveTo(p, "a", p.Pos)
	g.Submit(p.ID, UseStairsCmd{})
	if p.Level != "b" || p.Pos != (geom.Point{X: 1, Y: 1}) {
		t.Fatalf("level=%s pos=%v", p.Level, p.Pos)
	}
	g.Submit(p.ID, UseStairsCmd{})
	if p.Level != "a" || p.Pos != (geom.Point{X: 6, Y: 1}) {
		t.Fatalf("back: level=%s pos=%v", p.Level, p.Pos)
	}
}

const duel = `id: a
---
#####
#@g.#
#####
`

func TestAttackAndHostileGoblin(t *testing.T) {
	// Knight: roll 20 (IntN 19) hits AC 12; long sword 1d8 rolls IntN 1 -> 2, +2 STR = 4. Goblin 18 HP -> 14.
	// Goblin turn: roll 1 (IntN 0) misses.
	g, p := newGame(t, &seq{19, 1, 0}, map[string]string{"a": duel})
	gob := g.EntityAt("a", geom.Point{X: 2, Y: 1})
	if gob == nil || gob.Disposition != entity.Hostile {
		t.Fatal("goblin should be spawned and hostile")
	}
	g.Submit(p.ID, MoveCmd{geom.E}) // bump: no attack, no turn
	if gob.HP != gob.MaxHP || g.Turn != 0 {
		t.Fatal("bumping must not attack or spend a turn")
	}
	evs := g.Submit(p.ID, AttackCmd{geom.E})
	if gob.HP != 14 {
		t.Fatalf("goblin hp = %d, want 14", gob.HP)
	}
	if g.Turn != 1 || p.HP != p.MaxHP {
		t.Fatalf("turn=%d player hp=%d", g.Turn, p.HP)
	}
	var goblinActed bool
	for _, ev := range evs {
		goblinActed = goblinActed || (ev.Kind == EvAttack && ev.Actor == gob.ID)
	}
	if !goblinActed {
		t.Fatal("adjacent hostile goblin should attack after the player's turn")
	}
	if g.BuildView(p.ID, 9, 5).Messages[1] != "You hit the goblin for 4." {
		t.Fatalf("messages = %v", g.BuildView(p.ID, 9, 5).Messages)
	}
}

func TestNothingToAttackIsFree(t *testing.T) {
	g, p := newGame(t, &seq{}, map[string]string{"a": duel})
	g.Submit(p.ID, AttackCmd{geom.W})
	if g.Turn != 0 {
		t.Fatal("attacking empty air must not spend a turn")
	}
}

func TestDeath(t *testing.T) {
	g, p := newGame(t, &seq{19, 3, 19, 0}, map[string]string{"a": duel})
	gob := g.EntityAt("a", geom.Point{X: 2, Y: 1})
	gob.HP = 1
	g.Submit(p.ID, AttackCmd{geom.E})
	if g.Entity(gob.ID) != nil || g.EntityAt("a", geom.Point{X: 2, Y: 1}) != nil {
		t.Fatal("dead goblin must be removed")
	}

	g, p = newGame(t, &seq{0, 19, 1}, map[string]string{"a": duel}) // knight misses, goblin hits
	p.HP = 1
	g.Submit(p.ID, AttackCmd{geom.E})
	if !p.IsDead() {
		t.Fatalf("player should be dead, hp=%d", p.HP)
	}
	v := g.BuildView(p.ID, 9, 5)
	if !v.Dead || !strings.Contains(strings.Join(v.Messages, "|"), "You die.") {
		t.Fatalf("dead=%v messages=%v", v.Dead, v.Messages)
	}
	if evs := g.Submit(p.ID, MoveCmd{geom.E}); evs != nil {
		t.Fatal("dead players cannot act")
	}
}

func TestViewCenteredAndRemembered(t *testing.T) {
	g, p := newGame(t, &seq{}, map[string]string{"a": arena, "b": second})
	v := g.BuildView(p.ID, 11, 7)
	if a := v.At(5, 3).Actor; a == nil || !a.IsPlayer {
		t.Fatal("player must be at the view center")
	}
	g.Submit(p.ID, MoveCmd{geom.E})
	g.Submit(p.ID, MoveCmd{geom.E})
	v = g.BuildView(p.ID, 11, 7)
	if a := v.At(5, 3).Actor; a == nil || !a.IsPlayer {
		t.Fatal("player must stay centered after moving")
	}
	if c := v.At(5-2, 3); c.Tile != world.Floor || c.Vis != Visible {
		t.Fatalf("cell behind player: %+v", c)
	}
}

const yard = `id: a
---
######
#@g..#
#....#
######
`

func TestPassiveNPCsTurnHostileWhenAttacked(t *testing.T) {
	for _, d := range []entity.Disposition{entity.Neutral, entity.Friendly} {
		// knight misses (roll 1); goblin then hits with roll 20 for min damage 1
		g, p := newGame(t, &seq{0, 19, 0}, map[string]string{"a": yard})
		gob := g.EntityAt("a", geom.Point{X: 2, Y: 1})
		gob.Disposition = d

		g.Submit(p.ID, MoveCmd{geom.S}) // still diagonal-adjacent to the goblin
		g.Submit(p.ID, MoveCmd{geom.N})
		if p.HP != p.MaxHP || gob.Disposition != d {
			t.Fatalf("disposition %v: passive NPC must not attack (hp=%d)", d, p.HP)
		}
		g.Submit(p.ID, AttackCmd{geom.E})
		if gob.Disposition != entity.Hostile {
			t.Fatalf("disposition %v: attacked NPC must turn hostile", d)
		}
		if p.HP != p.MaxHP-1 {
			t.Fatalf("disposition %v: the now-hostile NPC should retaliate, hp=%d", d, p.HP)
		}
	}
}
