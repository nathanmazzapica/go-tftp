package netascii

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
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

func Test_ValidCRNULSequence(t *testing.T) {
	stream := []byte("hey\r\000")
	buf := make([]byte, len(stream))

	r := bytes.NewReader(stream)
	nar := NewReader(r)

	_, err := nar.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, stream, buf)
}

func Test_NullBuffer(t *testing.T) {
	stream := []byte("hey\r\000")

	r := bytes.NewReader(stream)
	nar := NewReader(r)

	n, err := nar.Read(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

func Test_ReadOverflow(t *testing.T) {
	stream := []byte("hey\rbro")
	buf := make([]byte, 4)
	expected := []byte{'h', 'e', 'y', '\r'}
	expected2 := []byte{'\000', 'b', 'r', 'o'}

	r := bytes.NewReader(stream)
	nar := NewReader(r)

	_, err := nar.Read(buf)

	assert.NoError(t, err)
	assert.Equal(t, expected, buf)

	_, err = nar.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected2, buf)

}

func Test_BufferLargerThanStream(t *testing.T) {
	stream := []byte("hey")
	buf := make([]byte, 4)
	expected := []byte("hey\000")

	r := bytes.NewReader(stream)
	nar := NewReader(r)

	n, err := nar.Read(buf)

	if err == io.EOF {
		// rationale: Go can return EOF and valid data simultaneously
		err = nil
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, buf)
	// important because in implementation buf[:n] is used to write to udp
	assert.Equal(t, len(stream), n)
}
