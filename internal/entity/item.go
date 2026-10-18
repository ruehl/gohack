package entity

// ItemKind is the broad category of an item.
type ItemKind uint8

const (
	KindMisc ItemKind = iota
	KindArmor
	KindWeapon
	KindAmulet
	KindRing
	KindShield
)

// Slot is a place on the body where an item can be worn.
type Slot uint8

const (
	SlotNone Slot = iota
	SlotHead
	SlotNeck
	SlotRing1
	SlotRing2
	SlotBody
	SlotHands
	SlotLegs
	SlotFeet
	SlotMainHand
	SlotOffHand
)

// Slots lists every wearable slot in display order.
var Slots = []Slot{
	SlotHead, SlotNeck, SlotRing1, SlotRing2, SlotBody,
	SlotHands, SlotLegs, SlotFeet, SlotMainHand, SlotOffHand,
}

// Dice is N dice with the given number of sides, e.g. 2d6.
type Dice struct{ N, Sides int }

// ItemDef is the static definition of an item type. Slot is where the item
// normally goes; rings may also go in the second ring slot.
type ItemDef struct {
	ID         string
	Name       string
	Kind       ItemKind
	Slot       Slot
	ArmorBonus int
	TwoHanded  bool // weapons only
	Damage     Dice // weapons only
}

// Item is one concrete item instance.
type Item struct {
	Def *ItemDef
}

var (
	ChainMail     = &ItemDef{ID: "chain_mail", Name: "chain mail", Kind: KindArmor, Slot: SlotBody, ArmorBonus: 5}
	RaggedLeather = &ItemDef{ID: "ragged_leather", Name: "ragged leather", Kind: KindArmor, Slot: SlotBody, ArmorBonus: 2}

	LeatherCap     = &ItemDef{ID: "leather_cap", Name: "leather cap", Kind: KindArmor, Slot: SlotHead, ArmorBonus: 1}
	LeatherGloves  = &ItemDef{ID: "leather_gloves", Name: "leather gloves", Kind: KindArmor, Slot: SlotHands, ArmorBonus: 1}
	LeatherGreaves = &ItemDef{ID: "leather_greaves", Name: "leather greaves", Kind: KindArmor, Slot: SlotLegs, ArmorBonus: 1}
	LeatherBoots   = &ItemDef{ID: "leather_boots", Name: "leather boots", Kind: KindArmor, Slot: SlotFeet, ArmorBonus: 1}
	WoodenShield   = &ItemDef{ID: "wooden_shield", Name: "wooden shield", Kind: KindShield, Slot: SlotOffHand, ArmorBonus: 1}

	BoneAmulet = &ItemDef{ID: "bone_amulet", Name: "bone amulet", Kind: KindAmulet, Slot: SlotNeck}
	CopperRing = &ItemDef{ID: "copper_ring", Name: "copper ring", Kind: KindRing, Slot: SlotRing1}

	Dagger     = &ItemDef{ID: "dagger", Name: "dagger", Kind: KindWeapon, Slot: SlotMainHand, Damage: Dice{1, 4}}
	LongSword  = &ItemDef{ID: "long_sword", Name: "long sword", Kind: KindWeapon, Slot: SlotMainHand, Damage: Dice{1, 8}}
	Greatsword = &ItemDef{ID: "greatsword", Name: "greatsword", Kind: KindWeapon, Slot: SlotMainHand, TwoHanded: true, Damage: Dice{2, 6}}
)
