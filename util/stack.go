package util

import (
	"errors"
	"fmt"
)

type Stack []interface{}

func (stack *Stack) Pop() (interface{}, error) {
	theStack := *stack
	if len(theStack) == 0 {
		return nil, errors.New("stack: can't Pop() an empty stack")
	}
	x := theStack[len(theStack)-1]
	*stack = theStack[:len(theStack)-1]
	return x, nil
}

func (stack *Stack) Push(x interface{}) {
	*stack = append(*stack, x)
}

func (stack Stack) Top() (interface{}, error) {
	if len(stack) == 0 {
		return nil, errors.New("stack: can't Top() an empty stack")
	}
	return stack[len(stack)-1], nil
}

func (stack Stack) Cap() int {
	return cap(stack)
}

func (stack Stack) Len() int {
	return len(stack)
}

func (stack Stack) IsEmpty() bool {
	return len(stack) == 0
}

func (stack Stack) Dump() (out string) {
	for i, it := range stack {
		out = out + fmt.Sprintf("%d: %#v\n", i, it)
	}
	return
}
