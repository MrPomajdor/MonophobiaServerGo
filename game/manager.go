package game

import (
	"MonophobiaServer/internal/entity"
)

type MessagePackage struct {
	Packet *entity.Packet
	Target *Player
}

type Manager struct {
	Lobbies               []*Lobby
	ResponsePacketChannel *chan (*MessagePackage)
	PlayerIDMap           map[int32]*Player
}

func (m *Manager) RegisterPlayer(player *Player) {
	m.PlayerIDMap[player.ID] = player
}

func (m *Manager) UnregisterPlayer(player *Player) {
	delete(m.PlayerIDMap, player.ID)
}

func (m *Manager) GetPlayerByID(id int32) *Player {
	return m.PlayerIDMap[id]
}

func (m *Manager) CreateLobby(creator *entity.LogicPlayer, name string, maxPlayers int32, passwordProtected bool, password string) {
	pl := m.createPlayer(creator)
	lb := m.createLobby(name, maxPlayers, passwordProtected, password)

	lb.Owner = pl
}

func (m *Manager) createPlayer(netPlayer *entity.LogicPlayer) *Player {
	newPlayer := &Player{}
	newPlayer.ID = newPlayer.ID
	newPlayer.UserData = netPlayer.UserData

	m.RegisterPlayer(newPlayer)

	return newPlayer
}

func (m *Manager) createLobby(name string, maxPlayers int32, passProtected bool, pass string) *Lobby {
	l := &Lobby{}
	l.Name = name
	l.MaxPlayers = maxPlayers
	l.PasswordProtected = passProtected
	l.Password = pass
	l.Game = &Game{}
	l.Game.Lobby = l

	return l
}
