package gotftp

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"path"
	"time"

	"github.com/z46-dev/go-logger"
)

type TFTPOptions struct {
	RootDir string
	Addr    string
}

type TFTPServer struct {
	rootDir    string
	log        *logger.Logger
	addr       *net.UDPAddr
	conn       *net.UDPConn
	readBuffer []byte
}

func NewTFTPServer(options *TFTPOptions) (*TFTPServer, error) {
	if options.RootDir == "" {
		return nil, fmt.Errorf("no RootDir provided in TFTPOptions")
	}

	if options.Addr == "" {
		return nil, fmt.Errorf("no Address provided in TFTPOptions")
	}

	addr, err := net.ResolveUDPAddr("udp4", options.Addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	server := &TFTPServer{
		rootDir:    options.RootDir,
		log:        logger.NewLogger().SetPrefix("[TFTP]", logger.BoldPurple).IncludeTimestamp(),
		addr:       addr,
		conn:       conn,
		readBuffer: make([]byte, 2048),
	}

	return server, nil
}

func (s *TFTPServer) ListenAndServe(ctx context.Context) error {
	for {
		s.log.Basicf("Listening...\n")
		select {
		case <-ctx.Done():
			s.log.Status("Server stopped due to quit signal")
			return nil
		default:

			err := s.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			if err != nil {
				s.log.Errorf("failed to set read deadline on UDP conn: %s\n", err.Error())
			}

			bytesRead, clientAddr, err := s.conn.ReadFromUDP(s.readBuffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// hit read deadline, continue
					continue
				}
				s.log.Error(err.Error())
				continue
			}

			// reset deadline
			s.conn.SetReadDeadline(time.Time{})

			// Smallest valid packet possible is 4 bytes (ACK: [opcode][block#])
			if bytesRead < 4 {
				s.log.Errorf("dropping invalid packet: %v from %s\n", s.readBuffer, clientAddr.String())
				continue
			}

			// TODO: Move to getOPCODE?
			// Why?: Testing, clarity
			opcode := binary.BigEndian.Uint16(s.readBuffer[:2])

			switch opcode {
			case OPCODE_RRQ:
				s.log.Basicf("processing RRQ\n")
				// parse request
				filename, mode, err := parseReadRequest(s.readBuffer)
				if err != nil {
					s.log.Errorf("error parsing read request packet: %v", err)
					continue
				}

				s.log.Basicf("transfering %s to %v\n", filename, clientAddr)

				filepath := path.Join(s.rootDir, filename)

				timestampStart := time.Now()

				tempAddr, err := net.ResolveUDPAddr("udp4", ":0")
				tempConn, err := net.ListenUDP("udp4", tempAddr)

				err = transferFile(filepath, mode, tempConn, clientAddr)
				if err != nil {
					s.log.Errorf("%v\n", err)
					err := sendError(err, 0, tempConn, clientAddr)
					if err != nil {
						s.log.Errorf("failed to send error packet. error: %v\n", err)
					}
				}
				duration := time.Now().Sub(timestampStart)
				s.log.Basicf("Transfer complete in %v\n", duration)
			case OPCODE_WRQ:
				s.log.Basicf("processing WRQ\n")
			case OPCODE_DATA:
				s.log.Basicf("processing DATA OP\n")
			case OPCODE_ACK:
				s.log.Basicf("processing ACK\n")
			case OPCODE_ERROR:
				s.log.Basicf("processing ERROR\n")
			default:
				s.log.Warningf("received invalid op code: %d", opcode)
				// send ERROR 4 Illegal TFTP operation
				continue
			}
		}
	}

}
