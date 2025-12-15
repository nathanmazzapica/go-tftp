package mock

import (
	"errors"
	"io"
	"net"
	"time"
)

type MockPacketConn struct {
	ReadBuffer  []byte // Bytes the mock receives
	WriteBuffer []byte // Bytes the mock "writes"
	ReadCount   int
}

var ErrNotImplemented = errors.New("not implemented")

func (m *MockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	if m.ReadCount >= len(m.ReadBuffer) {
		return 0, nil, io.EOF
	}

	data := m.ReadBuffer[m.ReadCount:]
	copy(p, data)
	n = len(data)

	m.ReadCount++

	addr = &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9069}
	return 0, addr, ErrNotImplemented
}

func (m *MockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	m.WriteBuffer = append(m.WriteBuffer, p...)
	return len(p), nil
}

func (m *MockPacketConn) Close() error {
	return ErrNotImplemented
}

func (m *MockPacketConn) LocalAddr() net.Addr {
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9069}
	return addr
}

func (m *MockPacketConn) SetDeadline(t time.Time) error {
	return ErrNotImplemented
}

func SetReadDeadline(t time.Time) error {
	return ErrNotImplemented
}

func SetWriteDeadline(t time.Time) error {
	return ErrNotImplemented
}
