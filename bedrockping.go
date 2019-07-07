// Package bedrockping is a simple library to ping Minecraft Bedrock/MCPE servers.
package bedrockping

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultPort is the default Minecraft Bedrock/MCPE server port.
	DefaultPort = 19132
)

// Response data returned from ReadUnconnectedPong.
type Response struct {
	Timestamp       uint64   `json:"timestamp"`
	ServerID        uint64   `json:"serverId"`
	GameID          string   `json:"gameId"`
	ServerName      string   `json:"serverName"`
	ProtocolVersion int      `json:"protocolVersion"`
	MCPEVersion     string   `json:"mcpeVersion"`
	PlayerCount     int      `json:"playerCount"`
	MaxPlayers      int      `json:"maxPlayers"`
	Extra           []string `json:"extra"`
}

var offlineMessageDataID = []byte{
	0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe,
	0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78,
}

// WriteUnconnectedPingPacket writes the 'Unconnected Ping (0x01)' as a single packet to a connection.
func WriteUnconnectedPingPacket(conn net.Conn, timestamp uint64) error {
	buf := new(bytes.Buffer)

	if err := WriteUnconnectedPing(buf, timestamp); err != nil {
		return err
	}

	_, err := conn.Write(buf.Bytes())
	return err
}

// WriteUnconnectedPing writes the 'Unconnected Ping (0x01)' packet to a writer.
// Details on the packet structure can be found:
// https://github.com/NiclasOlofsson/MiNET/blob/5bcfbfd94cff943f31208eb8614b3ff16269fdc7/src/MiNET/MiNET/Net/MCPE%20Protocol.cs#L1003
func WriteUnconnectedPing(writer io.Writer, timestamp uint64) error {
	if err := binary.Write(writer, binary.BigEndian, byte(0x01)); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, timestamp); err != nil {
		return err
	}
	if _, err := writer.Write(offlineMessageDataID); err != nil {
		return err
	}
	return nil
}

// ReadUTFString reads a UTF-8 string with a uint16 length header.
func ReadUTFString(reader io.Reader) (string, error) {
	var strLen uint16
	if err := binary.Read(reader, binary.BigEndian, &strLen); err != nil {
		return "", err
	}

	strBytes := make([]byte, strLen)
	if _, err := reader.Read(strBytes); err != nil {
		return "", err
	}

	return string(strBytes), nil
}

// ReadUnconnectedPong reads the 'Unconnected Pong (0x1C)' packet from a connection into a Response struct.
// Details on the packet structure can be found:
// https://github.com/NiclasOlofsson/MiNET/blob/5bcfbfd94cff943f31208eb8614b3ff16269fdc7/src/MiNET/MiNET/Net/MCPE%20Protocol.cs#L1154
func ReadUnconnectedPong(reader *bufio.Reader, resp *Response) error {
	id, err := reader.ReadByte()
	if err != nil {
		return err
	}
	if id != 0x1c {
		return fmt.Errorf("unexpected packet id: %d", id)
	}

	if err = binary.Read(reader, binary.BigEndian, &resp.Timestamp); err != nil {
		return err
	}
	if err = binary.Read(reader, binary.BigEndian, &resp.ServerID); err != nil {
		return err
	}

	temp := make([]byte, 16)
	if _, err = reader.Read(temp); err != nil {
		return err
	}
	if !bytes.Equal(offlineMessageDataID, temp) {
		return fmt.Errorf("invalid offline message data id: %x", temp)
	}

	payload, err := ReadUTFString(reader)
	if err != nil {
		return err
	}

	split := strings.Split(payload, ";")
	if len(split) < 6 {
		return fmt.Errorf("invalid payload: %s", payload)
	}
	if len(split) > 6 {
		resp.Extra = split[6:]
	}

	resp.GameID = split[0]
	resp.ServerName = split[1]

	resp.ProtocolVersion, err = strconv.Atoi(split[2])
	if err != nil {
		return err
	}

	resp.MCPEVersion = split[3]

	resp.PlayerCount, err = strconv.Atoi(split[4])
	if err != nil {
		return err
	}

	resp.MaxPlayers, err = strconv.Atoi(split[5])
	if err != nil {
		return err
	}

	return nil
}

// Query makes a query to the specified address via the Minecraft Bedrock protocol,
// if successful it returns a Response containing data from the pong packet.
// resend is the interval that the ping packet is sent in case there is packet loss.
func Query(address string, timeout time.Duration, resend time.Duration) (Response, error) {
	var resp Response

	deadline := time.Now().Add(timeout)

	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		return resp, err
	}
	defer conn.Close()

	if err = conn.SetDeadline(deadline); err != nil {
		return resp, err
	}

	ctx, cancel := context.WithDeadline(context.TODO(), deadline)
	defer cancel()

	var errs chan error

	// Repeat sending ping packet in case there is packet loss
	ticker := time.NewTicker(resend)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := WriteUnconnectedPingPacket(conn, 0); err != nil {
					errs <- err
					return
				}
			}
		}
	}()

	reader := bufio.NewReader(conn)
	if err = ReadUnconnectedPong(reader, &resp); err != nil {
		return resp, err
	}

	select {
	case err := <-errs:
		return resp, err
	default:
		return resp, nil
	}
}
