package cli

import (
"fmt"
"reflect"
"strconv"
"strings"
"text/template"
"unicode"
)

// Adds the default command template as well as some default functions to use for this template

var CobraUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

var CobraHelpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

var DefaultUsageTemplate = `Usage template for command {{.Name}}
Cant access helpTempl from here
`

var DefaultHelpTemplate = `Called {{.CommandPath}}:
{{if (ne .Short "") }} {{.Short}}
{{end}}
Flags {{range .Flags}}
    --{{ rpad .GetName 10 }} {{if (ne .GetShort "") }} -{{ rpad .GetShort 2 }} {{end}} {{ .GetUsage }}{{end}}

General Commands:{{range .AllSubcommands}}
  {{rpad .CommandPath 10 }} {{.Short}}{{end}}

`

// These are pretty much taken straight from cobra. Used to get a working template with similar behavior.
var templateFuncs = template.FuncMap{
    "trim":                    strings.TrimSpace,
    "trimRightSpace":          trimRightSpace,
    "trimTrailingWhitespaces": trimRightSpace,
    //"appendIfNotPresent":      appendIfNotPresent,
    "rpad": rpad,
    "gt":   Gt,
    "eq":   Eq,
}

func trimRightSpace(s string) string {
    return strings.TrimRightFunc(s, unicode.IsSpace)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
    template := fmt.Sprintf("%%-%ds", padding)
    return fmt.Sprintf(template, s)
}

// Gt takes two types and checks whether the first type is greater than the second. In case of types Arrays, Chans,
// Maps and Slices, Gt will compare their lengths. Ints are compared directly while strings are first parsed as
// ints and then compared.
func Gt(a interface{}, b interface{}) bool {
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

// Eq takes two types and checks whether they are equal. Supported types are int and string. Unsupported types will panic.
func Eq(a interface{}, b interface{}) bool {
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
