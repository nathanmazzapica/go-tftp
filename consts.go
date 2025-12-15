package gotftp

const (
	BLOCK_SIZE = 512
	// Read Request
	OPCODE_RRQ = iota
	// Write Request
	OPCODE_WRQ
	// Data
	OPCODE_DATA
	// Acknowledgement
	OPCODE_ACK
	// Error
	OPCODE_ERROR
)
