package gotftp

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func SendError(conn *net.UDPConn, addr *net.UDPAddr, packet []byte) (err error) {
	_, err = conn.WriteToUDP(packet, addr)
	return err
}

func ParseRQQRequest(buffer []byte) (file string, mode string, err error) {
	var (
		start int      = 2
		parts []string = make([]string, 0)
	)

	for i := 2; i < len(buffer); i++ {
		if buffer[i] == 0 {
			parts = append(parts, string(buffer[start:i]))
			start = i + 1
		}
	}

	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid request")
	}

	file = parts[0]
	mode = parts[1]
	return file, mode, nil
}

func SendFile(conn *net.UDPConn, addr *net.UDPAddr, filename string) (err error) {
	var file *os.File

	if file, err = os.Open(filename); err != nil {
		errPacket := buildErrorPacket(6, "File not found")
		SendError(conn, addr, errPacket)
		return err
	}
	defer file.Close()

	var (
		bytesRead int    = 0
		blockNum  uint16 = 1
		buffer    []byte = make([]byte, BLOCK_SIZE)
	)

	for {
		if bytesRead, err = file.Read(buffer); err != nil {
			return err
		}

		// TODO: Build Data Packet

		// TODO: Send Data Packet

		// -- handle ack -- //

		///////[ ACK PACKET ]/////////
		// [ 2 bytes ] [ 2 bytes ] //
		// [ OP CODE ] [ BLOCK # ] //
		////////////////////////////
		var ack []byte = make([]byte, 4)
		if _, _, err = conn.ReadFromUDP(ack); err != nil {
			return err
		}

		opcode := binary.BigEndian.Uint16(ack[:2])

		if opcode != OPCODE_ACK || ack[2] != byte(blockNum>>8) || ack[3] != byte(blockNum) {
			return fmt.Errorf("invalid ACK received: %v", ack)
		}

		blockNum++
		if bytesRead < BLOCK_SIZE {
			return nil
		}
	}
}
