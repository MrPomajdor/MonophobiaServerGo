package entity

import (
	"MonophobiaServer/internal/messages"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"reflect"

	log "github.com/sirupsen/logrus"
)

type Packet struct {
	Client         *Client
	Network        Network
	Header         messages.Header
	Flag           messages.Flag
	FullMsgLen     int32
	Payload        []byte
	payloadPointer int32
}

var (
	Endianess = binary.LittleEndian
)

type Network int

const (
	NET_TCP Network = 0
	NET_UDP Network = 1
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
	buf := bytes.NewReader(*data)
	//packet.Header = messages.Header(binary.BigEndian.Uint16((*data)[0:2]))
	if err := binary.Read(buf, binary.BigEndian, &packet.Header); err != nil {
		return fmt.Errorf("failed reading packet header %w", err)
	}
	packet.Flag = messages.Flag((*data)[2])
	buf.Seek(1, io.SeekCurrent)

	//packet.FullMsgLen = binary.LittleEndian.Uint32((*data)[3:7])

	if err := binary.Read(buf, binary.LittleEndian, &packet.FullMsgLen); err != nil {
		return fmt.Errorf("failed reading packet message length %w", err)
	}

	if packet.FullMsgLen < 7 || packet.FullMsgLen > (int32)(len(*data)) {
		return fmt.Errorf("data corrupted: invalid payload : %d while data slice len is %d", packet.FullMsgLen, (int32)(len(*data)))
	}
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

func (packet *Packet) SendUDP(conn net.UDPConn) {
	payload, err := packet.assembleMessage()
	if err != nil {
		log.WithField("error", err.Error()).Error("Failed to assemble packet")
		return
	}
	conn.Write(payload)
}

func (packet *Packet) AddToPayload(data interface{}) error {
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return fmt.Errorf("data must be a pointer to a struct, got %T", data)
	}
	buffer := bytes.NewBuffer(packet.Payload)
	if err := serializeData(buffer, data); err != nil {
		return fmt.Errorf("failed to serialize data : %w", err)
	}

	packet.Payload = buffer.Bytes()
	return nil
}
func serializeData(buffer *bytes.Buffer, data interface{}) error {
	//log.Trace("Serialization start")
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return fmt.Errorf("data must be a pointer to a struct, got %T", data)
	}
	v := reflect.ValueOf(data).Elem()
	//log.WithFields(log.Fields{"Struct len": v.NumField()}).Trace("Serializing data")
	for i := 0; i < v.NumField(); i += 1 {
		field := v.Field(i)
		//log.WithFields(log.Fields{"Type": field.Kind()}).Trace("Serializing data")
		switch field.Kind() {
		case reflect.String:
			var strlen int32
			strlen = (int32)(len(field.String()))
			if err := binary.Write(buffer, Endianess, strlen); err != nil {
				return fmt.Errorf("failed to write string length for field %d: %w", i, err)
			}
			if err := binary.Write(buffer, Endianess, ([]byte)(field.String())); err != nil {
				return fmt.Errorf("failed to write string data for field %d: %w", i, err)
			}
		case reflect.Int32:
			if err := binary.Write(buffer, Endianess, (int32)(field.Int())); err != nil {
				return fmt.Errorf("failed to write data for field %d: %w", i, err)
			}
		case reflect.Float32:
			if err := binary.Write(buffer, Endianess, (float32)(field.Float())); err != nil {
				return fmt.Errorf("failed to write data for field %d: %w", i, err)
			}
		case reflect.Bool:
			if err := binary.Write(buffer, Endianess, field.Bool()); err != nil {
				return fmt.Errorf("failed to write data for field %d: %w", i, err)
			}
		case reflect.Struct:
			if err := serializeData(buffer, field.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to serialize struct for field %d: %w", i, err)
			}
		case reflect.Pointer:
			if field.Elem().Kind() == reflect.Struct {
				if err := serializeData(buffer, field.Elem().Interface()); err != nil {
					return fmt.Errorf("failed to serialize struct pointer for field %d: %w", i, err)
				}
			} else {
				return fmt.Errorf("unsupported pointer type: %v", field.Type())
			}
		case reflect.Slice:
			var slicelen int32
			slicelen = (int32)(field.Len())
			//log.WithFields(log.Fields{"slicelen": slicelen}).Trace("slice len")

			if err := binary.Write(buffer, Endianess, slicelen); err != nil {
				return fmt.Errorf("failed to write slice length for field %d: %w", i, err)
			}
			for k := 0; k < int(slicelen); k += 1 {
				//log.WithFields(log.Fields{"Type": field.Type().Elem().Kind()}).Trace("Serializing slice element")

				switch field.Type().Elem().Kind() {
				case reflect.Pointer:
					if field.Index(k).Elem().Kind() == reflect.Struct {
						if err := serializeData(buffer, field.Index(k).Elem().Addr().Interface()); err != nil {
							return fmt.Errorf("failed to serialize struct pointer for slize element %d: %w", i, err)
						}
					} else {
						return fmt.Errorf("unsupported pointer type: %v", field.Type())
					}
				case reflect.Struct:
					//log.WithFields(log.Fields{"Type": field.Type().Elem().Kind(), "xd": field.Index(k)}).Trace("dupsko serialization")
					if err := serializeData(buffer, field.Index(k).Addr().Interface()); err != nil {
						return fmt.Errorf("failed to serialize slize element %d: %w", k, err)
					}
				case reflect.String:
					var strlen int32
					strlen = (int32)(len(field.Index(k).String()))
					if err := binary.Write(buffer, Endianess, strlen); err != nil {
						return fmt.Errorf("failed to write string length for slize element %d: %w", i, err)
					}
					if err := binary.Write(buffer, Endianess, ([]byte)(field.Index(k).String())); err != nil {
						return fmt.Errorf("failed to write string data for slize element %d: %w", i, err)
					}
				case reflect.Int32:
					if err := binary.Write(buffer, Endianess, (int32)(field.Index(k).Int())); err != nil {
						return fmt.Errorf("failed to write data for slize element %d: %w", i, err)
					}
				case reflect.Float32:
					if err := binary.Write(buffer, Endianess, (float32)(field.Index(k).Float())); err != nil {
						return fmt.Errorf("failed to write data for slize element %d: %w", i, err)
					}
				case reflect.Bool:
					if err := binary.Write(buffer, Endianess, field.Index(k).Bool()); err != nil {
						return fmt.Errorf("failed to write data for slize element %d: %w", i, err)
					}
				default:
					return fmt.Errorf("unsupported slice field type: %v", field.Index(k).Kind())
				}
			}
		default:
			return fmt.Errorf("unsupported field type: %v", field.Kind())
		}
	}
	return nil
}
func (packet *Packet) ReadPayload(out interface{}) error {
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		return fmt.Errorf("out must be a pointer to a struct, got %T", out)
	}
	r := bytes.NewReader(packet.Payload)
	return deserializeData(r, out)
}

func deserializeData(reader *bytes.Reader, out interface{}) error {
	v := reflect.ValueOf(out).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		switch field.Kind() {
		case reflect.String:
			// Read int32 length for the string
			var strlen int32
			if err := binary.Read(reader, Endianess, &strlen); err != nil {
				return fmt.Errorf("failed to read string length for field %d: %w", i, err)
			}
			// Read the actual string bytes
			buf := make([]byte, strlen)
			if _, err := io.ReadFull(reader, buf); err != nil {
				return fmt.Errorf("failed to read string data for field %d: %w", i, err)
			}
			field.SetString(string(buf))
		case reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64, reflect.Bool:
			// Directly read these supported types
			if err := binary.Read(reader, Endianess, field.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to read field %d: %w", i, err)
			}
		case reflect.Struct:
			deserializeData(reader, field.Addr().Interface())
		case reflect.Slice:
			elemType := field.Type().Elem()
			var slicelen int32
			if err := binary.Read(reader, Endianess, &slicelen); err != nil {
				return fmt.Errorf("failed to read slice length for field %d: %w", i, err)
			}
			field.Set(reflect.MakeSlice(field.Type(), int(slicelen), int(slicelen)))
			for j := 0; j < int(slicelen); j += 1 {
				elem := field.Index(j)
				switch elemType.Kind() {
				case reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64, reflect.Bool:
					// Directly read these supported types
					if err := binary.Read(reader, Endianess, elem.Addr().Interface()); err != nil {
						return fmt.Errorf("failed to read field %d: %w", j, err)
					}
				case reflect.String:
					// Read int32 length for the string
					var strlen int32
					if err := binary.Read(reader, Endianess, &strlen); err != nil {
						return fmt.Errorf("failed to read string length for field %d: %w", j, err)
					}
					// Read the actual string bytes
					buf := make([]byte, strlen)
					if _, err := io.ReadFull(reader, buf); err != nil {
						return fmt.Errorf("failed to read string data for field %d: %w", j, err)
					}
					elem.SetString(string(buf))
				case reflect.Struct:
					deserializeData(reader, elem)
				}
			}
		default:
			return fmt.Errorf("unsupported field type: %v", field.Kind())
		}
	}
	return nil
}
