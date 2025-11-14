package layout

import (
	"fmt"
	"strings"

	"github.tools.sap/project-piper/piper-azure-task/pkg/model"
)

func GetYMLSnippet(task model.Task) []byte {
	snippet := make([]byte, 0)
	snippet = append(snippet, []byte(
		"```yml\n"+
			fmt.Sprintf("# %s\n# %s\n", task.FriendlyName, task.Description)+
			fmt.Sprintf("- task: %s@%v\n", task.Name, task.Version.Major)+
			"  # or if you would like to pin specific version of the task, use '"+
			fmt.Sprintf("%s@1.2.3", task.Name)+"'\n"+
			"  inputs:\n",
	)...)
	for _, input := range task.Inputs {
		snippet = append(snippet, []byte(fmt.Sprintf("    %s\n", getInputDocumentation(input)))...)
	}
	return append(snippet, []byte("```")...)
}

func GetArgumentsTable(task model.Task) []byte {
	table := ""
	table += "| Argument | Description |\n| -------- | ----------- |\n"
	for _, input := range task.Inputs {
		table += fmt.Sprintf("%s\n", getArgumentAndDescription(input))
	}
	return []byte(strings.TrimSuffix(table, "\n"))
}

func getInputDocumentation(input model.Input) string {
	var defaultValueString string
	defaultValue := input.DefaultValue
	if defaultValue == nil {
		defaultValue = ""
	}
	switch defaultValue.(type) {
	case string:
		defaultValueString = fmt.Sprintf("'%s'", defaultValue)
	default:
		defaultValueString = fmt.Sprintf("%v", defaultValue)
	}
	if !input.Required {
		defaultValueString += " # Optional"
	}
	return fmt.Sprintf("%s: %s", input.Name, defaultValueString)
}

func getArgumentAndDescription(input model.Input) string {
	var defaultValueString string
	defaultValue := input.DefaultValue
	switch input.DefaultValue.(type) {
	case string:
		if len(defaultValue.(string)) != 0 {
			defaultValueString = fmt.Sprintf("%s", input.DefaultValue)
		}
	default:
		if defaultValue != nil {
			defaultValueString = fmt.Sprintf("%v", input.DefaultValue)
		}
	}
	if len(defaultValueString) != 0 {
		defaultValueString = fmt.Sprintf(" </br> Default value: `%s`", defaultValueString)
	}
	return fmt.Sprintf(
		"| `%s` </br> %s | (%s) %s%s |",
		input.Name, input.Label,
		func() string {
			if input.Required {
				return "Required"
			}
			return "Optional"
		}(),
		input.HelpMarkDown, defaultValueString,
	)
}
