package main

import (
	"fmt"
	"os"
	"time"

	"gohack/internal/entity"
	"gohack/internal/game"
	"gohack/internal/rules"
	"gohack/internal/ui/term"
	"gohack/internal/world"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gohack:", err)
		os.Exit(1)
	}
}

func run() error {
	w, err := world.LoadEmbedded()
	if err != nil {
		return err
	}
	g, err := game.New(w, rules.NewRNG(uint64(time.Now().UnixNano())))
	if err != nil {
		return err
	}
	player := g.AddPlayer("Hero", entity.Human, entity.Knight)
	return term.Run(g, player.ID)
}
