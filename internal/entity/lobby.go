package entity

import (
	"MonophobiaServer/internal/messages"
	"fmt"
	"slices"
	"time"

	log "github.com/sirupsen/logrus"
)

type LobbyMessage int

const (
	LobbyShutdown LobbyMessage = iota
	LobbyRefresh
)

type Lobby struct {
	Owner             *LogicPlayer
	Name              string
	Players           []*LogicPlayer
	Map               string
	MapSeed           int32
	MaxPlayers        int32
	PasswordProtected bool
	Password          string
	ID                int32
	Started           bool
	WorldState        WorldState
	LogicChannel      chan Packet
	MessageChannel    chan LobbyMessage
	TickRate          time.Duration
}

func (l *Lobby) ToNetwork() *NetworkLobbyInfo {
	inf := &NetworkLobbyInfo{}
	inf.LobbyName = l.Name
	inf.MapName = l.Map
	inf.Time = 0
	inf.Players = make([]NetworkPlayerInfo, len(l.Players))
	for i, pl := range l.Players {
		inf.Players[i] = *pl.ToNetwork()
	}
	return inf
}

type NetworkLobbyInfo struct {
	LobbyName string
	MapName   string
	Time      int32
	Players   []NetworkPlayerInfo
}

func (l *Lobby) BroadcastInfo() {
	pac := Packet{}
	pac.Header = messages.Data
	pac.Flag = messages.Response.LobbyInfo
	if err := pac.AddToPayload(l.ToNetwork()); err != nil {
		log.WithField("Error", err.Error()).Error("Adding lobby data to packet failed")
	}
	for _, pl := range l.Players {
		pac.Send(*pl.NetworkClient.Conn)
	}
}

func (l *Lobby) AddPlayer(pl *LogicPlayer) error {
	if l.MaxPlayers <= (int32)(len(l.Players)) {
		return fmt.Errorf("Lobby full")
	}
	log.WithFields(log.Fields{"Player": pl.Name, "Lobby": l.Name}).Trace("Added player to lobby")
	l.Players = append(l.Players, pl)
	l.BroadcastInfo()
	return nil
}

func (l *Lobby) RemovePlayer(pl *LogicPlayer) {
	log.WithFields(log.Fields{"Player": pl.Name, "Lobby": l.Name}).Trace("Removed player from lobby")
	if len(l.Players) == 1 {
		log.WithFields(log.Fields{"Player": pl.Name, "Lobby": l.Name}).Trace("Player was last in lobby")
		l.MessageChannel <- LobbyShutdown
		return
	}
	l.Players = slices.DeleteFunc(l.Players, func(n *LogicPlayer) bool {
		return n == pl
	})
	pl.LogicLobby = nil
	pl.IsHost = false
	l.BroadcastInfo()
}
