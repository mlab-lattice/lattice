package template

import (
	"log"
	"os"
	"text/template"
)

// These here for reference

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

var DefaultTemplate = `
{{define "UsageTemplateHeader"}}Usage template for command {{.Name}}{{end}}
{{define "HelpTemplate"}}{{.CommandPath}}{{if (ne .Short "") }} - {{.Short}}{{end}}
{{if (gt (len .Flags) 0)}}
Flags: {{range .Flags}}
    --{{ rpad .GetName $.FlagNamePadding }} {{if (ne .GetShort "") }} -{{ rpad .GetShort 2 }} {{ .GetUsage }} {{else}} {{rpad " " 4}}{{ .GetUsage }}{{end}}{{end}}
{{end}}
{{if .HasSubcommands}}Available Commands:{{range .AllSubcommands}}
    {{rpad .CommandPath .NamePadding }} {{.Short}}{{end}}

{{end}}{{end}}
{{define "HelpTemplateGrouped"}}{{.CommandPath}}{{if (ne .Short "") }} - {{.Short}}{{end}}
{{if (gt (len .Flags) 0)}}
{{colored "Flags:" "white"}} {{range .Flags}}
    --{{ rpad .GetName $.FlagNamePadding }} {{if (ne .GetShort "") }} -{{ rpad .GetShort 2 }} {{ .GetUsage }} {{else}} {{rpad " " 4}}{{ .GetUsage }}{{end}}{{end}}
{{end}}
{{if .HasSubcommands}}Available Commands:

{{range .SubcommandsByGroup}}{{ colored .GroupName "blue" }}: {{range .Commands}}
    {{ colored (rpad .Name .NamePadding) "none" }} {{colored .Short "gray"}}{{end}}

{{end}}{{end}}{{end}}
{{define "UsageTemplate"}}{{template "HelpTemplate" .}}{{end}}
{{define "UsageTemplateGrouped"}}{{template "HelpTemplateGrouped" .}}{{end}}`

// TryExecuteTemplate provides a simple wrapper to try and execute a template with some common options, and write the result to Stdout.
func TryExecuteTemplate(templateContents string, templateToCreate string, subtemplateToExecute string, templateFunctions template.FuncMap, c interface{}) error {
	tmpl, err := template.New(templateToCreate).Funcs(templateFunctions).Parse(templateContents)
	if err != nil {
		log.Fatalf("error creating %v template: %v \n", templateToCreate, err)
		return err
	}

	if subtemplateToExecute != "" {
		// Execute a named template within the definition
		err = tmpl.ExecuteTemplate(os.Stdout, subtemplateToExecute, c)
	} else {
		// Execute the whole template
		err = tmpl.Execute(os.Stdout, c)
	}
	if err != nil {
		log.Fatalf("error executing %v: %v \n", templateToCreate, err)
	}
	return err
}
