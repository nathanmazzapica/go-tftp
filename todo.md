# TODO:
- [ ] Parse incoming packet
    - [x] Get opcode
    - [ ] Get options / args
    - [ ] Route to appropriate handler
- [ ] Handle read operations 
- [ ] Handle write operations
- [ ] Handle data transfer
    - [x] Build data packet
    - [ ] Write data from file to buffer
- [ ] Handle acks
    - [x] Build ack packet
- [ ] Handle errors
    - [x] Build error packet

# Notes

## Operations

### OPCODE 01/02 READ/WRITE

#### Packet Format
           2 bytes    string   1 byte     string   1 byte
          -----------------------------------------------
   RRQ/  | 01/02 |  Filename  |   0  |    Mode    |   0  |
   WRQ    -----------------------------------------------

`0` represents a null terminator

### OPCODE 03 DATA

#### Packet Format
          2 bytes    2 bytes       n bytes
          ---------------------------------
   DATA  | 03    |   Block #  |    Data    |
          ---------------------------------

### OPCODE 04 ACK

#### Packet Format
          2 bytes    2 bytes
          -------------------
   ACK   | 04    |   Block #  |
          --------------------

### OPCODE 05 ERROR

#### Packet Format

          2 bytes  2 bytes        string    1 byte
          ----------------------------------------
   ERROR | 05    |  ErrorCode |   ErrMsg   |   0  |
          ----------------------------------------
          `Length = 5 + len(ErrMsg)`

#### Error Codes
| Value | Meaning                                   |
|------:|:------------------------------------------|
|     0 | Not defined, see error message (if any).  |
|     1 | File not found.                           |
|     2 | Access violation.                         |
|     3 | Disk full or allocation exceeded.         |
|     4 | Illegal TFTP operation.                   |
|     5 | Unknown transfer ID.                      |
|     6 | File already exists.                      |
|     7 | No such user.                             |



## Scratchpad

Go's UDP.ReadFrom([]byte) only copies the TFTP payload. The rest is handled by kernel
                                                  2 bytes
    ----------------------------------------------------------
   |  Local Medium  |  Internet  |  Datagram  |  TFTP Opcode  |
    ----------------------------------------------------------

Likewise, a client source code would only write the TFTP payload, the kernel would wrap it in the network details. Neat.



data packet of < 512 bytes = transfer over

[RFC 1350](https://datatracker.ietf.org/doc/html/rfc1350#autoid-2)
zx;lcxkxzckjzkcjklcjzxklcjxzkcjxcjzxlkcjzxklcjzxkcjzxcjzxlcjzxlkcjzxlkcjzxlcjzxkcjzxlkcjzxklcjzxkcjzxlcjzxkl

```go
// pseudocode-ish

server, err := gotftp.NewTFTPServer(&options{...})
if err != nil {
    panic(err)
}

// server.ListenAndServer() is a blocking call, much like net/http.ListenAndServe()
err = server.ListenAndServe()
```



  
