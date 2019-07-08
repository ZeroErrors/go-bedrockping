package bedrockping

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestWriteUnconnectedPing(t *testing.T) {
	buf := new(bytes.Buffer)

	if err := WriteUnconnectedPing(buf, 0); err != nil {
		t.Error(err)
	}

	if buf.Len() != 25 {
		t.Error("didn't write 25 bytes")
	}

	var id byte
	if err := binary.Read(buf, binary.BigEndian, &id); err != nil {
		t.Error(err)
	}
	if id != 0x01 {
		t.Errorf("invalid packet id: %d", id)
	}

	var timestamp uint64
	if err := binary.Read(buf, binary.BigEndian, &timestamp); err != nil {
		t.Error(err)
	}
	if timestamp != 0 {
		t.Errorf("invalid timestamp: %d", timestamp)
	}

	temp := make([]byte, 16)
	if _, err := buf.Read(temp); err != nil {
		t.Error(err)
	}
	if !bytes.Equal(offlineMessageDataID, temp) {
		t.Errorf("invalid offline message data id: %x", temp)
	}

	if buf.Len() != 0 {
		t.Error("failed to read all bytes")
	}
}

func writeUTFString(buf io.Writer, str string) error {
	var strLen = uint16(len(str))
	if err := binary.Write(buf, binary.BigEndian, &strLen); err != nil {
		return err
	}

	if _, err := buf.Write([]byte(str)); err != nil {
		return err
	}

	return nil
}

func TestReadUTFString(t *testing.T) {
	var testString = "This is some test string"

	buf := new(bytes.Buffer)

	if err := writeUTFString(buf, testString); err != nil {
		t.Error(err)
	}

	str, err := ReadUTFString(buf)
	if err != nil {
		t.Error(err)
	}

	if str != testString {
		t.Errorf("failed to read string: '%s'", str)
	}
}

func TestReadUnconnectedPong(t *testing.T) {
	expect := Response{
		Timestamp:       0,
		ServerID:        0,
		GameID:          "MCPE",
		ServerName:      "ServerName",
		ProtocolVersion: 0,
		MCPEVersion:     "0.0.0",
		PlayerCount:     0,
		MaxPlayers:      0,
		Extra:           []string{"Extra", "Stuff"},
	}

	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, byte(0x1c)); err != nil {
		t.Error(err)
	}
	if err := binary.Write(buf, binary.BigEndian, expect.Timestamp); err != nil {
		t.Error(err)
	}
	if err := binary.Write(buf, binary.BigEndian, expect.ServerID); err != nil {
		t.Error(err)
	}
	if _, err := buf.Write(offlineMessageDataID); err != nil {
		t.Error(err)
	}

	payload := fmt.Sprintf("%s;%s;%d;%s;%d;%d",
		expect.GameID,
		expect.ServerName,
		expect.ProtocolVersion,
		expect.MCPEVersion,
		expect.PlayerCount,
		expect.MaxPlayers)
	if expect.Extra != nil {
		payload = payload + ";" + strings.Join(expect.Extra, ";")
	}
	if err := writeUTFString(buf, payload); err != nil {
		t.Error(err)
	}

	var resp Response

	reader := bufio.NewReader(buf)
	if err := ReadUnconnectedPong(reader, &resp); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expect, resp) {
		t.Errorf("incorrect resp: %v", resp)
	}
}

func TestQuery(t *testing.T) {
	_, err := Query("hivebedrock.network:19132", 5*time.Second, 150*time.Millisecond)
	if err != nil {
		t.Error(err)
	}
}
