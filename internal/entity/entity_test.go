package entity

import "testing"

func TestMod(t *testing.T) {
	for v, want := range map[int]int{6: -2, 8: -1, 9: -1, 10: 0, 11: 0, 12: 1, 15: 2} {
		if got := Mod(v); got != want {
			t.Errorf("Mod(%d) = %d, want %d", v, got, want)
		}
	}
}

func TestKnightAndGoblin(t *testing.T) {
	k := NewPlayer("Hero", Human, Knight)
	if k.ArmorClass() != 15 || k.MaxHP != 24 || k.HP != 24 || k.IsDead() {
		t.Fatalf("knight: ac=%d maxhp=%d hp=%d", k.ArmorClass(), k.MaxHP, k.HP)
	}
	g := NPCTemplates["goblin"].Spawn()
	if g.ArmorClass() != 12 || g.MaxHP != 18 {
		t.Fatalf("goblin: ac=%d maxhp=%d", g.ArmorClass(), g.MaxHP)
	}
	g.HP = 0
	if !g.IsDead() {
		t.Fatal("hp 0 must be dead")
	}
}
