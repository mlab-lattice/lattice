package language

import (
	"errors"
	"fmt"
	"strings"
)

// environment. Template evaluation environment
type environment struct {
	options *Options
	engine  *TemplateEngine
	stack   *environmentStack

	propertyMetadataMap map[string]*PropertyMetadata // mapping of property paths and metadata
	propertyStack       *environmentStack

	// track reference recipients. Map of reference to array of recipients
	referenceRecipients map[string][]string
}

// newEnvironment creates a new environment object
func newEnvironment(engine *TemplateEngine, options *Options) *environment {
	env := &environment{
		engine:              engine,
		stack:               newStack(10),
		options:             options,
		propertyMetadataMap: make(map[string]*PropertyMetadata),
		propertyStack:       newStack(10),
		referenceRecipients: make(map[string][]string),
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
	template *Template,
	parameters map[string]interface{},
	variables map[string]interface{}) {
	env.stack.push(&environmentStackFrame{
		template:   template,
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
	if currentFrame != nil {
		return currentFrame.(*environmentStackFrame)
	}

	return nil
}

// pushProperty
func (env *environment) pushProperty(property string) string {
	env.propertyStack.push(property)

	propertyPath := env.getCurrentPropertyPath()
	env.fillPropertyMetadata(propertyPath)

	return propertyPath
}

// popProperty
func (env *environment) popProperty() (string, error) {
	prop, err := env.propertyStack.pop()
	return prop.(string), err
}

// fillPropertyMetadata
func (env *environment) fillPropertyMetadata(propertyPath string) {
	var currentTemplate *Template

	if env.currentFrame() != nil {
		currentTemplate = env.currentFrame().template
	}

	metadata := &PropertyMetadata{
		propertyPath: propertyPath,
		template:     currentTemplate,
	}

	env.propertyMetadataMap[propertyPath] = metadata

	// compute relative property path after creating/registering it
	metadata.relativePropertyPath = env.computeRelativePropertyPathFor(metadata)
}

// computeRelativePropertyPathFor determines the relative path of the property, i.e. relative to the template
// that contains that property
func (env *environment) computeRelativePropertyPathFor(metadata *PropertyMetadata) string {
	parentPropertyPath := getParentPropertyPath(metadata.propertyPath)
	relativePropertyPath := metadata.PropertyName()
	for parentPropertyPath != "" {
		parentMeta := env.getPropertyMetaData(parentPropertyPath)

		// if the parent is in a different template then
		if parentMeta.template == nil || parentMeta.template != metadata.template {
			break
		}
		relativePropertyPath = fmt.Sprintf("%v.%v", parentMeta.PropertyName(), relativePropertyPath)
		parentPropertyPath = getParentPropertyPath(parentPropertyPath)
	}

	return relativePropertyPath
}

// relativePathToAbsolute
func (env *environment) relativePathToAbsolute(relativePath string) string {
	currentFrame := env.currentFrame()

	if currentFrame == nil || currentFrame.template == nil {
		return relativePath
	}

	rootPath := env.getCurrentPropertyPath()

	for rootPath != "" {
		propertyMeta := env.getPropertyMetaData(rootPath)

		// climb up until we reach the root property path for the current template
		if propertyMeta == nil || propertyMeta.template == nil || propertyMeta.template != currentFrame.template {
			break
		}
		rootPath = getParentPropertyPath(rootPath)
	}

	if rootPath != "" {
		return fmt.Sprintf("%v.%v", rootPath, relativePath)
	}
	return relativePath
}

// getPropertyMetaData
func (env *environment) getPropertyMetaData(propertyPath string) *PropertyMetadata {

	return env.propertyMetadataMap[propertyPath]
}

// getCurrentPropertyPath
func (env *environment) getCurrentPropertyPath() string {
	propertyPath := make([]string, len(env.propertyStack.data))
	for i, property := range env.propertyStack.data {
		propertyPath[i] = property.(string)
	}
	return strings.Join(propertyPath, ".")
}

// captureReferenceRecipient
func (env *environment) captureReferenceRecipient(reference Reference) {
	recipient := env.getCurrentPropertyPath()

	// if we are part of an operator eval like $variables or $parameters or parameters of an $include then bail
	if env.propertyPathHasOperator(recipient) {
		return
	}

	allRecipients := env.referenceRecipients[reference.getTarget()]
	if allRecipients == nil {
		allRecipients = []string{recipient}
	} else {
		// append recipient to array
		allRecipients = append(allRecipients, recipient)
	}

	env.referenceRecipients[reference.getTarget()] = allRecipients

}

// propertyPathHasOperator
func (env *environment) propertyPathHasOperator(propertyPath string) bool {
	parts := strings.Split(propertyPath, ".")
	// ensure that this is a valid recipient
	for _, part := range parts {
		for operatorKey := range env.engine.operatorMap {
			if part == operatorKey {
				return true
			}
		}
	}
	return false
}

// environment stack
type environmentStack struct {
	data []interface{}
}

// environment stack frame
type environmentStackFrame struct {
	template   *Template
	parameters map[string]interface{}
	variables  map[string]interface{}
}

// ErrEmptyStack raised when the stack is empty on pop or peek
var ErrEmptyStack = errors.New("stack.go : stack is empty")

func newStack(number uint) *environmentStack {
	return &environmentStack{data: make([]interface{}, 0, number)}
}

// length return the number of items in stack
func (s *environmentStack) length() int {
	return len(s.data)
}

//Push pushes a frame into stack
func (s *environmentStack) push(value interface{}) {
	s.data = append(s.data, value)
}

//pop the top item out, if stack is empty, will return ErrEmptyStack decleared above
func (s *environmentStack) pop() (interface{}, error) {
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
func (s *environmentStack) peek() (interface{}, error) {
	if s.length() > 0 {
		return s.data[s.length()-1], nil
	}
	return nil, ErrEmptyStack
}
