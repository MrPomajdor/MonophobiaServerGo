package game

import (
	"MonophobiaServer/internal/entity"
	"MonophobiaServer/internal/messages"
)

type Lobby struct {
	Manager           *Manager
	Name              string
	Owner             *Player
	MaxPlayers        int32
	Map               string
	PasswordProtected bool
	Password          string
	ID                int32
	Started           bool

	Game *Game
}

func (l *Lobby) ChatMessage(pl *Player, message string) {
	for _, v := range l.Game.Players {

		s_packet := entity.Packet{}
		s_packet.Header = messages.Data
		s_packet.Flag = messages.Response.ChatMessage
		s_packet.AddString(message)

		v.SendPacket(entity.NET_TCP, &s_packet)
	}
}

func (l *Lobby) AddPlayer(pl *Player) {
	l.Manager.RegisterPlayer(pl)
	l.Game.Players = append(l.Game.Players, pl)
}
