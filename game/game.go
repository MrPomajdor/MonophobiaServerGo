package game

import (
	"time"
)

type Game struct {
	Lobby    *Lobby
	TickRate time.Duration

	Players []*Player
}

func (g *Game) Initialize() {
	go g.loop()
}

func (g *Game) loop() {
	ticker := time.NewTicker(g.TickRate)
	for {
		select {
		case <-ticker.C:
			g.checkPlayerMovement()
		}
	}
}
