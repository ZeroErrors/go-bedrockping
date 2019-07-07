// Package bedrockping is a simple library to ping Minecraft Bedrock/MCPE servers.
package bedrockping

import (
	"bufio"
	"bytes"
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

// WriteUnconnectedPing writes the 'Unconnected Ping (0x01)' packet to a connection.
// Details on the packet structure can be found:
// https://github.com/NiclasOlofsson/MiNET/blob/5bcfbfd94cff943f31208eb8614b3ff16269fdc7/src/MiNET/MiNET/Net/MCPE%20Protocol.cs#L1003
func WriteUnconnectedPing(conn net.Conn, timestamp uint64) error {
	buf := new(bytes.Buffer)
	if err := buf.WriteByte(0x01); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, timestamp); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, offlineMessageDataID); err != nil {
		return err
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
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
	_, err := reader.Read(strBytes)
	if err != nil {
		return "", err
	}

	return string(strBytes), nil
}

// ReadUnconnectedPong reads the 'Unconnected Pong (0x1C)' packet from a connection into a Response struct.
// Details on the packet structure can be found:
// https://github.com/NiclasOlofsson/MiNET/blob/5bcfbfd94cff943f31208eb8614b3ff16269fdc7/src/MiNET/MiNET/Net/MCPE%20Protocol.cs#L1154
func ReadUnconnectedPong(conn net.Conn, resp *Response) error {
	reader := bufio.NewReader(conn)

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

	magicValue := make([]byte, 16)
	_, err = reader.Read(magicValue)
	if err != nil {
		return err
	}
	if !bytes.Equal(offlineMessageDataID, magicValue) {
		return fmt.Errorf("invalid offline message data id: %x", magicValue)
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
func Query(address string, timeout time.Duration) (Response, error) {
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

	if err = WriteUnconnectedPing(conn, 0); err != nil {
		return resp, err
	}

	if err = ReadUnconnectedPong(conn, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
