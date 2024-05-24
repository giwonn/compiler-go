package vm

import (
	"compiler-go/code"
	"compiler-go/compiler"
	"compiler-go/object"
	"fmt"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // 언제나 다음 값을 가리킴. 따라서 스택 최상단은 stack[sp-1]
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		// **인출**
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			// **복호화**
			// 현재 ip는 명령코드이므로 피연산자는 ip+1부터 2byte까지만 존재함
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			// 피연산자의 끝으로 이동함. for문이 끝나면서 ip++되기 때문에 사실상 ip += 3
			ip += 2
			// **실행**
			// 상수풀에서 필요한 상수를 꺼내서 스택에 push
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd:
			// compiler.Compile()에서 left -> right -> add 순으로 instruction 저장
			right := vm.pop()
			left := vm.pop()
			leftValue := left.(*object.Integer).Value
			rightValue := right.(*object.Integer).Value

			vm.push(&object.Integer{Value: leftValue + rightValue})
		}
	}

	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}
