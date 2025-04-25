package server

import (
	"fmt"

	"MonophobiaServer/internal/entity"
	"MonophobiaServer/internal/messages"
)

type FlagReceiver func(packet *entity.Packet)

type Parser struct {
	flagMap map[messages.Flag]FlagReceiver
}

// Parse incoming packet
func (p *Parser) Parse(packet *entity.Packet) error {
	if receiver, ok := p.flagMap[packet.Flag]; ok {
		receiver(packet)
		return nil
	}
	return fmt.Errorf("flag's %X receiver not found", packet.Flag)
}

// Register a flag receiver function
func (p *Parser) Register(flag messages.Flag, receiver FlagReceiver) error {
	if _, ok := p.flagMap[flag]; !ok {
		p.flagMap[flag] = receiver
		return nil
	}
	return fmt.Errorf("flag %X already registered", flag)
}
