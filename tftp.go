package gotftp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/z46-dev/go-logger"
)

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
	TFTPAddress string
	HTTPOptions *HTTPOptions
}

// Serve creates a TFTP Server and tells it to listen
// TODO: Divide into NewTFTPServer & (receiver) Listen()
func Serve(options TFTPOptions) (quit chan bool, err error) {
	if options.RootDir == "" {
		return nil, fmt.Errorf("config err: options.RootDir is required")
	}

	if options.TFTPAddress == "" {
		return nil, fmt.Errorf("config err: options.TFTP_Address is required")
	}

	var (
		addr *net.UDPAddr
		conn *net.UDPConn
	)

	if addr, err = net.ResolveUDPAddr("udp4", options.TFTPAddress); err != nil {
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
			buffer   []byte = make([]byte, 1024)
			filename string
			opcode   int
			log      *logger.Logger = logger.NewLogger().SetPrefix("[TFTP]", logger.BoldPurple).IncludeTimestamp()
		)

		log.Status("Server started")

		for {
			fmt.Printf("listening\n")
			select {
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

				if bytesRead < 4 {
					log.Errorf("Invalid request from %s\n", clientAddr.String())
					continue
				}

				opcode = int(buffer[1])

				if opcode == OPCODE_RRQ {
					if filename, _, err = ParseRQQRequest(buffer[:bytesRead]); err != nil {
						log.Errorf("Failed to parse RRQ request for %s: %s\n", clientAddr.String(), err.Error())
						continue
					}

					var fPath string = path.Join(options.RootDir, filename)
					log.Warningf("Received RRQ request for %s from %s\n", fPath, clientAddr.String())

					if !strings.HasPrefix(fPath, options.RootDir) {
						SendError(conn, addr, 1, "File not found")
					} else if err = SendFile(conn, clientAddr, fPath); err != nil {
						log.Errorf("Failed to send file: %s\n", err.Error())
					} else {
						log.Successf("File %s sent to %s\n", filename, clientAddr.String())
					}
				}
			}
		}

	}()

	return quit, nil
}
