package entity

// NPCTemplate is the static definition an NPC is spawned from.
type NPCTemplate struct {
	ID          string
	Name        string
	Glyph       rune
	Attrs       Attributes
	Disposition Disposition
	Gear        []*ItemDef // equipped at spawn
}

// NPCTemplates is the NPC registry, keyed by template ID.
var NPCTemplates = map[string]*NPCTemplate{
	"goblin": {
		ID: "goblin", Name: "goblin", Glyph: 'g',
		Attrs:       Attributes{Stamina: 8, Strength: 8, Dexterity: 12, Intelligence: 6},
		Disposition: Hostile,
		Gear:        []*ItemDef{RaggedLeather},
	},
}

// Spawn builds an NPC entity from the template. The caller assigns ID, level and position.
func (t *NPCTemplate) Spawn() *Entity {
	e := &Entity{
		Name: t.Name, Template: t.ID, Glyph: t.Glyph,
		Attrs: t.Attrs, Disposition: t.Disposition,
		Equipment: map[Slot]*Item{},
	}
	e.MaxHP = MaxHPFor(e.Attrs)
	e.HP = e.MaxHP
	for _, g := range t.Gear {
		e.Equip(&Item{Def: g})
	}
	return e
}
