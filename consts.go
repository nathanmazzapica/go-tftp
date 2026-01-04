package gotftp

const (
	BLOCK_SIZE = 512
	// [2 bytes] [2 bytes] [0-512 bytes]
	// [OP CODE] [BLOCK #] [...Data....]
	MAX_PACKET_SIZE = 4 + BLOCK_SIZE
	// Read Request
	OPCODE_RRQ uint16 = 1
	// Write Request
	OPCODE_WRQ uint16 = 2
	// Data
	OPCODE_DATA uint16 = 3
	// Acknowledgement
	OPCODE_ACK uint16 = 4
	// Error
	OPCODE_ERROR uint16 = 5

	MODE_OCTET    = "octet"
	MODE_NETASCII = "netascii"
)
