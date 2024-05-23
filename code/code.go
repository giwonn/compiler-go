package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// 명령코드
const (
	OpConstant Opcode = iota
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		// opcode 유효성 검증
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		// read = 읽은 인덱스 개수
		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def,
			operands))

		// 읽은 만큼 더해주고 +1을 해서 읽어야 할 시작지점 갱신
		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandByteWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

type Opcode byte

type Definition struct {
	Name              string
	OperandByteWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}}, // OpContstant는 2byte(=16bit) 크기의 단일 피연산자를 가질 수 있음
}

// Lookup 존재하는 opcode인지 체크
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[OpConstant]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// Make 바이트코드 피연산자 부호화
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1 // 첫 바이트는 opcode 고정이기에 1로 초기화
	for _, width := range def.OperandByteWidths {
		instructionLen += width
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op) // 첫 번째 인덱스는 opcode 할당

	offset := 1
	for i, operand := range operands {
		width := def.OperandByteWidths[i]
		switch width {
		case 2:
			// operand를 16bit(=2byte) 단위로 빅엔디안으로 쪼개서 offset부터 저장
			binary.BigEndian.PutUint16(instruction[offset:], uint16(operand))
		}
		offset += width
	}

	return instruction
}

// ReadOperands 부호화 된 피연산자 복호화
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandByteWidths))
	offset := 0
	for i, width := range def.OperandByteWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

// 입력값으로 들어온 배열의 두 바이트만 리턴함
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
