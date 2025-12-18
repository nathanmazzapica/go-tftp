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
		{"undefined errcode", 9, "e", []byte{0, 5, 0, 0, 'e', 0x00}},
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

func TestTFTP_buildReadRequestPacket(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		mode     string
		expected []byte
	}{
		{
			name:     "octet mode single char filename",
			filename: "f",
			mode:     "octet",
			expected: []byte{
				0x00, 0x01, // RRQ opcode
				'f',
				0x00,
				'o', 'c', 't', 'e', 't',
				0x00,
			},
		},
		{
			name:     "netascii mode",
			filename: "test.txt",
			mode:     "netascii",
			expected: []byte{
				0x00, 0x01,
				't', 'e', 's', 't', '.', 't', 'x', 't',
				0x00,
				'n', 'e', 't', 'a', 's', 'c', 'i', 'i',
				0x00,
			},
		},
	}

	for _, tc := range cases {
		packet := buildReadRequestPacket(tc.filename, tc.mode)
		t.Log(tc.name)

		if !bytes.Equal(packet, tc.expected) {
			t.Errorf(
				"[test: %s] got %v expecting: %v\n",
				tc.name,
				packet,
				tc.expected,
			)
		}
	}
}

func TestTFTP_parseReadRequest(t *testing.T) {
	cases := []struct {
		name     string
		packet   []byte
		filename string
		mode     string
		wantErr  bool
	}{
		{
			name: "valid rrq octet",
			packet: []byte{
				0x00, 0x01,
				'f',
				0x00,
				'o', 'c', 't', 'e', 't',
				0x00,
			},
			filename: "f",
			mode:     "octet",
			wantErr:  false,
		},
		{
			name: "valid rrq netascii",
			packet: []byte{
				0x00, 0x01,
				't', 'e', 's', 't', '.', 't', 'x', 't',
				0x00,
				'n', 'e', 't', 'a', 's', 'c', 'i', 'i',
				0x00,
			},
			filename: "test.txt",
			mode:     "netascii",
			wantErr:  false,
		},
		{
			name: "invalid opcode",
			packet: []byte{
				0x00, 0x02, // WRQ instead of RRQ
				'f',
				0x00,
				'o', 'c', 't', 'e', 't',
				0x00,
			},
			wantErr: true,
		},
		{
			name: "missing null terminator",
			packet: []byte{
				0x00, 0x01,
				'f',
				// missing 0x00 null terminator
				'o', 'c', 't', 'e', 't',
				0x00,
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		filename, mode, err := parseReadRequest(tc.packet)
		t.Log(tc.name)

		if tc.wantErr {
			if err == nil {
				t.Errorf("[test: %s] expected error, got nil", tc.name)
			}
			continue
		}

		if err != nil {
			t.Errorf("[test: %s] unexpected error: %v", tc.name, err)
			continue
		}

		if filename != tc.filename || mode != tc.mode {
			t.Errorf(
				"[test: %s] got (filename=%q, mode=%q) expecting (filename=%q, mode=%q)",
				tc.name,
				filename,
				mode,
				tc.filename,
				tc.mode,
			)
		}
	}
}
