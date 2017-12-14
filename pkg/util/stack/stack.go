//stack based on slice

// credit https://github.com/junpengxiao/Stack
package stack

import "errors"

type Stack struct {
	data []interface{}
}

var ErrEmptyStack = errors.New("stack.go : stack is empty")

func NewStack(number uint) *Stack {
	return &Stack{data: make([]interface{}, 0, number)}
}

//return the number of items in stack
func (s *Stack) Len() int {
	return len(s.data)
}

func (s *Stack) Push(value interface{}) {
	s.data = append(s.data, value)
}

//pop the top item out, if stack is empty, will return ErrEmptyStack decleared above
func (s *Stack) Pop() (interface{}, error) {
	if s.Len() > 0 {
		rect := s.data[s.Len()-1]
		s.data = s.data[:s.Len()-1]
		return rect, nil
	}
	return nil, ErrEmptyStack
}

//peek the top item. Notice, this is like a pointer:
//tmp, _ := s.Peek(); tmp = 123;
//s.Pop() should return 123, nil.
func (s *Stack) Peek() (interface{}, error) {
	if s.Len() > 0 {
		return s.data[s.Len()-1], nil
	}
	return nil, ErrEmptyStack
}
