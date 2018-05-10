package template

import (
	"log"
	"os"
	"text/template"
)

//FIXME :: Seem to get : rather than " " when running command systems -h, but not when running command -h or command systems:status -h
//FIXME :: Add a CommandPath variable that isn't prefixed with the program name beforehand. Subcommands should not say `lattice systems ...` but rather `systems`
var DefaultTemplate = `{{define "Header"}}{{ colored "Usage: " "white" }}{{.CommandPath}}{{.CommandSeparator}}{{if not .IsRunnable }}{{colored "COMMAND" "bold"}}{{end}}{{if .HasFlags}}{{ colored "[FLAGS] " "bold"}}{{else}}{{colored "COMMAND" "bold"}}{{end}}
{{if not (eq .Short "") }}
    {{colored .Short "bold"}}{{end}}

Type {{.CommandPath}}{{.CommandSeparator}}{{if .HasSubcommands}}{{ colored "[COMMAND] " "bold"}}{{end}}{{colored "-h" "bold"}} for help and examples.{{end}}

{{define "HelpTemplate"}}{{template "Header" .}}
{{if (gt (len .Flags) 0)}}
{{colored "Flags:" "white"}} {{range .FlagsSorted}}
    --{{ rpad .GetName $.FlagNamePadding }} {{if (ne .GetShort "") }} -{{ rpad .GetShort 2 }} {{ .GetUsage }} {{else}} {{rpad " " 4}}{{ .GetUsage }}{{end}}{{end}}
{{end}}
{{if .HasSubcommands}}{{ colored "Subcommands: " "white" }}{{range .AllSubcommands}}
    {{ colored (rpad .CommandPath .NamePadding) "blue" }} {{ colored .Short "gray" }}{{end}}
{{end}}{{end}}

{{define "HelpTemplateGrouped"}}{{template "Header" .}}
{{if (gt (len .Flags) 0)}}
{{colored "Flags:" "white"}} {{range .Flags}}
    --{{ rpad .GetName $.FlagNamePadding }} {{if (ne .GetShort "") }} -{{ rpad .GetShort 2 }} {{ .GetUsage }} {{else}} {{rpad " " 4}}{{ .GetUsage }}{{end}}{{end}}
{{end}}
{{if .HasSubcommands}}{{ colored "Subcommands: " "white" }}
{{range .SubcommandsByGroup}}
{{ colored .GroupName "blue" }}: {{range .Commands}}
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
