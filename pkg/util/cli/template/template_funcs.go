package template

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/fatih/color"
)

// DefaultTemplateFuncs provides the template functions used by the default help/usage parser. This is exported so that a new implementation can use these functions too.
var DefaultTemplateFuncs = template.FuncMap{
	"rpad":    rpad,
	"lpad":    lpad,
	"gt":      gt,
	"eq":      eq,
	"colored": colored,
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// rpad adds padding to the right of a string.
func colored(s string, colChoice string) string {
	col := color.New(color.FgBlue).SprintFunc()

	switch colChoice {
	case "bold":
		col = color.New(color.Bold).SprintFunc()
	case "blue":
		col = color.New(color.FgBlue).SprintFunc()
	case "hiblue":
		col = color.New(color.Faint + color.FgBlue).SprintFunc()
	case "gray":
		col = color.New(color.FgBlack).SprintFunc()
	case "red":
		col = color.New(color.FgRed).SprintFunc()
	case "cyan":
		col = color.New(color.FgCyan).SprintFunc()
	case "white":
		col = color.New(color.FgWhite).SprintFunc()
	case "none":
		return fmt.Sprint(s)
	default:
		return fmt.Sprint(s)
	}
	return fmt.Sprintf("%s", col(s))
}

// rpad adds whitespace padding to the right of a string.
func rpad(s string, padding int) string {
	paddedString := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(paddedString, s)
}

// lpad adds padding to the left of a string.
func lpad(s string, padding int) string {
	paddedString := fmt.Sprintf("%% %ds", padding)
	return fmt.Sprintf(paddedString, s)
}

// gt takes two types and checks whether the first type is greater than the second. In case of types Arrays, Chans,
// Maps and Slices, Gt will compare their lengths. Ints are compared directly while strings are first parsed as
// ints and then compared.
func gt(a interface{}, b interface{}) bool {
	var left, right int64
	av := reflect.ValueOf(a)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		left = int64(av.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		left = av.Int()
	case reflect.String:
		left, _ = strconv.ParseInt(av.String(), 10, 64)
	}

	bv := reflect.ValueOf(b)

	switch bv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		right = int64(bv.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		right = bv.Int()
	case reflect.String:
		right, _ = strconv.ParseInt(bv.String(), 10, 64)
	}

	return left > right
}

// eq takes two types and checks whether they are equal. Supported types are int and string. Unsupported types will panic.
func eq(a interface{}, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		panic("Eq called on unsupported type")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return av.Int() == bv.Int()
	case reflect.String:
		return av.String() == bv.String()
	}
	return false
}
