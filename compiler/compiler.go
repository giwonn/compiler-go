package compiler

import (
	"compiler-go/ast"
	"compiler-go/code"
	"compiler-go/object"
	"fmt"
)

type Compiler struct {
	instructions code.Instructions // 생성한 바이트코드
	constants    []object.Object   // 상수 풀

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction

	symbolTable *SymbolTable
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

// 가상머신에 전달할 대상
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		symbolTable:         NewSymbolTable(),
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.BlockStatement:
		for _, statement := range node.Statements {
			err := c.Compile(statement)
			if err != nil {
				return err
			}
		}

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// OpJumpNotTruthy 명령어에 쓰레기값 9999를 넣어서 배출
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		// if문이 true일 때 if문 내부 표현식이 실행되고 나서 Alternative를 건너뛰게 해줌
		jumpPos := c.emit(code.OpJump, 9999)

		// if문이 false일 때 jumpNotTruthyPos가 jump문 다음으로 스킵하여 Alternative를 읽도록 해줌
		afterConsequencePos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil { // Alternative가 없으면 OpNull을 스택에 추가함
			c.emit(code.OpNull)
		} else { // Alternative가 있으면 해당 표현식문의 평가값을 추가함
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.instructions)
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.emit(code.OpGetGlobal, symbol.Index)
	}

	return nil
}

// addConstant 상수풀에 obj 추가 후 obj의 인덱스 값 리턴
func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// emit 피연산자 instruction 생성 후 추가. 추가한 instruction 시작 위치 리턴
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	instruction := code.Make(op, operands...)
	pos := c.addInstruction(instruction)

	c.setLastInstruction(op, pos)

	return pos
}

// addInstruction 명령어 추가 후 추가한 명령어의 시작 위치를 리턴
func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}

// 이러면 c.previousInstruction이 c.lastInstruction과 같아지지 않나?
func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}
