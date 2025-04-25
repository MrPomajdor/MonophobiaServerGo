package entity

import (
	"MonophobiaServer/internal/errors"
	"MonophobiaServer/internal/messages"
	"net"
)

type Client struct {
	Conn        *net.Conn
	UDPPort     int
	IP          string
	LogicPlayer *LogicPlayer
}

func NewClient() *Client {
	cl := &Client{}
	cl.UDPPort = -1
	return cl
}

func (c *Client) RespondError(msg errors.Error) {
	respPacket := Packet{}
	respPacket.Header = messages.Rejected

	respPacket.Flag = messages.None

	respPacket.AddString(msg.Message)
	respPacket.AddString(msg.Description)
	respPacket.Send(*c.Conn)
}

func (c *Client) SendDisconnect(msg errors.Error) {
	respPacket := Packet{}
	respPacket.Header = messages.Disconnecting

	respPacket.Flag = messages.None

	respPacket.AddString(msg.Message)
	respPacket.AddString(msg.Description)
	respPacket.Send(*c.Conn)
}
