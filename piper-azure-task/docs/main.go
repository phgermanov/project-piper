package main

//go:generate go run main.go

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/goccy/go-json"

	"github.tools.sap/project-piper/piper-azure-task/pkg/layout"
	"github.tools.sap/project-piper/piper-azure-task/pkg/model"
)

var (
	taskFilePath           = flag.String("task", "../piper/task.json", "Path to the task config file")
	readmeTemplateFilePath = flag.String("readme-template", "resources/README.md", "Path to the README.md template")
)

func main() {
	flag.Parse()
	// Read task from task file
	taskBuf, err := ioutil.ReadFile(*taskFilePath)
	if err != nil || len(taskBuf) == 0 {
		log.Fatalln(fmt.Errorf("can't read task: %w", err))
	}
	// Unmarshal task
	var task model.Task
	if err = json.Unmarshal(taskBuf, &task); err != nil {
		log.Fatalln(fmt.Errorf("can't umarshal task: %w", err))
	}
	// Read README template
	document, err := ioutil.ReadFile(*readmeTemplateFilePath)
	if err != nil || len(document) == 0 {
		log.Fatalln(fmt.Errorf("can't read README.md template: %w", err))
	}
	// Fill placeholders
	document = bytes.Replace(document, []byte("[[YML_SNIPPET]]"), layout.GetYMLSnippet(task), -1)
	document = bytes.Replace(document, []byte("[[ARGUMENTS_TABLE]]"), layout.GetArgumentsTable(task), -1)
	// Write document
	if err := ioutil.WriteFile("../README.md", document, 0644); err != nil {
		log.Fatalln(fmt.Errorf("can't write readme: %w", err))
	}
}
