package bedrock_ping

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
	// Default port
	DefaultPort = 19132
)

var offlineMessageDataId = []byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}

type Response struct {
	Timestamp		uint64		`json:"timestamp"`
	ServerId		uint64		`json:"serverId"`
	GameId			string		`json:"gameId"`
	ServerName		string		`json:"serverName"`
	ProtocolVersion	int			`json:"protocolVersion"`
	MCPEVersion		string		`json:"mcpeVersion"`
	PlayerCount		int			`json:"playerCount"`
	MaxPlayers		int			`json:"maxPlayers"`
	Extra			[]string	`json:"extra"`
}

// https://github.com/NiclasOlofsson/MiNET/blob/5bcfbfd94cff943f31208eb8614b3ff16269fdc7/src/MiNET/MiNET/Net/MCPE%20Protocol.cs#L1003
func WriteUnconnectedPing(conn net.Conn, timestamp uint64) error {
	buf := new(bytes.Buffer)
	if err := buf.WriteByte(0x01); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, timestamp); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, offlineMessageDataId); err != nil {
		return err
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

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

// https://github.com/NiclasOlofsson/MiNET/blob/5bcfbfd94cff943f31208eb8614b3ff16269fdc7/src/MiNET/MiNET/Net/MCPE%20Protocol.cs#L1154
func ReadUnconnectedPong(conn net.Conn, response *Response) error {
	reader := bufio.NewReader(conn)

	id, err := reader.ReadByte()
	if err != nil {
		return err
	}
	if id != 0x1c {
		return fmt.Errorf("unexpected packet id %d", id)
	}

	if err = binary.Read(reader, binary.BigEndian, &response.Timestamp); err != nil {
		return err
	}
	if err = binary.Read(reader, binary.BigEndian, &response.ServerId); err != nil {
		return err
	}

	magicValue := make([]byte, 16)
	_, err = reader.Read(magicValue)
	if err != nil {
		return err
	}
	if !bytes.Equal(offlineMessageDataId, magicValue) {
		return fmt.Errorf("invalid offlineMessageDataId %x", magicValue)
	}

	payload, err := ReadUTFString(reader)
	if err != nil {
		return err
	}

	split := strings.Split(payload, ";")
	if len(split) < 6 {
		return fmt.Errorf("invalid payload %s", payload)
	}
	if len(split) > 6 {
		response.Extra = split[6:]
	}

	response.GameId		= split[0]
	response.ServerName	= split[1]

	response.ProtocolVersion, err = strconv.Atoi(split[2])
	if err != nil {
		return err
	}

	response.MCPEVersion = split[3]

	response.PlayerCount, err = strconv.Atoi(split[4])
	if err != nil {
		return err
	}

	response.MaxPlayers, err = strconv.Atoi(split[5])
	if err != nil {
		return err
	}

	return nil
}

func Query(address string, timeout time.Duration) (Response, error) {
	var response Response

	deadline := time.Now().Add(timeout)

	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		return response, err
	}
	defer conn.Close()

	if err = conn.SetDeadline(deadline); err != nil {
		return response, err
	}

	if err = WriteUnconnectedPing(conn, 0); err != nil {
		return response, err
	}

	if err = ReadUnconnectedPong(conn, &response); err != nil {
		return response, err
	}

	return response, nil
}
