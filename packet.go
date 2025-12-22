package gotftp

import (
	"encoding/binary"
	"fmt"
)

// buildReadRequestPacket creates a TFTP ReadRequest packet as defined in [RFC 1350]
//
// [RFC 1350]: https://datatracker.ietf.org/doc/html/rfc1350#autoid-5
func buildReadRequestPacket(filename, mode string) (packet []byte) {
	packetLength := 4 + len(filename) + len(mode)
	packet = make([]byte, packetLength)

	binary.BigEndian.PutUint16(packet[0:2], OPCODE_RRQ)
	copy(packet[2:], []byte(filename))
	offset := 2 + len(filename)
	packet[offset] = 0x00
	copy(packet[offset+1:], []byte(mode))

	packet[packetLength-1] = 0x00

	return packet
}

// buildDataPacket creates a TFTP data packet as defined in [RFC 1350]
//
// [RFC 1350]: https://datatracker.ietf.org/doc/html/rfc1350#autoid-5
func buildDataPacket(blockNumber uint16, data []byte) (packet []byte) {
	packetLength := 4 + len(data)
	packet = make([]byte, packetLength)

	binary.BigEndian.PutUint16(packet[0:2], OPCODE_DATA)
	binary.BigEndian.PutUint16(packet[2:4], blockNumber)

	copy(packet[4:], data)

	return packet
}

// buildAckPacket creates a TFTP ack packet as defined in [RFC 1350]
//
// [RFC 1350]: https://datatracker.ietf.org/doc/html/rfc1350#autoid-5
func buildAckPacket(blockNumber uint16) (packet []byte) {
	packetLength := 4
	packet = make([]byte, packetLength)

	binary.BigEndian.PutUint16(packet[0:2], OPCODE_ACK)
	binary.BigEndian.PutUint16(packet[2:4], blockNumber)

	return packet
}

// parseAckPacket extracts the block number from the ack packet
func parseAckPacket(packet []byte) (blockNumber uint16, err error) {
	if len(packet) != 4 {
		return 0, fmt.Errorf("invalid ack packet: %v\n", packet)
	}

	blockNumber = binary.BigEndian.Uint16(packet[2:4])

	return
}

// buildErrorPacket creates a TFTP error packet as defined in [RFC 1350]
//
// [RFC 1350]: https://datatracker.ietf.org/doc/html/rfc1350#autoid-5
func buildErrorPacket(errCode uint16, errMsg string) (packet []byte) {
	packetLength := 4 + len(errMsg) + 1 // opcode: 2 + ErrorCode: 2 + ... + terminator: 1
	packet = make([]byte, packetLength)

	// RFC 1350 Defined error codes: 1-7; 0 = Undefined Error
	if errCode > 7 {
		errCode = 0
	}

	binary.BigEndian.PutUint16(packet[0:2], OPCODE_ERROR)
	binary.BigEndian.PutUint16(packet[2:4], errCode)

	copy(packet[4:], errMsg)
	packet[packetLength-1] = 0x00

	return packet
}
