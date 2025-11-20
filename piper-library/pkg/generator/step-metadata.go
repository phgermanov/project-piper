package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	piperOsGenerator "github.com/SAP/jenkins-library/pkg/generator/helper"
)

func main() {
	var metadataPath string
	var targetDir string

	flag.StringVar(&metadataPath, "metadataDir", "./resources/metadata", "The directory containing the step metadata. Default points to \\'resources/metadata\\'.")
	flag.StringVar(&targetDir, "targetDir", "./cmd", "The target directory for the generated commands.")
	flag.Parse()

	metadataFiles, err := piperOsGenerator.MetadataFiles(metadataPath)
	checkError(err)

	err = piperOsGenerator.ProcessMetaFiles(metadataFiles, targetDir, piperOsGenerator.StepHelperData{openMetaFile, fileWriter, "piperOsCmd"})
	checkError(err)

	fmt.Printf("Running go fmt %v\n", targetDir)
	cmd := exec.Command("go", "fmt", targetDir)
	err = cmd.Run()
	checkError(err)
}
func openMetaFile(name string) (io.ReadCloser, error) {
	return os.Open(filepath.Clean(name))
}

func fileWriter(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err)
		os.Exit(1)
	}
}
