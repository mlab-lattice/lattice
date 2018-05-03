package cli

import (
	"log"

	"github.com/mlab-lattice/lattice/pkg/util/cli/template"
)

// UsageFuncGroupedCommands provides a usage function that groups commands by the subtree that they are located in
func UsageFuncGroupedCommands(command *Command) error {
	tmplName := "groupedUsageTemplate"
	templateToExecute := "UsageTemplateGrouped"
	return template.TryExecuteTemplate(template.DefaultTemplate, tmplName, templateToExecute, template.DefaultTemplateFuncs, command)
}

// HelpFuncGroupedCommands is a help function that groups commands by the subtree that they are located in
func HelpFuncGroupedCommands(command *Command) {
	tmplName := "groupedHelpTemplate"
	templateToExecute := "HelpTemplateGrouped"
	err := template.TryExecuteTemplate(template.DefaultTemplate, tmplName, templateToExecute, template.DefaultTemplateFuncs, command)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
