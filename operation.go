package gotftp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"
)

// parseReadRequest takes a raw request packet and extracts 'filename' and 'mode'.
//
// Returns an error in the following scenarios:
//
// 1. missing null terminator for filename or mode
//
// 2. filename or mode exceed maximum lengths (1024 and 8 respectively)
func parseReadRequest(req []byte, rootDir string) (filename, mode string, err error) {
	// this means no string & missing NT
	if len(req) < 4 {
		return "", "", fmt.Errorf("rrq too short")
	}

	// [opcode] [filename] [nt] [mode] [nt]
	opcode := binary.BigEndian.Uint16(req[0:2])

	if opcode != uint16(1) {
		return "", "", fmt.Errorf("invalid opcode")
	}

	// MacOS max filepath length is 1024. Linux/ext4 is 4096.
	// Windows is 260 without Win32 Long Paths enabled
	filenameCap := 1024
	filenameBytes := make([]byte, 0, filenameCap)

	var p int
	for p = 2; ; p++ {
		if p >= len(req) {
			return "", "", fmt.Errorf("malformed rrq: missing filename null terminator")
		}

		if req[p] == 0x00 {
			break
		}

		filenameBytes = append(filenameBytes, req[p])
		if len(filenameBytes) > filenameCap {
			return "", "", fmt.Errorf("filename length exceeded limits: %d", filenameCap)
		}
	}

	// advance to next non-null-terminator byte
	p++

	// The longest mode (netascii) is 8 characters, so we set the cap of modeBytes to 8
	modeCap := 8
	modeBytes := make([]byte, 0, modeCap)
	for ; ; p++ {
		if p >= len(req) {
			return "", "", fmt.Errorf("malformed rrq: missing mode null terminator")
		}

		if req[p] == 0x00 {
			break
		}

		modeBytes = append(modeBytes, req[p])
		if len(modeBytes) > modeCap {
			return "", "", fmt.Errorf("mode length exceeded limits: %d", modeCap)
		}
	}

	filename = string(filenameBytes)
	filename = path.Join(rootDir, filename)
	mode = string(modeBytes)

	return filename, mode, nil
}

func processRRQ(rootDir string, req []byte, conn *net.UDPConn, clientAddr *net.UDPAddr) error {
	filepath, mode, err := parseReadRequest(req, rootDir)
	if err != nil {
		return err
	}
	_ = mode

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}

	var blockNum uint16
	blockNum = 1
	for {
		data, err := readBlock(f, blockNum)

		err = transmitDataPacket(blockNum, data, conn, clientAddr)
		if err != nil {
			return err
		}

		if len(data) < BLOCK_SIZE {
			return nil
		}
	}
}

// TODO: reflect is ReadAt is even necessary. The cursor is already at the calculated offset.
func readBlock(f *os.File, blockNum uint16) ([]byte, error) {
	buffer := make([]byte, BLOCK_SIZE)
	offset := int64((blockNum - 1) * BLOCK_SIZE)
	bytesRead, err := f.ReadAt(buffer, offset)
	return buffer[:bytesRead], err
}

func transmitDataPacket(blockNum uint16, data []byte, conn *net.UDPConn, addr *net.UDPAddr) error {
	// === TRANSMIT BUFFER === //
	var retryCount int
	ack := make([]byte, BLOCK_SIZE)
	packet := buildDataPacket(blockNum, data)

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// send the data
		err := sendPacket(packet, conn, addr)
		if err != nil {
			log.Println(err)
			return err
		}

		n, _, err := conn.ReadFromUDP(ack)

		if err == nil {
			ackBlockNum, err := parseAckPacket(ack[:n])
			if err != nil {
				// client sent something unexpected
				return fmt.Errorf("unexpected client response")
			}

			if ackBlockNum != blockNum {
				// todo: handle this
				fmt.Printf("Expecting #%d\nGot: #%d\n", blockNum, ackBlockNum)
				return fmt.Errorf("blocknum mismatch handling not implemented")
			}

			break
		}

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			retryCount++
			if retryCount >= 3 {
				return fmt.Errorf("timed out waiting for ack after 3 retries")
			}
			fmt.Println("timeout... resending packet")
			continue
		}
		return err
	}
	return nil
}

func sendError(err error, errCode uint16, conn *net.UDPConn, addr *net.UDPAddr) error {
	packet := buildErrorPacket(errCode, err.Error())
	return sendPacket(packet, conn, addr)
}

func sendPacket(packet []byte, conn *net.UDPConn, addr *net.UDPAddr) error {
	_, err := conn.WriteTo(packet, addr)
	return err
}
