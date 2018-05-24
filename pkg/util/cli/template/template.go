package template

import (
	"os"
	"text/template"
)

const DefaultTemplateName = "defaultTemplate"
const DefaultHelpTemplate = "HelpTemplate"
const DefaultUsageTemplate = "UsageTemplate"
const DefaultUsageTemplateGrouped = "UsageTemplateGrouped"
const DefaultHelpTemplateGrouped = "HelpTemplateGrouped"

//FIXME :: Seem to get : rather than " " when running command systems -h, but not when running command -h or command systems:status -h
var DefaultTemplate = `
{{define "Header"}}{{ colored "Usage: " "white" }}{{.CommandPathBinary}}{{.CommandSeparator}}{{if not .IsRunnable }}{{colored "COMMAND" "bold"}}{{end}}{{if .HasFlags}}{{ colored "[FLAGS] " "bold"}}{{else}}{{colored "COMMAND" "bold"}}{{end}}{{if not (eq .Short "") }}

	{{colored .Short "bold"}}{{end}}{{end}}

{{define "HeaderHelp"}}{{template "Header" .}}{{end}}

{{define "HeaderUsage"}}{{template "Header" .}}

Type {{.CommandPathBinary}}{{.CommandSeparator}}{{if .HasSubcommands}}{{ colored "[COMMAND] " "bold"}}{{end}}{{colored "-h" "bold"}} for help and examples.{{end}}

{{define "Flags"}}{{if (gt (len .Flags) 0)}}{{ $namePadding := 10 }}

{{colored "Flags:" "white"}} {{range .FlagsSorted}}
    --{{ rpad .GetName $namePadding }} {{if (ne .GetShort "") }} -{{ rpad .GetShort 2 }} {{ .GetUsage }} {{else}} {{rpad " " 4}}{{ .GetUsage }}{{end}}{{end}}{{end}}{{end}}

{{define "TemplateBody"}}{{template "Flags" .}}{{ $namePadding := 35 }}
{{if .HasSubcommands}}
{{ colored "Subcommands: " "white" }}{{range .AllSubcommands}}
    {{ colored (rpad .CommandPathBinary $namePadding) "blue" }} {{ colored .Short "gray" }}{{end}}
{{end}}{{end}}

{{define "TemplateGroupedBody"}}{{template "Flags" .}}{{ $namePadding := 35 }}
{{if .HasSubcommands}}
{{ colored "Subcommands: " "white" }}
{{range .SubcommandsByGroup}}
{{ colored .GroupName "blue" }}: {{range .Commands}}
 {{ colored (rpad .Name $namePadding) "none" }} {{colored .Short "gray"}}{{end}}
{{end}}{{end}}{{end}}

{{define "HelpTemplate"}}{{template "HeaderHelp" .}}{{template "TemplateBody" .}}{{end}}
{{define "HelpTemplateGrouped"}}{{template "HeaderHelp" .}}{{template "TemplateGroupedBody" .}}{{end}}

{{define "UsageTemplate"}}{{template "HeaderUsage" .}}{{template "TemplateBody" .}}{{end}}
{{define "UsageTemplateGrouped"}}{{template "HeaderUsage" .}}{{template "TemplateGroupedBody" .}}{{end}}`

// TryExecuteTemplate provides a simple wrapper to try and execute a template with some common options, and write the result to Stdout.
func TryExecuteTemplate(name, contents, subtemplate string, funcs template.FuncMap, c interface{}) error {
	tmpl, err := template.New(name).Funcs(funcs).Parse(contents)
	if err != nil {
		return err
	}

	if subtemplate != "" {
		// Execute a named template within the definition
		err = tmpl.ExecuteTemplate(os.Stdout, subtemplate, c)
	} else {
		// Execute the whole template
		err = tmpl.Execute(os.Stdout, c)
	}
	return err
}
