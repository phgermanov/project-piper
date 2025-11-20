package main

//go:generate go run pkg/generator/step-metadata.go --metadataDir=./resources/metadata/ --targetDir=./cmd/

import (
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/cmd"
)

func main() {
	cmd.Execute()
}
