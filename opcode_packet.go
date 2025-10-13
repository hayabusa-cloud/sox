// ©Hayabusa Cloud Co., Ltd. 2024. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sox

const (
	// OpcodeTextMessage denotes a text message
	OpcodeTextMessage = 1 << 4
	// OpcodeBinaryMessage denotes a binary message
	OpcodeBinaryMessage = 2 << 4
	// OpcodeBuiltinCommand denotes a builtin-command message with 4-bits command code
	OpcodeBuiltinCommand = 4 << 4
	// OpcodeExtBuiltinCommand denotes a builtin-command message with 1-byte command code
	OpcodeExtBuiltinCommand = 5 << 4
	// OpcodeClose denotes a connection close
	OpcodeClose = 8 << 4
	// OpcodePing denotes a ping
	OpcodePing = 9 << 4
	// OpcodePong denotes a pong
	OpcodePong = 10 << 4
)

// OpcodePacketReader is an interface that extends the PacketReader interface,
// adding a method ReadOpcodeCommand, which reads an opcode command from the packet.
type OpcodePacketReader interface {
	PacketReader

	// ReadOpcodeCommand reads an opcode command from the packet.
	// It returns the opcode and cmd integers, and any error that occurred.
	ReadOpcodeCommand() (opcode int, cmd int, err error)
}

// OpcodePacketWriter is an interface that extends the PacketWriter interface,
// adding a method for writing opcode commands.
type OpcodePacketWriter interface {
	PacketWriter

	// WriteOpcodeCommand writes an opcode command to the PacketWriter.
	// The opcode specifies the type of command, and the cmd parameter
	// provides additional information about the command.
	// The returned error indicates any error that occurred during writing.
	//
	// Parameters:
	// - opcode: The opcode of the command to be written.
	// - cmd: Additional information about the command.
	//
	// Returns:
	// - error: An error if any occurred during writing.
	WriteOpcodeCommand(opcode int, cmd int) error
}

type opcodePacketReadWriter struct {
	PacketReader
	PacketWriter
}

// ReadOpcodeCommand reads an opcode and a command from the underlying reader. It returns
// the opcode and command as separate integers. If there is an error reading from
// the underlying reader, it returns zero values for opcode, command, and the error.
func (rw *opcodePacketReadWriter) ReadOpcodeCommand() (opcode int, cmd int, err error) {
	val, err := rw.ReadByte()
	if err != nil {
		return 0, 0, nil
	}
	return int(val & 0xf0), int(val & 0x0f), nil
}

// WriteOpcodeCommand writes an opcode and a command to the underlying writer.
// It takes the opcode and command as a byte type integer.
// It returns an error if there is an error writing to the underlying writer.
func (rw *opcodePacketReadWriter) WriteOpcodeCommand(opcode int, cmd int) error {
	return rw.WriteByte(byte(opcode) | byte(cmd))
}
