// Package rules holds game rules that are independent of turn flow and rendering.
package rules

import (
	"math/rand/v2"

	"gohack/internal/entity"
)

// RNG is the source of randomness; inject a seeded one for reproducibility.
type RNG interface {
	IntN(n int) int
}

// NewRNG returns a deterministic RNG for the given seed.
func NewRNG(seed uint64) RNG { return rand.New(rand.NewPCG(seed, 0)) }

// AttackResult is the outcome of a single attack.
type AttackResult struct {
	Roll   int // the d20 roll
	Hit    bool
	Damage int
}

// BareHands is the damage of an unarmed attack.
var BareHands = entity.Dice{N: 1, Sides: 2}

// Roll sums the dice.
func Roll(d entity.Dice, rng RNG) int {
	total := 0
	for range d.N {
		total += rng.IntN(d.Sides) + 1
	}
	return total
}

// Attack resolves one attack (placeholder until the combat system is designed):
// hit if d20 + DEX mod >= target AC; damage dmg + STR mod, at least 1.
// It does not modify the defender.
func Attack(att, def *entity.Entity, dmg entity.Dice, rng RNG) AttackResult {
	roll := rng.IntN(20) + 1
	res := AttackResult{Roll: roll}
	if roll+entity.Mod(att.Attrs.Dexterity) < def.ArmorClass() {
		return res
	}
	res.Hit = true
	res.Damage = max(1, Roll(dmg, rng)+entity.Mod(att.Attrs.Strength))
	return res
}
