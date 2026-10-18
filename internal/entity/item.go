package entity

// ItemKind is the broad category of an item.
type ItemKind uint8

const (
	KindMisc ItemKind = iota
	KindArmor
	KindWeapon
	KindAmulet
)

// Slot is a place on the body where an item can be worn.
type Slot uint8

const (
	SlotNone Slot = iota
	SlotBody
)

// ItemDef is the static definition of an item type.
type ItemDef struct {
	ID         string
	Name       string
	Kind       ItemKind
	Slot       Slot
	ArmorBonus int
}

// Item is one concrete item instance.
type Item struct {
	Def *ItemDef
}

var (
	ChainMail     = &ItemDef{ID: "chain_mail", Name: "chain mail", Kind: KindArmor, Slot: SlotBody, ArmorBonus: 5}
	RaggedLeather = &ItemDef{ID: "ragged_leather", Name: "ragged leather", Kind: KindArmor, Slot: SlotBody, ArmorBonus: 2}
)
