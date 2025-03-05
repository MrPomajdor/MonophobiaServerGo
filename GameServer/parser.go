package GameServer

import (
	"strconv"
	"time"

	"MonophobiaServer/messages"

	log "github.com/sirupsen/logrus"
)

// Here we only parse and execute actions that are non lobby dependant (these we pass down to the lobby)
func (s *GameServer) ParsePacket(packet *Packet) { //, client *Client) {
	client := packet.Client
	switch packet.Header {
	case messages.Data:
		switch packet.Flag {
		case messages.Post.CreateLobby:
			var createPacketStruct struct {
				Name                string
				MaxPlayers          int32
				IsPasswordProtected bool
				Password            string
			}
			err := packet.ReadPayload(&createPacketStruct)
			if err != nil {
				log.Debug(err.Error())
				client.RespondError("INVALID_PACKET", false)
				return
			}
			if createPacketStruct.MaxPlayers < 3 {
				client.RespondError("MAX_PLAYERS_TOO_SMALL", false)
				return
			}

			newLobby := &Lobby{}
			newLobby.Name = createPacketStruct.Name
			newLobby.MaxPlayers = createPacketStruct.MaxPlayers
			newLobby.PasswordProtected = createPacketStruct.IsPasswordProtected
			newLobby.Password = createPacketStruct.Password
			newLobby.Owner = client.ConnectedPlayer
			newLobby.Map = Maps.Lobby
			newLobby.ID = client.ConnectedPlayer.ID
			client.ConnectedPlayer.Lobby = newLobby
			newLobby.TickRate = time.Millisecond * 20
			s.InitializeLobby(newLobby)

			if err := newLobby.AddPlayer(client.ConnectedPlayer); err != nil {
				log.Debug(err.Error())
			}

			listChanged := Packet{}
			listChanged.Header = messages.Data
			listChanged.Flag = messages.Response.LobbyListChanged

			for _, cl := range s.Clients {
				if cl.ConnectedPlayer.Lobby == nil {
					listChanged.Send(*cl.Conn)
				}
			}
		case messages.Request.LobbyList:
			resp := Packet{}
			resp.Header = messages.Data
			resp.Flag = messages.Response.LobbyList
			resp.AddInt((int32)(len(s.Lobbies)))
			for _, lb := range s.Lobbies {
				resp.AddInt(lb.ID)
				resp.AddString(lb.Name)
				resp.AddBool(lb.PasswordProtected)
				resp.AddInt((int32)(len(lb.Players)))
				resp.AddInt(lb.MaxPlayers)
			}
			resp.Send(*client.Conn)
		case messages.Post.JoinLobby:
			if client.ConnectedPlayer.Lobby != nil {
				client.RespondError("ALREADY_IN_LOBBY", false)
				return
			}
			var joinPacket struct {
				LobbyID  int32
				Password string
			}
			if err := packet.ReadPayload(&joinPacket); err != nil {
				client.RespondError("INVALID_PACKET", false)
				return
			}
			for _, lb := range s.Lobbies { // O(n) shouldn't be an issue, right?
				if lb.ID == joinPacket.LobbyID {
					if len(lb.Players) >= int(lb.MaxPlayers) {
						client.RespondError("LOBBY_FULL", false)
						return
					}
					if lb.PasswordProtected && lb.Password != joinPacket.Password {
						client.RespondError("INVALID_PASSWORD", false)
						return
					}
					packet.Client.ConnectedPlayer.Lobby = lb
					lb.AddPlayer(packet.Client.ConnectedPlayer)
					return
				}
			}
			client.RespondError("LOBBY_NOT_FOUND", false)
		case messages.Post.PlayerTransformData, messages.Post.ItemPickup, messages.Post.ItemDrop, messages.Post.ItemIntInf:
			//All this wierdness is because somehoow the payload of the packet was cleared when pulled out of the channel.
			r := *packet
			r.Flag = packet.Flag
			r.Header = packet.Header
			r.Payload = append([]byte{}, packet.Payload...)
			r.Client = packet.Client
			client.ConnectedPlayer.Lobby.LogicChannel <- r
		default:
			log.WithFields(log.Fields{"Flag": strconv.FormatInt((int64)(packet.Flag), 16), "IP": client.IP}).Warn("Flag not recognized")
			client.RespondError("FLAG_NOT_RECOGNIZED", false)

		}
	case messages.Echo:
		resp := Packet{}
		resp.Header = messages.Echo
		resp.Flag = messages.None
		resp.Send(*packet.Client.Conn)
	default:
		log.WithFields(log.Fields{"Header": strconv.FormatInt((int64)(packet.Header), 16), "IP": client.IP}).Warn("Header not recognized")
		client.RespondError("HEADER_NOT_RECOGNIZED", false)
	}
}
