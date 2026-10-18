package entity

import (
	"gohack/internal/geom"
	"gohack/internal/world"
)

// EntityID uniquely identifies an entity within a game.
type EntityID int

// Disposition is how an NPC feels about the player.
type Disposition uint8

const (
	Neutral Disposition = iota
	Friendly
	Hostile
)

// Inventory holds everything an entity carries, worn or not.
type Inventory struct {
	Items []*Item
}

// Add puts an item into the inventory.
func (inv *Inventory) Add(it *Item) { inv.Items = append(inv.Items, it) }

// Entity is a player or NPC.
type Entity struct {
	ID       EntityID
	Name     string
	Template string // NPC template ID, "player" for players
	Glyph    rune
	IsPlayer bool

	Race  *RaceDef  // nil for NPCs without a race
	Class *ClassDef // nil for NPCs without a class

	Attrs       Attributes
	HP, MaxHP   int
	Disposition Disposition

	Skills    map[Skill]bool
	Inventory Inventory
	Equipment map[Slot]*Item // worn items; every one is also in Inventory

	Level world.LevelID
	Pos   geom.Point
}

// IsDead reports whether the entity has no hitpoints left.
func (e *Entity) IsDead() bool { return e.HP <= 0 }

// Has reports whether the entity has the skill.
func (e *Entity) Has(s Skill) bool { return e.Skills[s] }

func skillSet(skills []Skill) map[Skill]bool {
	set := make(map[Skill]bool, len(skills))
	for _, s := range skills {
		set[s] = true
	}
	return set
}

// ArmorClass is 10 plus the armor bonus of everything worn.
func (e *Entity) ArmorClass() int {
	ac := 10
	for _, it := range e.Equipment {
		ac += it.Def.ArmorBonus
	}
	return ac
}

// MaxHPFor derives maximum hitpoints from attributes.
func MaxHPFor(a Attributes) int { return 10 + a.Stamina }

// NewPlayer builds a player entity. The caller assigns ID, level and position.
func NewPlayer(name string, race *RaceDef, class *ClassDef) *Entity {
	e := &Entity{
		Name: name, Template: "player", Glyph: '@', IsPlayer: true,
		Race: race, Class: class,
		Attrs:     class.Attrs.Add(race.Bonus),
		Skills:    skillSet(class.Skills),
		Equipment: map[Slot]*Item{},
	}
	e.MaxHP = MaxHPFor(e.Attrs)
	e.HP = e.MaxHP
	for _, g := range class.Gear {
		e.Equip(&Item{Def: g})
	}
	for _, p := range class.Pack {
		e.Inventory.Add(&Item{Def: p})
	}
	return e
}
