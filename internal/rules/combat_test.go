package rules

import (
	"testing"

	"gohack/internal/entity"
)

// seq is an RNG returning fixed values.
type seq []int

func (s *seq) IntN(int) int { v := (*s)[0]; *s = (*s)[1:]; return v }

func TestBareHands(t *testing.T) {
	k := entity.NewPlayer("Hero", entity.Human, entity.Knight) // DEX 10 (+0), STR 15 (+2)
	g := entity.NPCTemplates["goblin"].Spawn()                 // AC 12

	// roll 11 (IntN=10) + 0 < 12: miss.
	if r := BareHands(k, g, &seq{10}); r.Hit {
		t.Fatalf("expected miss, got %+v", r)
	}
	// roll 12 meets AC 12: hit, damage 1+1+2 = 4.
	if r := BareHands(k, g, &seq{11, 1}); !r.Hit || r.Damage != 4 {
		t.Fatalf("expected hit for 4, got %+v", r)
	}
	// goblin vs knight: DEX 12 (+1), AC 15 needs roll 14.
	if r := BareHands(g, k, &seq{12}); r.Hit {
		t.Fatalf("roll 13+1 < 15 must miss, got %+v", r)
	}
	// goblin STR 8 (-1): damage 1+0-1 = 0, floored to 1.
	if r := BareHands(g, k, &seq{19, 0}); !r.Hit || r.Damage != 1 {
		t.Fatalf("expected min damage 1, got %+v", r)
	}
}
