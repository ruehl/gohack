package game

import (
	"gohack/internal/entity"
	"gohack/internal/geom"
	"gohack/internal/world"
)

// MessageLines is how many recent messages a View carries.
const MessageLines = 4

// VisState says how a cell is known to the player.
type VisState uint8

const (
	Unseen VisState = iota
	Remembered
	Visible
)

// ActorView is an entity as shown to a frontend.
type ActorView struct {
	ID          entity.EntityID
	Name        string
	Glyph       rune
	IsPlayer    bool
	Disposition entity.Disposition
}

// Cell is one tile of the viewport. Tile is Void for unseen cells; Actor is
// set only for visible cells.
type Cell struct {
	Tile  world.TileType
	Vis   VisState
	Actor *ActorView
}

// HUD holds the player's stats for display.
type HUD struct {
	Name, Race, Class, Level string
	Attrs                    entity.Attributes
	ArmorClass               int
	HP, MaxHP                int
	Turn                     int
}

// SlotView is one equipment slot; Item is empty when nothing is worn there.
type SlotView struct {
	Slot entity.Slot
	Item string
}

// ItemView is one carried item. Worn is the slot it is worn in (SlotNone if
// carried only) and Fits lists the slots it could be worn in.
type ItemView struct {
	Name string
	Worn entity.Slot
	Fits []entity.Slot
}

// Gear is the player's equipment. Items are in inventory order, so an index
// into Items is the Item of an EquipCmd.
type Gear struct {
	Slots []SlotView
	Items []ItemView
}

// View is everything a frontend needs to draw one player's screen. The
// player is always at cell (W/2, H/2).
type View struct {
	W, H     int
	Cells    []Cell // row-major, len W*H
	HUD      HUD
	Gear     Gear
	Messages []string
	Dead     bool
}

func (v *View) At(x, y int) Cell { return v.Cells[y*v.W+x] }

// BuildView renders the world around a player into a w×h viewport centered on them.
func (g *Game) BuildView(id entity.EntityID, w, h int) *View {
	e := g.entities[id]
	p := g.views[id]
	if e == nil || p == nil || w <= 0 || h <= 0 {
		return nil
	}
	l := g.World.Level(e.Level)
	seen := p.seen[e.Level]
	v := &View{W: w, H: h, Cells: make([]Cell, w*h), Dead: e.IsDead()}

	origin := geom.Point{X: e.Pos.X - w/2, Y: e.Pos.Y - h/2}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			mp := geom.Point{X: origin.X + x, Y: origin.Y + y}
			c := &v.Cells[y*w+x]
			switch {
			case p.visible.Has(mp.X, mp.Y):
				c.Vis, c.Tile = Visible, l.At(mp)
				if a := g.EntityAt(e.Level, mp); a != nil {
					c.Actor = &ActorView{a.ID, a.Name, a.Glyph, a.IsPlayer, a.Disposition}
				}
			case seen.Has(mp.X, mp.Y):
				c.Vis, c.Tile = Remembered, l.At(mp)
			}
		}
	}

	v.HUD = HUD{
		Name: e.Name, Level: l.Name, Attrs: e.Attrs,
		ArmorClass: e.ArmorClass(), HP: e.HP, MaxHP: e.MaxHP, Turn: g.Turn,
	}
	if e.Race != nil {
		v.HUD.Race = e.Race.Name
	}
	if e.Class != nil {
		v.HUD.Class = e.Class.Name
	}
	v.Gear = gearView(e)
	log := g.logs[id]
	v.Messages = log[max(0, len(log)-MessageLines):]
	return v
}

func gearView(e *entity.Entity) Gear {
	var gv Gear
	for _, s := range entity.Slots {
		sv := SlotView{Slot: s}
		if it := e.Equipment[s]; it != nil {
			sv.Item = it.Def.Name
		}
		gv.Slots = append(gv.Slots, sv)
	}
	for _, it := range e.Inventory.Items {
		iv := ItemView{Name: it.Def.Name, Worn: e.SlotOf(it)}
		for _, s := range entity.Slots {
			if e.Fits(it, s) {
				iv.Fits = append(iv.Fits, s)
			}
		}
		gv.Items = append(gv.Items, iv)
	}
	return gv
}
