package gotftp

import (
	"bytes"
	"testing"
)

func TestTFTP_buildErrorPacket(t *testing.T) {
	cases := []struct {
		name     string
		errCode  uint16
		errMsg   string
		expected []byte
	}{
		{"file not found", 1, "f", []byte{0, 5, 0, 1, 'f', 0x00}},
		{"invalid opcode", 4, "i", []byte{0, 5, 0, 4, 'i', 0x00}},
		{"undefined errcode", 9, "e", []byte{0, 5, 0, 0, 'i', 0x00}},
	}

	for _, tc := range cases {
		packet := buildErrorPacket(tc.errCode, tc.errMsg)
		t.Log(tc.errMsg)

		if !bytes.Equal(packet, tc.expected) {
			t.Errorf("[test: %s] got %v expecting: %v\n",
				tc.name,
				packet,
				tc.expected,
			)
		}
	}
}
