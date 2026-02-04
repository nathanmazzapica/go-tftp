package netascii

import (
	"bufio"
	"io"
	"log"
)

type Reader struct {
	r *bufio.Reader
}

const (
	_CR  = '\r'
	_LF  = '\n'
	_NUL = 0
)

func NewReader(rd io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(rd)}
}

// Read file byte by byte
// if byte = 'CR' && byte+1 = 'LF' -> do nothing
// else
// if byte = 'CR' -> 'CRNUL'
// if byte = 'LF' -> 'CRLF'

// Read reads byte by byte from a filestream, converting OS specific line-endings into netascii compliant sequences
func (nr *Reader) Read(p []byte) (int, error) {
	n := 0

	for ; n < len(p); n++ {
		b, err := nr.r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return n, io.EOF
			}
			return n, err
		}
		switch b {
		case _CR:
			p[n] = _CR
			next, err := nr.r.Peek(1)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println(err)
			}
			if next[0] != _LF {
				n++
				if n == len(p) {
					return n, nil
				}
				p[n] = _NUL
			}
		case _LF:
			p[n] = _CR
			n++
			p[n] = _LF
		default:
			p[n] = b
		}
	}

	return n, nil
}
