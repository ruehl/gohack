package game

import (
	"fmt"
	"slices"

	"gohack/internal/entity"
	"gohack/internal/geom"
	"gohack/internal/rules"
	"gohack/internal/world"
)

type posKey struct {
	Level world.LevelID
	Pos   geom.Point
}

// Game is the complete simulation state.
type Game struct {
	World *world.World
	Turn  int

	rng      rules.RNG
	nextID   entity.EntityID
	entities map[entity.EntityID]*entity.Entity
	order    []entity.EntityID // spawn order, for deterministic NPC turns
	occ      map[posKey]entity.EntityID
	views    map[entity.EntityID]*perception
	logs     map[entity.EntityID][]string
	fields   map[*world.Light]*lightField // lighting cache, see lighting.go
}

// New creates a game over w and spawns the NPCs declared by its levels.
func New(w *world.World, rng rules.RNG) (*Game, error) {
	g := &Game{
		World:    w,
		rng:      rng,
		nextID:   1,
		entities: map[entity.EntityID]*entity.Entity{},
		occ:      map[posKey]entity.EntityID{},
		views:    map[entity.EntityID]*perception{},
		logs:     map[entity.EntityID][]string{},
		fields:   map[*world.Light]*lightField{},
	}
	for _, id := range w.Order {
		l := w.Level(id)
		for _, s := range l.Spawns {
			tpl, ok := entity.NPCTemplates[s.Template]
			if !ok {
				return nil, fmt.Errorf("level %q: unknown NPC template %q", id, s.Template)
			}
			g.place(tpl.Spawn(), id, s.Pos)
		}
	}
	return g, nil
}

// place registers e at the given location.
func (g *Game) place(e *entity.Entity, level world.LevelID, pos geom.Point) {
	e.ID = g.nextID
	g.nextID++
	e.Level, e.Pos = level, pos
	g.entities[e.ID] = e
	g.order = append(g.order, e.ID)
	g.occ[posKey{level, pos}] = e.ID
}

// AddPlayer creates a player at the world's start position.
func (g *Game) AddPlayer(name string, race *entity.RaceDef, class *entity.ClassDef) *entity.Entity {
	e := entity.NewPlayer(name, race, class)
	start := g.World.Level(g.World.Start)
	g.place(e, start.ID, start.Start)
	g.views[e.ID] = &perception{seen: map[world.LevelID]*bitgrid{}}
	g.refreshView(e)
	return e
}

// Entity returns the entity with the given ID, or nil.
func (g *Game) Entity(id entity.EntityID) *entity.Entity { return g.entities[id] }

// EntityAt returns the entity standing at pos on a level, or nil.
func (g *Game) EntityAt(level world.LevelID, pos geom.Point) *entity.Entity {
	return g.entities[g.occ[posKey{level, pos}]]
}

func (g *Game) refreshView(e *entity.Entity) {
	g.views[e.ID].refresh(g.World.Level(e.Level), e.Pos)
}

// Submit applies a command on behalf of an entity. If the command used up the
// actor's turn, NPCs act and the turn counter advances. It returns what happened.
func (g *Game) Submit(actor entity.EntityID, cmd Command) []Event {
	e := g.entities[actor]
	if e == nil || e.IsDead() {
		return nil
	}
	var evs []Event
	var spent bool
	switch c := cmd.(type) {
	case MoveCmd:
		evs, spent = g.move(e, c.Dir)
	case AttackCmd:
		evs, spent = g.attackDir(e, c.Dir)
	case UseStairsCmd:
		evs, spent = g.useStairs(e)
	case EquipCmd:
		evs, spent = g.equip(e, c)
	case UnequipCmd:
		evs, spent = g.unequip(e, c.Slot)
	case ToggleLightCmd:
		evs, spent = g.toggleLight(e, c.Dir)
	}
	if spent {
		evs = append(evs, g.npcPhase()...)
		g.Turn++
	}
	for id := range g.views {
		if p := g.entities[id]; p != nil && !p.IsDead() {
			g.refreshView(p)
		}
	}
	g.record(evs)
	return evs
}

func (g *Game) move(e *entity.Entity, dir geom.Dir) ([]Event, bool) {
	to := e.Pos.Add(dir.Delta())
	l := g.World.Level(e.Level)
	if other := g.EntityAt(e.Level, to); other != nil {
		return []Event{g.event(EvBlocked, e, other, to)}, false
	}
	switch l.At(to) {
	case world.DoorClosed:
		l.Set(to, world.DoorOpen)
		return []Event{g.event(EvDoorOpened, e, nil, to)}, true
	}
	if !l.Walkable(to) {
		return nil, false
	}
	g.moveTo(e, e.Level, to)
	return []Event{g.event(EvMoved, e, nil, to)}, true
}

func (g *Game) moveTo(e *entity.Entity, level world.LevelID, pos geom.Point) {
	delete(g.occ, posKey{e.Level, e.Pos})
	e.Level, e.Pos = level, pos
	g.occ[posKey{level, pos}] = e.ID
}

func (g *Game) attackDir(e *entity.Entity, dir geom.Dir) ([]Event, bool) {
	at := e.Pos.Add(dir.Delta())
	target := g.EntityAt(e.Level, at)
	if target == nil {
		ev := g.event(EvMessage, e, nil, at)
		ev.Text = "There is nothing to attack there."
		return []Event{ev}, false
	}
	return g.attack(e, target), true
}

// attack attacks once with each held weapon (bare hands if there is none),
// stopping when the defender dies, and removes defeated NPCs from the world.
func (g *Game) attack(att, def *entity.Entity) []Event {
	if att.IsPlayer && !def.IsPlayer {
		def.Disposition = entity.Hostile // attacked NPCs turn hostile, even friendly ones
	}
	dmg := []entity.Dice{rules.BareHands}
	if ws := att.Weapons(); len(ws) > 0 {
		dmg = dmg[:0]
		for _, w := range ws {
			dmg = append(dmg, w.Def.Damage)
		}
	}
	var evs []Event
	for _, d := range dmg {
		res := rules.Attack(att, def, d, g.rng)
		ev := g.event(EvAttack, att, def, def.Pos)
		ev.Hit, ev.Damage = res.Hit, res.Damage
		evs = append(evs, ev)
		if res.Hit {
			def.HP -= res.Damage
		}
		if def.IsDead() {
			evs = append(evs, g.event(EvDied, def, nil, def.Pos))
			if !def.IsPlayer {
				g.remove(def)
			}
			break
		}
	}
	return evs
}

// equip wears an inventory item in a slot. Items that had to come off are
// reported first. A refused change costs no turn.
func (g *Game) equip(e *entity.Entity, c EquipCmd) ([]Event, bool) {
	if c.Item < 0 || c.Item >= len(e.Inventory.Items) {
		return g.refuse(e, "You don't have that."), false
	}
	it := e.Inventory.Items[c.Item]
	removed, err := e.EquipAt(it, c.Slot)
	switch err {
	case nil:
	case entity.ErrDoesNotFit:
		return g.refuse(e, "The "+it.Def.Name+" doesn't fit there."), false
	case entity.ErrAlreadyWorn:
		return g.refuse(e, "You already wear the "+it.Def.Name+" there."), false
	default:
		return g.refuse(e, "You can't do that."), false
	}
	var evs []Event
	for _, r := range removed {
		ev := g.event(EvUnequipped, e, nil, e.Pos)
		ev.Text = r.Def.Name
		evs = append(evs, ev)
	}
	ev := g.event(EvEquipped, e, nil, e.Pos)
	ev.Text = it.Def.Name
	return append(evs, ev), true
}

// unequip takes off the item in a slot. An empty slot costs no turn.
func (g *Game) unequip(e *entity.Entity, s entity.Slot) ([]Event, bool) {
	it := e.Unequip(s)
	if it == nil {
		return g.refuse(e, "Nothing is worn there."), false
	}
	ev := g.event(EvUnequipped, e, nil, e.Pos)
	ev.Text = it.Def.Name
	return []Event{ev}, true
}

// toggleLight switches the light source on the adjacent tile in dir on or off.
func (g *Game) toggleLight(e *entity.Entity, dir geom.Dir) ([]Event, bool) {
	at := e.Pos.Add(dir.Delta())
	li := g.World.Level(e.Level).LightAt(at)
	if li == nil {
		return g.refuse(e, "There is no light there."), false
	}
	li.On = !li.On
	kind := EvLightOff
	if li.On {
		kind = EvLightOn
	}
	ev := g.event(kind, e, nil, at)
	ev.Text = li.Def.Name
	return []Event{ev}, true
}

// refuse is the event list for a command that was rejected; Text is shown to e only.
func (g *Game) refuse(e *entity.Entity, text string) []Event {
	ev := g.event(EvMessage, e, nil, e.Pos)
	ev.Text = text
	return []Event{ev}
}

func (g *Game) remove(e *entity.Entity) {
	delete(g.entities, e.ID)
	delete(g.occ, posKey{e.Level, e.Pos})
	g.order = slices.DeleteFunc(g.order, func(id entity.EntityID) bool { return id == e.ID })
}

func (g *Game) useStairs(e *entity.Entity) ([]Event, bool) {
	fail := func(text string) ([]Event, bool) { return g.refuse(e, text), false }
	p, ok := g.World.Level(e.Level).Portals[e.Pos]
	if !ok {
		return fail("There are no stairs here.")
	}
	if g.EntityAt(p.Target, p.TargetPos) != nil {
		return fail("Something blocks the stairs.")
	}
	g.moveTo(e, p.Target, p.TargetPos)
	ev := g.event(EvLevelChanged, e, nil, p.TargetPos)
	ev.Text = g.World.Level(p.Target).Name
	return []Event{ev}, true
}

// event builds an event for actor (and optional target) at pos on the actor's level.
func (g *Game) event(kind EventKind, actor, target *entity.Entity, pos geom.Point) Event {
	ev := Event{Kind: kind, Actor: actor.ID, ActorName: actor.Name, Level: actor.Level, Pos: pos}
	if target != nil {
		ev.Target, ev.TargetName = target.ID, target.Name
	}
	return ev
}
