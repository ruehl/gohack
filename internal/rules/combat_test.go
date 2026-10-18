package rules

import (
	"testing"

	"gohack/internal/entity"
)

// seq is an RNG returning fixed values.
type seq []int

func (s *seq) IntN(int) int { v := (*s)[0]; *s = (*s)[1:]; return v }

func TestAttackBareHands(t *testing.T) {
	k := entity.NewPlayer("Hero", entity.Human, entity.Knight) // DEX 10 (+0), STR 15 (+2)
	g := entity.NPCTemplates["goblin"].Spawn()                 // AC 12

	// roll 11 (IntN=10) + 0 < 12: miss.
	if r := Attack(k, g, BareHands, &seq{10}); r.Hit {
		t.Fatalf("expected miss, got %+v", r)
	}
	// roll 12 meets AC 12: hit, damage 1+1+2 = 4.
	if r := Attack(k, g, BareHands, &seq{11, 1}); !r.Hit || r.Damage != 4 {
		t.Fatalf("expected hit for 4, got %+v", r)
	}
	// goblin vs knight: DEX 12 (+1), AC 16 needs roll 15.
	if r := Attack(g, k, BareHands, &seq{13}); r.Hit {
		t.Fatalf("roll 14+1 < 16 must miss, got %+v", r)
	}
	// goblin STR 8 (-1): damage 1+0-1 = 0, floored to 1.
	if r := Attack(g, k, BareHands, &seq{19, 0}); !r.Hit || r.Damage != 1 {
		t.Fatalf("expected min damage 1, got %+v", r)
	}
}

func TestAttackWeaponDice(t *testing.T) {
	k := entity.NewPlayer("Hero", entity.Human, entity.Knight)
	g := entity.NPCTemplates["goblin"].Spawn()

	// Greatsword 2d6 + STR 2: rolls 4 and 6 -> 10 + 2.
	if r := Attack(k, g, entity.Greatsword.Damage, &seq{19, 3, 5}); !r.Hit || r.Damage != 12 {
		t.Fatalf("expected hit for 12, got %+v", r)
	}
}

func TestRoll(t *testing.T) {
	if got := Roll(entity.Dice{N: 3, Sides: 6}, &seq{0, 2, 5}); got != 1+3+6 {
		t.Fatalf("3d6 = %d, want 10", got)
	}
}
