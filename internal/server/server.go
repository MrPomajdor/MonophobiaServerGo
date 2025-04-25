package server

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"MonophobiaServer/game"
	"MonophobiaServer/internal/entity"
	"MonophobiaServer/internal/errors"
	"MonophobiaServer/internal/messages"
	"MonophobiaServer/internal/state"

	log "github.com/sirupsen/logrus"
)

type GameServer struct {
	bufLen           int
	IP               net.IP
	Port             int
	GameVersion      string
	ServerState      *state.ServerState
	UDPConnectionMap map[string]*entity.Client

	udpConn      *net.UDPConn
	GameManager  *game.Manager
	PacketParser *Parser
}

func (s *GameServer) SetAddress(IP string, Port int) {
	s.IP = net.ParseIP(IP)
	s.Port = Port
}

func (s *GameServer) Start() {
	s.bufLen = 4086
	log.WithField("Buffer", s.bufLen).Debug("Buffer len set")

	go s.bindTCP()
	s.udpConn = s.bindUDP()
	s.UDPConnectionMap = make(map[string]*entity.Client)

	s.GameManager = &game.Manager{}

	pac_chan := make(chan (*game.MessagePackage), 1024)
	s.GameManager.ResponsePacketChannel = &pac_chan
	go s.responsePacketsListen(&pac_chan)

	s.PacketParser = s.initializeParser()

	log.Info("Started server!")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs // Blocks until a signal is received

	log.Info("Shutting down gracefully...")

	os.Exit(0)
}

func (s *GameServer) initializeParser() *Parser {
	packet_parser := Parser{}
	return &packet_parser
}

func (s *GameServer) responsePacketsListen(pac *chan (*game.MessagePackage)) {
	for {
		pac := <-*pac
		packet := pac.Packet
		if cl, ok := s.ServerState.ClientIDMap[pac.Target.ID]; ok {
			packet.Client = cl
			s.SendPacket(packet.Network, packet)
		} else {
			log.WithFields(log.Fields{"id": pac.Target.ID}).Debug("Error while sending packet: Player not found")
		}
	}
}
func (s *GameServer) SendPacket(over entity.Network, packet *entity.Packet) {
	switch over {
	case entity.NET_TCP:
		packet.Send(*packet.Client.Conn)
	case entity.NET_UDP:
		packet.SendUDP(*s.udpConn)
	}
}
func (s *GameServer) bindUDP() *net.UDPConn {
	log.WithFields(log.Fields{"server_ip": s.IP, "server_port": s.Port}).Debug("Starting UDP server")
	address := net.UDPAddr{}
	address.IP = s.IP
	address.Port = int(s.Port)
	ln, err := net.ListenUDP("udp4", &address)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to bind UDP port")
		return nil
	}
	go s.ReadFromUDP(ln)
	return ln
}
func (s *GameServer) ReadFromUDP(ln *net.UDPConn) {
	defer ln.Close()

	log.Trace("Succesfully bound to UDP port")
	buf := make([]byte, 2048)
	for {
		clear(buf)
		msglen, addr, err := ln.ReadFromUDP(buf)
		if !(msglen > 0) {
			continue
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error(), "IP": addr.IP.String()}).Warn("Error receiving UDP")
			continue
		}
		// log.Info(buf)
		// buf = buf[:msglen]
		if bytes.HasPrefix(buf, []byte("holepunch")) {
			continue
		}
		packet := &entity.Packet{}
		err = packet.DigestData(&buf)
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error(), "IP": addr.IP.String()}).Trace("Error digesting udp packet data")
			continue
		}
		// log.Info(packet.Payload)
		if packet.Header == messages.ImHere {
			var imHerePacket struct {
				ID int32
			}
			err = packet.ReadPayload(&imHerePacket)
			if err != nil {
				log.WithField("error", err.Error()).Debug("Failed to read imhere packet")
				continue
			}
			for _, plc := range s.ServerState.Clients {
				if plc.UDPPort == addr.Port && plc.IP == string(addr.IP) {
					break
				}
				if (plc.LogicPlayer != nil) && (plc.UDPPort == -1) && (plc.LogicPlayer.ID == imHerePacket.ID) {
					plc.UDPPort = addr.Port
					s.UDPConnectionMap[strconv.FormatInt((int64)(plc.UDPPort), 10)+":"+plc.IP] = plc
					log.WithFields(log.Fields{"PlayerID": plc.LogicPlayer.ID, "New_UDP_Port": plc.UDPPort}).Trace("Player initialized UDP port")
					break
				}
			}
			continue

		}
		if client, ok := s.UDPConnectionMap[strconv.FormatInt((int64)(addr.Port), 10)+":"+addr.IP.String()]; ok {
			packet.Client = client
			s.PacketParser.Parse(packet)
		} else {
			log.WithFields(log.Fields{"IP": addr.IP.String(), "Port": addr.Port}).Trace("Got UDP data that doesnt match any client")
		}
	}
}
func (s *GameServer) bindTCP() {
	log.WithFields(log.Fields{"server_ip": s.IP.String(), "server_port": s.Port}).Debug("Starting TCP server")
	address := s.IP.String() + ":" + strconv.FormatInt((int64)(s.Port), 10)

	ln, err := net.Listen("tcp4", address)
	if err != nil {
		log.Fatal(err)
	}
	log.Trace("Succesfully bound to TCP port")
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Trace("Error accepting connection: " + err.Error())
			continue
		}
		log.WithFields(log.Fields{"IP": conn.RemoteAddr().String()}).Trace("Accepted connection")
		go s.handleConnection(conn)
	}
}

func (s *GameServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	clientInitialized := false
	LocalClient := entity.NewClient()
	LocalClient.Conn = &conn
	LocalClient.IP = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	buf := make([]byte, s.bufLen)

	for {
		clear(buf)
		var receivedData []byte
		nbytes, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Warn("Error receiving from " + conn.RemoteAddr().String() + ": " + err.Error())
			break
		}
		receivedData = append(receivedData, buf...)
		if nbytes > 7 {
			provLen := binary.LittleEndian.Uint32(buf[3:7])
			if provLen > uint32(s.bufLen) {
				log.Trace("Provlen ", provLen)

				blocks := int(math.Ceil(float64(provLen) / float64(s.bufLen)))
				log.Trace("blocks ", blocks)
				for i := 0; i < blocks; i++ {
					clear(buf)
					nbytes, _ := conn.Read(buf)
					if nbytes < s.bufLen && i != blocks-1 {
						log.Debug("Invalid TCP Data Message Len")
						return
					}
					receivedData = append(receivedData, buf...)
				}
			}
		}
		// log.WithFields(log.Fields{"nbytes": nbytes, "received": receivedData, "buf": buf}).Info("Received")
		packet := &entity.Packet{}

		err = packet.DigestData(&receivedData)
		if err != nil {

			log.WithFields(log.Fields{"IP": conn.RemoteAddr().String(), "err": err.Error()}).Trace("Failed to digest packet")
			continue
		}
		if !clientInitialized {
			if packet.Header != messages.Hello {
				LocalClient.SendDisconnect(errors.ExpectedPacket("HELLO"))
				log.WithFields(log.Fields{"IP": conn.RemoteAddr().String()}).Trace("Rejecting connection - client did not send a valid Hello packet")
				break
			}

			var hello_packet_struct struct {
				Name    string
				SteamID string
				Version string
			}

			err = packet.ReadPayload(&hello_packet_struct)
			if err != nil {
				LocalClient.SendDisconnect(errors.InvalidPacket("HELLO"))
				log.WithFields(log.Fields{"IP": conn.RemoteAddr().String(), "read_error": err.Error()}).Trace("Rejecting client - client sent an invalid Hello packet")
				break
			}
			// packet data is correct
			if hello_packet_struct.Version != s.GameVersion {
				LocalClient.SendDisconnect(errors.DataMismatch("VERSION"))
				log.WithFields(log.Fields{"IP": conn.RemoteAddr().String(), "server_version": s.GameVersion, "client_version": hello_packet_struct.Version}).Trace("Rejecting client - invalid version")
				break
			}

			// initializing client
			pl := s.ServerState.InitializePlayer(hello_packet_struct.Name, LocalClient)
			pl.SteamID = hello_packet_struct.SteamID
			log.WithFields(log.Fields{"Name": pl.Name, "ID": pl.ID}).Debug("Client initialized")
			respPacket := entity.Packet{}
			respPacket.Header = messages.Data
			respPacket.Flag = messages.Response.IDAssign
			var respContent struct {
				ID int32
			}
			respContent.ID = pl.ID
			if err := respPacket.AddToPayload(&respContent); err != nil {
				log.Error(err.Error())
			}
			respPacket.Send(conn)
			LocalClient.LogicPlayer = pl
			clientInitialized = true
			continue
		}
		packet.Client = LocalClient
		s.PacketParser.Parse(packet)
	}
	if LocalClient.LogicPlayer == nil {
		return
	}
	log.WithFields(log.Fields{"name": LocalClient.LogicPlayer.Name, "id": LocalClient.LogicPlayer.ID}).Trace("Player disconnected")
	delete(s.UDPConnectionMap, strconv.FormatInt((int64)(LocalClient.LogicPlayer.NetworkClient.UDPPort), 10)+":"+LocalClient.IP)
	s.ServerState.RemoveClient(LocalClient)

}
