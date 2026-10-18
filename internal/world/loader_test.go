package world

import (
	"strings"
	"testing"

	"gohack/internal/geom"
)

func TestLoadEmbedded(t *testing.T) {
	w, err := LoadEmbedded()
	if err != nil {
		t.Fatal(err)
	}
	if w.Start != "level1" {
		t.Fatalf("start level = %q", w.Start)
	}
	l1, l2 := w.Level("level1"), w.Level("level2")
	if len(l1.Spawns) != 1 || l1.Spawns[0].Template != "goblin" {
		t.Fatalf("spawns = %+v", l1.Spawns)
	}
	// Both ends of the stair exist and point at each other.
	p1 := l1.Portals[geom.Point{X: 33, Y: 3}]
	if p1.Target != "level2" || p1.TargetPos != (geom.Point{X: 3, Y: 1}) {
		t.Fatalf("portal 1 = %+v", p1)
	}
	p2 := l2.Portals[geom.Point{X: 3, Y: 1}]
	if p2.Target != "level1" || p2.TargetPos != (geom.Point{X: 33, Y: 3}) {
		t.Fatalf("portal 2 = %+v", p2)
	}
}

func TestParseLevelErrors(t *testing.T) {
	tooWide := "id: a\n---\n" + strings.Repeat(".", MaxLevelSize+1) + "\n"
	tooTall := "id: a\n---\n" + strings.Repeat(".\n", MaxLevelSize+1)
	for name, src := range map[string]string{
		"no separator": "id: a\n",
		"no id":        "name: x\n---\n.\n",
		"bad glyph":    "id: a\n---\n?\n",
		"too wide":     tooWide,
		"too tall":     tooTall,
	} {
		if _, _, err := ParseLevel(src); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
	if _, _, err := ParseLevel("id: a\n---\n" + strings.Repeat(".", MaxLevelSize) + "\n"); err != nil {
		t.Errorf("max size level rejected: %v", err)
	}
}

func TestPortalMustBeOnStairs(t *testing.T) {
	_, _, err := ParseLevel("id: a\nportal: 0,0 -> b:0,0\n---\n@\n")
	if err != nil {
		t.Fatal(err) // parse succeeds; linking validates
	}
	l, ds, _ := ParseLevel("id: a\nportal: 0,0 -> a:1,0\n---\n@>\n")
	w := &World{Levels: map[LevelID]*Level{"a": l}}
	if err := w.link(ds[0]); err == nil {
		t.Fatal("expected error for portal not on stairs")
	}
}
