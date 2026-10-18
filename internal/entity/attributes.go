// Package entity defines players and NPCs: attributes, items and their data tables.
package entity

// Attributes are the skill stats of an entity.
type Attributes struct {
	Stamina      int
	Strength     int
	Dexterity    int
	Intelligence int
}

// Add returns the field-wise sum of a and b.
func (a Attributes) Add(b Attributes) Attributes {
	return Attributes{
		a.Stamina + b.Stamina,
		a.Strength + b.Strength,
		a.Dexterity + b.Dexterity,
		a.Intelligence + b.Intelligence,
	}
}

// Mod converts an attribute value to its modifier: (v-10)/2, rounded down.
func Mod(v int) int {
	d := v - 10
	if d < 0 {
		d-- // floor division for negatives
	}
	return d / 2
}
