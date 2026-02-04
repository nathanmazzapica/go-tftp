package netascii

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ReadConvertsLFtoCRLF(t *testing.T) {
	stream := []byte("hey\nbro\n")
	buf := make([]byte, 5)
	expected := []byte("hey\r\n")

	reader := bytes.NewReader(stream)
	netasciiReader := NewReader(reader)

	_, err := netasciiReader.Read(buf)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected, buf)

	expected2 := []byte("bro\r\n")

	_, err = netasciiReader.Read(buf)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected2, buf)
}

func Test_ReadConvertsCRtoCRNUL(t *testing.T) {
	stream := []byte("hey\rbro")
	buf := make([]byte, 5)
	expected := []byte{'h', 'e', 'y', '\r', 0}

	r := bytes.NewReader(stream)
	nar := NewReader(r)

	_, err := nar.Read(buf)

	assert.NoError(t, err)
	assert.Equal(t, expected, buf)
}

func Test_ReadMismatchedLength(t *testing.T) {
	stream := []byte("hey\rbro")
	buf := make([]byte, 4)
	expected := []byte{'h', 'e', 'y', '\r'}
	expected2 := []byte{'\n', 'b', 'r', 'o'}

	r := bytes.NewReader(stream)
	nar := NewReader(r)

	_, err := nar.Read(buf)

	assert.NoError(t, err)
	assert.Equal(t, expected, buf)

	_, err = nar.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected2, buf)

}
