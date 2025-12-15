package gotftp

const (
	BLOCK_SIZE = 512
	// [2 bytes] [2 bytes] [0-512 bytes]
	// [OP CODE] [BLOCK #] [...Data....]
	MAX_PACKET_SIZE = 516
	// Read Request
	OPCODE_RRQ uint16 = iota
	// Write Request
	OPCODE_WRQ
	// Data
	OPCODE_DATA
	// Acknowledgement
	OPCODE_ACK
	// Error
	OPCODE_ERROR
)
