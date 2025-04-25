package server

import (
	"MonophobiaServer/internal/entity"
	"MonophobiaServer/internal/errors"
	"MonophobiaServer/internal/messages"
)

func (s *GameServer) registerFlags(parser *Parser) {
	parser.Register(messages.Post.ChatMessage, func(packet *entity.Packet) {
		player := s.GameManager.GetPlayerByID(packet.Client.LogicPlayer.ID)
		if player == nil {
			packet.Client.RespondError(errors.Unauthorized("NOT_IN_LOBBY"))
			return
		}
		var msg struct {
			Message string
		}
		if err := packet.ReadPayload(&msg); err != nil {
			packet.Client.RespondError(errors.InvalidPacket(""))
			return
		}
		player.Lobby.ChatMessage(player, msg.Message)
	})

	parser.Register(messages.Post.CreateLobby, func(packet *entity.Packet) {
		var createLobbyPayload struct {
			Name                string
			MaxPlayers          int32
			IsPasswordProtected bool
			Password            string
		}

		if err := packet.ReadPayload(&createLobbyPayload); err != nil {
			packet.Client.RespondError(errors.InvalidPacket(""))
			return
		}

		if createLobbyPayload.MaxPlayers < 3 {
			packet.Client.RespondError(errors.InvalidData("MAX_PLAYERS_LESS_THAN_THREE"))
			return
		}

		s.GameManager.CreateLobby(packet.Client, createLobbyPayload.Name, createLobbyPayload.MaxPlayers, createLobbyPayload.IsPasswordProtected, createLobbyPayload.Password)
	})
}
