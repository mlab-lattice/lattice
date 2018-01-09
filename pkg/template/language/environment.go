// TemplateEngine environment
package language

import (
	"errors"
)

// environment. Template Parsing environment
type environment struct {
	options *Options
	engine  *TemplateEngine
	stack   *environmentStack
}

// newenvironment creates a new environment object
func newEnvironment(engine *TemplateEngine, options *Options) *environment {
	env := &environment{
		engine:  engine,
		stack:   newStack(10),
		options: options,
	}

	return env
}

// currentDir returns the current directory of the file being parsed
func (env *environment) parametersAndVariables() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range env.currentFrame().parameters {
		result[k] = v
	}

	for k, v := range env.currentFrame().variables {
		result[k] = v
	}

	return result
}

func (env *environment) push(
	baseUrl string,
	parameters map[string]interface{},
	variables map[string]interface{}) {
	env.stack.Push(&environmentStackFrame{
		baseUrl:    baseUrl,
		parameters: parameters,
		variables:  variables,
	})
}

func (env *environment) pop() error {
	_, err := env.stack.Pop()
	if err != nil {
		return err
	}

	return nil
}

func (env *environment) currentFrame() *environmentStackFrame {
	currentFrame, _ := env.stack.Peek()
	return currentFrame

}

// environment stack
type environmentStack struct {
	data []*environmentStackFrame
}

// environment stack frame
type environmentStackFrame struct {
	baseUrl    string
	parameters map[string]interface{}
	variables  map[string]interface{}
}

// ErrEmptyStack raised when the stack is empty on pop or peek
var ErrEmptyStack = errors.New("stack.go : stack is empty")

func newStack(number uint) *environmentStack {
	return &environmentStack{data: make([]*environmentStackFrame, 0, number)}
}

// length return the number of items in stack
func (s *environmentStack) length() int {
	return len(s.data)
}

//Push pushes a frame into stack
func (s *environmentStack) Push(value *environmentStackFrame) {
	s.data = append(s.data, value)
}

//pop the top item out, if stack is empty, will return ErrEmptyStack decleared above
func (s *environmentStack) Pop() (*environmentStackFrame, error) {
	if s.length() > 0 {
		rect := s.data[s.length()-1]
		s.data = s.data[:s.length()-1]
		return rect, nil
	}
	return nil, ErrEmptyStack
}

//peek the top item. Notice, this is like a pointer:
//tmp, _ := s.Peek(); tmp = 123;
//s.Pop() should return 123, nil.
func (s *environmentStack) Peek() (*environmentStackFrame, error) {
	if s.length() > 0 {
		return s.data[s.length()-1], nil
	}
	return nil, ErrEmptyStack
}
