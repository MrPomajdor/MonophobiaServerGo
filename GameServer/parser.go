package GameServer

import (
	"MonophobiaServer/messages"

	log "github.com/sirupsen/logrus"
)

func (s *GameServer) ParsePacket(packet Packet, client *Client) {
	switch packet.Header {
	case messages.Data:
		switch packet.Flag {

		default:
			log.WithFields(log.Fields{"Flag": packet.Flag, "IP": client.IP}).Warn("Flag not recognized")
			respondError(*client.Conn, "FLAG_NOT_RECOGNIZED", false)

		}
	default:
		log.WithFields(log.Fields{"Header": packet.Header, "IP": client.IP}).Warn("Header not recognized")
		respondError(*client.Conn, "HEADER_NOT_RECOGNIZED", false)
	}
}
