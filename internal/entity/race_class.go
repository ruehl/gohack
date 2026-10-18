package entity

// RaceDef is a playable race. Bonus is added to the class' base attributes.
type RaceDef struct {
	ID    string
	Name  string
	Bonus Attributes
}

// ClassDef is a playable class with base attributes, skills and starting items.
type ClassDef struct {
	ID     string
	Name   string
	Attrs  Attributes
	Skills []Skill
	Gear   []*ItemDef // equipped at creation
	Pack   []*ItemDef // carried at creation, not worn
}

var (
	Human = &RaceDef{ID: "human", Name: "Human"}

	Knight = &ClassDef{
		ID: "knight", Name: "Knight",
		Attrs: Attributes{Stamina: 14, Strength: 15, Dexterity: 10, Intelligence: 8},
		Gear:  []*ItemDef{ChainMail, LongSword, WoodenShield},
		Pack: []*ItemDef{
			Dagger, Greatsword, CopperRing, CopperRing, LeatherCap, BoneAmulet,
			LeatherGloves, LeatherGreaves, LeatherBoots,
		},
	}
)
