package GameServer

import (
	"fmt"
	"slices"
	"time"

	"MonophobiaServer/messages"

	log "github.com/sirupsen/logrus"
)

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

func (l *Lobby) AddPlayer(pl *Player) error {
	if l.MaxPlayers <= (int32)(len(l.Players)) {
		return fmt.Errorf("Lobby full")
	}
	log.WithFields(log.Fields{"Player": pl.Name, "Lobby": l.Name}).Trace("Added player to lobby")
	l.Players = append(l.Players, pl)
	l.BroadcastInfo()
	return nil
}

func (l *Lobby) RemovePlayer(pl *Player) {
	log.WithFields(log.Fields{"Player": pl.Name, "Lobby": l.Name}).Trace("Removed player from lobby")
	if len(l.Players) == 1 {
		log.WithFields(log.Fields{"Player": pl.Name, "Lobby": l.Name}).Trace("Player was last in lobby")
		l.MessageChannel <- LobbyShutdown
		return
	}
	l.Players = slices.DeleteFunc(l.Players, func(n *Player) bool {
		return n == pl
	})
	pl.Lobby = nil

	l.BroadcastInfo()
}

func (server *GameServer) InitializeLobby(l *Lobby) {
	log.WithFields(log.Fields{"Name": l.Name, "Owner": l.Owner.Name, "Max_players": l.MaxPlayers, "Password": l.Password}).Trace("Initializing lobby")
	l.LogicChannel = make(chan Packet, 100)
	l.MessageChannel = make(chan LobbyMessage, 30)
	server.Lobbies = append(server.Lobbies, l)
	go l.lobbyLogicLoop(server)
}

func (lobby *Lobby) lobbyLogicLoop(server *GameServer) {
	ticker := time.NewTicker(lobby.TickRate)
	for {
		select {
		case msg := <-lobby.MessageChannel:
			switch msg {
			case LobbyShutdown:

				server.Lobbies = slices.DeleteFunc(server.Lobbies, func(n *Lobby) bool {
					return n == lobby
				})

				log.WithFields(log.Fields{"lobby": lobby.Name, "id": lobby.ID}).Trace("Lobby shutting down")
				ticker.Stop()
				return
			}
		case <-ticker.C:
			lobby.lobbyTick(server)
		case msg := <-lobby.LogicChannel:
			switch msg.Flag {
			case messages.Post.PlayerTransformData:
				var transformPacStruct struct {
					ID         int32
					Transforms Transforms
					Inputs     Inputs
				}
				if err := msg.ReadPayload(&transformPacStruct); err != nil {
					log.WithFields(log.Fields{"Player": msg.Client.ConnectedPlayer.Name, "err": err}).Debug("Invalid player transform packet")
					msg.Client.RespondError("INVALID_PACKET", false)
					return
				}
				msg.Client.ConnectedPlayer.FutureTransforms = transformPacStruct.Transforms
			}

		}
	}
}

func (lobby *Lobby) lobbyTick(server *GameServer) {
	var updatedPlayersPos []PlayerData
	for _, pl := range lobby.Players {
		if pl.Transforms != pl.FutureTransforms {
			// TODO : some calculations
			pl.Transforms = pl.FutureTransforms
			elem := PlayerData{pl.ID, pl.Transforms, pl.PlayerData.Inputs}
			updatedPlayersPos = append(updatedPlayersPos, elem)
		}
	}

	if len(updatedPlayersPos) != 0 && len(lobby.Players) > 1 {
		var playersPosUpdatePacket struct {
			Players []PlayerData
		}
		playersPosUpdatePacket.Players = updatedPlayersPos

		pac := Packet{}
		pac.Header = messages.Data
		pac.Flag = messages.Response.PlayerTransforms
		pac.AddToPayload(&playersPosUpdatePacket)
		//log.Debug(err)
		for _, pl := range lobby.Players {
			pac.Send(*pl.NetworkClient.Conn)
		}
	}
}
