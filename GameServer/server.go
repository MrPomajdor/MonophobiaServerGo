package GameServer

import (
	"bytes"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"syscall"

	"MonophobiaServer/messages"

	log "github.com/sirupsen/logrus"
)

type GameServer struct {
	IP               net.IP
	Port             int
	GameVersion      string
	Clients          []*Client
	Lobbies          []*Lobby
	UDPConnectionMap map[string]*Client
}

type Client struct {
	Conn            *net.Conn
	UDPPort         int
	IP              string
	ConnectedPlayer *Player
}

func (c *Client) RespondError(msg string, disconnect bool) {
	respPacket := Packet{}
	if disconnect {
		respPacket.Header = messages.Disconnecting
	} else {
		respPacket.Header = messages.Rejected
	}
	respPacket.Flag = messages.None

	respPacket.AddString(msg)
	respPacket.Send(*c.Conn)
}

func newClient() *Client {
	cl := &Client{}
	cl.UDPPort = -1
	return cl
}

func newPlayer() *Player {
	pl := &Player{}
	pl.ID = -1
	return pl
}

func (s *GameServer) SetAddress(IP string, Port int) {
	s.IP = net.ParseIP(IP)
	s.Port = Port
}

func (s *GameServer) Start() {
	go s.bindTCP()
	go s.bindUDP()

	s.UDPConnectionMap = make(map[string]*Client)

	log.Info("Started server!")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs // Blocks until a signal is received

	log.Info("Shutting down gracefully...")
}

func (s *GameServer) bindUDP() {
	log.WithFields(log.Fields{"server_ip": s.IP, "server_port": s.Port}).Debug("Starting UDP server")
	address := net.UDPAddr{}
	address.IP = s.IP
	address.Port = int(s.Port)
	ln, err := net.ListenUDP("udp4", &address)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to bind UDP port")
		return
	}
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
		packet := &Packet{}
		err = packet.DigestData(&buf)
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error(), "IP": addr.IP.String()}).Trace("Error digesting udp packet data")
			continue
		}
		//log.Info(packet.Payload)
		if packet.Header == messages.ImHere {
			var imHerePacket struct {
				ID int32
			}
			err = packet.ReadPayload(&imHerePacket)
			if err != nil {
				log.WithField("error", err.Error()).Debug("Failed to read imhere packet")
				continue
			}
			for _, plc := range s.Clients {
				if plc.UDPPort == addr.Port && plc.IP == string(addr.IP) {
					break
				}
				if (plc.ConnectedPlayer != nil) && (plc.UDPPort == -1) && (plc.ConnectedPlayer.ID == imHerePacket.ID) {
					plc.UDPPort = addr.Port
					s.UDPConnectionMap[strconv.FormatInt((int64)(plc.UDPPort), 10)+":"+plc.IP] = plc
					log.WithFields(log.Fields{"PlayerID": plc.ConnectedPlayer.ID, "New_UDP_Port": plc.UDPPort}).Trace("Player initialized UDP port")
					break
				}
			}
			continue

		}
		if client, ok := s.UDPConnectionMap[strconv.FormatInt((int64)(addr.Port), 10)+":"+addr.IP.String()]; ok {
			packet.Client = client
			s.ParsePacket(packet)
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
	buf := make([]byte, 2048)
	clientInitialized := false
	LocalClient := newClient()
	LocalClient.Conn = &conn
	LocalClient.IP = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	for {
		clear(buf)
		_, err := conn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Warn("Error receiving from " + conn.RemoteAddr().String() + ": " + err.Error())
			break
		}

		packet := &Packet{}
		err = packet.DigestData(&buf)
		if err != nil {
			log.WithFields(log.Fields{"IP": conn.RemoteAddr().String(), "err": err.Error()}).Trace("Failed to digest packet")
			continue
		}
		if !clientInitialized {
			if packet.Header != messages.Hello {
				LocalClient.RespondError("NO_HELLO", true)
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
				LocalClient.RespondError("INVALID_HELLO", true)
				log.WithFields(log.Fields{"IP": conn.RemoteAddr().String(), "read_error": err.Error()}).Trace("Rejecting client - client sent an invalid Hello packet")
				break
			}
			// packet data is correct
			if hello_packet_struct.Version != s.GameVersion {
				LocalClient.RespondError("INVALID_VERSION", true)
				log.WithFields(log.Fields{"IP": conn.RemoteAddr().String(), "server_version": s.GameVersion, "client_version": hello_packet_struct.Version}).Trace("Rejecting client - invalid version")
				break
			}

			// initializing client
			pl := s.initializePlayer(hello_packet_struct.Name, LocalClient)
			pl.SteamID = hello_packet_struct.SteamID
			log.WithFields(log.Fields{"Name": pl.Name, "ID": pl.ID}).Debug("Client initialized")
			respPacket := Packet{}
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
			LocalClient.ConnectedPlayer = pl
			clientInitialized = true
			continue
		}
		packet.Client = LocalClient
		s.ParsePacket(packet)
	}
	if LocalClient.ConnectedPlayer == nil {
		return
	}
	log.WithFields(log.Fields{"name": LocalClient.ConnectedPlayer.Name, "id": LocalClient.ConnectedPlayer.ID}).Trace("Player disconnected")
	delete(s.UDPConnectionMap, strconv.FormatInt((int64)(LocalClient.ConnectedPlayer.NetworkClient.UDPPort), 10)+":"+LocalClient.IP)
	if LocalClient.ConnectedPlayer.Lobby != nil {
		LocalClient.ConnectedPlayer.Lobby.RemovePlayer(LocalClient.ConnectedPlayer)
	}

	s.Clients = slices.DeleteFunc(s.Clients, func(n *Client) bool {
		return n == LocalClient
	})
}

func (s *GameServer) initializePlayer(name string, client *Client) *Player {
	NewPlayer := newPlayer()
	NewPlayer.Name = name
	NewPlayer.IP = client.IP
	NewPlayer.NetworkClient = client
	client.ConnectedPlayer = NewPlayer
	idValid := true
	for {

		NewPlayer.ID = rand.Int32()
		for _, v := range s.Clients {
			if (v.ConnectedPlayer != nil) && (v.ConnectedPlayer.ID == NewPlayer.ID) {
				idValid = false
				break
			}
			idValid = true
		}
		if idValid {
			break
		}
	}
	s.Clients = append(s.Clients, client)
	return NewPlayer
}
