// TemplateEngine environment
package language

import (
	"errors"
)

// environment. Template evaluation environment
type environment struct {
	options *Options
	engine  *TemplateEngine
	stack   *environmentStack
}

// newEnvironment creates a new environment object
func newEnvironment(engine *TemplateEngine, options *Options) *environment {
	env := &environment{
		engine:  engine,
		stack:   newStack(10),
		options: options,
	}

	return env
}

// parametersAndVariables returns a union of parameters and variables
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

// push pushes current environment to the stack. Should be called in $include
func (env *environment) push(
	baseUrl string,
	parameters map[string]interface{},
	variables map[string]interface{}) {
	env.stack.push(&environmentStackFrame{
		baseUrl:    baseUrl,
		parameters: parameters,
		variables:  variables,
	})
}

// pop pops environment stack. Called after $include is done
func (env *environment) pop() error {
	_, err := env.stack.pop()
	if err != nil {
		return err
	}

	return nil
}

// currentFrame returns the current stack frame. nil if stack is empty
func (env *environment) currentFrame() *environmentStackFrame {
	currentFrame, _ := env.stack.peek()
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
func (s *environmentStack) push(value *environmentStackFrame) {
	s.data = append(s.data, value)
}

//pop the top item out, if stack is empty, will return ErrEmptyStack decleared above
func (s *environmentStack) pop() (*environmentStackFrame, error) {
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
func (s *environmentStack) peek() (*environmentStackFrame, error) {
	if s.length() > 0 {
		return s.data[s.length()-1], nil
	}
	return nil, ErrEmptyStack
}
