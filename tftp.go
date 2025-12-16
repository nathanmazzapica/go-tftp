package gotftp

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/z46-dev/go-logger"
)

// TODO: Remove this
func serveHTTP(rootDir, httpAddr string) *http.Server {
	var mux *http.ServeMux = http.NewServeMux()
	var log *logger.Logger = logger.NewLogger().SetPrefix("[HTTP]", logger.BoldGreen).IncludeTimestamp()

	log.Statusf("Serving HTTP on %s\n", httpAddr)

	fs := http.FileServer(http.Dir(rootDir))

	mux.Handle("/", fs)

	var server *http.Server = &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error(err.Error())
		}

		log.Status("Server stopped")
	}()

	return server
}

type HTTPOptions struct {
	RootDir string
	Address string
}

type TFTPOptions struct {
	RootDir     string
	Addr        string
	HTTPOptions *HTTPOptions
}

type TFTPServer struct {
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
		log:        logger.NewLogger().SetPrefix("[TFTP]", logger.BoldPurple).IncludeTimestamp(),
		conn:       conn,
		readBuffer: make([]byte, 2048),
	}

	return server, nil
}

func (s *TFTPServer) ListenAndServe(ctx context.Context) error {
	for {
		fmt.Printf("listening\n")
		select {
		// TODO: Use context instead
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
				s.log.Basicf("processing RRQ")
			case OPCODE_WRQ:
				s.log.Basicf("processing WRQ")
			case OPCODE_DATA:
				s.log.Basicf("processing DATA OP")
			case OPCODE_ACK:
				s.log.Basicf("processing ACK")
			case OPCODE_ERROR:
				s.log.Basicf("processing ERROR")
			default:
				s.log.Warningf("received invalid op code: %d", opcode)
				// send ERROR 4 Illegal TFTP operation
				continue
			}
		}
	}

}

// Serve creates a TFTP Server and tells it to listen
// TODO: Divide into NewTFTPServer & (receiver) Listen()
func Serve(options TFTPOptions) (quit chan bool, err error) {
	if options.RootDir == "" {
		return nil, fmt.Errorf("config err: options.RootDir is required")
	}

	if options.Addr == "" {
		return nil, fmt.Errorf("config err: options.TFTP_Address is required")
	}

	var (
		addr *net.UDPAddr
		conn *net.UDPConn
	)

	if addr, err = net.ResolveUDPAddr("udp4", options.Addr); err != nil {
		return nil, err
	}

	if conn, err = net.ListenUDP("udp4", addr); err != nil {
		return nil, err
	}

	var server *http.Server = nil
	if opts := options.HTTPOptions; opts != nil {
		if opts.RootDir == "" {
			return nil, fmt.Errorf("config err: must provide http root dir")
		}

		if opts.Address == "" {
			return nil, fmt.Errorf("config err: must provide http address")
		}

		server = serveHTTP(opts.RootDir, opts.Address)
	} else {
		fmt.Println("serving TFTP without HTTP")
	}

	quit = make(chan bool)

	go func() {
		defer conn.Close()

		if server != nil {
			defer server.Shutdown(context.TODO())
		}

		var (
			buffer []byte         = make([]byte, 1024)
			log    *logger.Logger = logger.NewLogger().SetPrefix("[TFTP]", logger.BoldPurple).IncludeTimestamp()
		)

		log.Status("Server started")

		for {
			fmt.Printf("listening\n")
			select {
			// TODO: Use context instead
			case <-quit:
				log.Status("Server stopped due to quit signal")
				return
			default:

				err := conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				if err != nil {
					log.Errorf("failed to set read deadline on UDP conn: %s\n", err.Error())
				}

				bytesRead, clientAddr, err := conn.ReadFromUDP(buffer)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// hit read deadline, continue
						continue
					}
					log.Error(err.Error())
					continue
				}

				// reset deadline
				conn.SetReadDeadline(time.Time{})

				// Smallest valid packet possible is 4 bytes (ACK: [opcode][block#])
				if bytesRead < 4 {
					log.Errorf("dropping invalid packet: %v from %s\n", buffer, clientAddr.String())
					continue
				}

				// TODO: Move to getOPCODE?
				// Why?: Testing, clarity
				opcode := binary.BigEndian.Uint16(buffer[:2])

				switch opcode {
				case OPCODE_RRQ:
					log.Basicf("processing RRQ")
				case OPCODE_WRQ:
					log.Basicf("processing WRQ")
				case OPCODE_DATA:
					log.Basicf("processing DATA OP")
				case OPCODE_ACK:
					log.Basicf("processing ACK")
				case OPCODE_ERROR:
					log.Basicf("processing ERROR")
				default:
					log.Warningf("received invalid op code: %d", opcode)
					// send ERROR 4 Illegal TFTP operation
					continue
				}
			}
		}

	}()

	return quit, nil
}
