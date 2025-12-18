package gotftp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

// parseReadRequest takes a raw request packet and extracts 'filename' and 'mode'.
//
// Returns an error in the following scenarios:
//
// 1. missing null terminator for filename or mode
//
// 2. filename or mode exceed maximum lengths (1024 and 8 respectively)
func parseReadRequest(req []byte) (filename, mode string, err error) {
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
	mode = string(modeBytes)

	return filename, mode, nil
}

func transferFile(filename string, conn *net.UDPConn, addr *net.Addr) {
	f, err := os.Open(filename)
	if err != nil {
		if errors.Is(os.ErrNotExist, err) {
			// write error to socket
		}
	}
	defer f.Close()

	blockNum := 1
	for {

		buffer := make([]byte, 512)
		bytesRead, err := f.Read(buffer)

		fmt.Printf("Read %d bytes from file\n", bytesRead)

		if err != nil {
			if errors.Is(io.EOF, err) {
				// we're done sport
			}
			fmt.Printf("err: %v", err)
			break
		}

		packet := buildDataPacket(0, buffer)
		buffer = nil

		err = sendPacket(packet, conn, addr)
		if err != nil {
			fmt.Printf("%v", err)
		}

		// wait for ack

		// increment blocknum
		if bytesRead < 512 {
			blockNum++
			continue
		}
		// data < 512 signals completion
		break
	}

	fmt.Println("Transfer complete")

}

func sendPacket(packet []byte, conn *net.UDPConn, addr *net.Addr) error {
	_, err := conn.Write(packet)
	return err
}
