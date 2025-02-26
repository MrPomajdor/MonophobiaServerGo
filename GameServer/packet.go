package GameServer

import (
	"MonophobiaServer/messages"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"reflect"

	log "github.com/sirupsen/logrus"
)

var (
	Endianess = binary.LittleEndian
)

func (packet *Packet) assembleMessage() ([]byte, error) {
	var result []byte //make([]byte,7+len(packet.Payload))
	result, _ = binary.Append(result, binary.BigEndian, (int16)(packet.Header))
	result, _ = binary.Append(result, binary.BigEndian, (byte)(packet.Flag))
	result, _ = binary.Append(result, binary.LittleEndian, int32(len(packet.Payload)))
	result = append(result, packet.Payload...)
	return result, nil
	//TODO: Check for invalid packet data

}

func (packet *Packet) DigestData(data *[]byte) error {
	if len(*data) < 4 {
		log.Debug("Tried digesting data with less than 4 data length!")
		return fmt.Errorf("Packet too short")
	}

	packet.Header = messages.Header(binary.BigEndian.Uint16((*data)[0:2]))

	packet.Flag = messages.Flag((*data)[2])

	packet.FullMsgLen = binary.LittleEndian.Uint32((*data)[3:7])

	packet.Payload = (*data)[7:packet.FullMsgLen]
	packet.payloadPointer = 0
	return nil
}

func (packet *Packet) AddString(value string) {
	packet.Payload, _ = binary.Append(packet.Payload, binary.LittleEndian, (int32)(len(value)))
	packet.Payload = append(packet.Payload, []byte(value)...)
}

func (packet *Packet) AddFloat(value float32) {
	packet.Payload = append(packet.Payload, byte(math.Float32bits(value)))
}
func (packet *Packet) AddBool(value bool) {
	if value {
		packet.Payload = append(packet.Payload, 0x01)
	} else {
		packet.Payload = append(packet.Payload, 0x00)
	}
}
func (packet *Packet) AddInt(value int32) {
	packet.Payload, _ = binary.Append(packet.Payload, binary.LittleEndian, value)
}

func (packet *Packet) Send(conn net.Conn) {
	payload, err := packet.assembleMessage()
	if err != nil {
		log.WithField("error", err.Error()).Error("Failed to assemble packet")
		return
	}
	conn.Write(payload)
}

func (packet *Packet) ReadPayload(out interface{}) error {
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		return fmt.Errorf("out must be a pointer to a struct, got %T", out)
	}
	r := bytes.NewReader(packet.Payload)
	v := reflect.ValueOf(out).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		switch field.Kind() {
		case reflect.String:
			// Read int32 length for the string
			var strlen int32
			if err := binary.Read(r, Endianess, &strlen); err != nil {
				return fmt.Errorf("failed to read string length for field %d: %w", i, err)
			}
			// Read the actual string bytes
			buf := make([]byte, strlen)
			if _, err := io.ReadFull(r, buf); err != nil {
				return fmt.Errorf("failed to read string data for field %d: %w", i, err)
			}
			field.SetString(string(buf))
		case reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64, reflect.Bool:
			// Directly read these supported types
			if err := binary.Read(r, Endianess, field.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to read field %d: %w", i, err)
			}
		default:
			return fmt.Errorf("unsupported field type: %v", field.Kind())
		}
	}

	return nil
}
