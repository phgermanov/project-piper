package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

// File defines the method subset we use from os.File
type uFile interface {
	io.Writer
	io.StringWriter
	io.Closer
}

type sapURLScanUtils interface {
	piperutils.FileUtils
	FileOpen(name string, flag int, perm os.FileMode) (uFile, error)
}

type sapURLScanUtilsBundle struct {
	*piperutils.Files
}

func (u *sapURLScanUtilsBundle) FileOpen(name string, flag int, perm os.FileMode) (uFile, error) {
	return os.OpenFile(name, flag, perm)
}

func newSapURLScanUtils(config *sapURLScanOptions) *sapURLScanUtilsBundle {
	utils := sapURLScanUtilsBundle{
		Files: &piperutils.Files{},
	}
	return &utils
}

func sapURLScan(config sapURLScanOptions, telemetryData *telemetry.CustomData, influx *sapURLScanInflux) {
	telemetryData.ProxyLogFile = config.ProxyLogFile

	utils := newSapURLScanUtils(&config)

	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	influx.step_data.fields.url = "false"

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runSapURLScan(&config, utils, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	influx.step_data.fields.url = "true"
}

func runSapURLScan(config *sapURLScanOptions, utils sapURLScanUtils, workdir string) error {

	log.Entry().Infof("Scanning file \"%s\" for insecure URLs", config.ProxyLogFile)
	log.Entry().Infof("URLs matching the following patterns will be marked as insecure: %v", config.InsecureURLPatterns)
	log.Entry().Infof("Output report will be saved in %s", config.ReportLogFile)

	// Read proxy log file content
	proxyLogCont, err := utils.FileRead(config.ProxyLogFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to read proxy log file '%s'", config.ProxyLogFile)
	}

	// Create file to write output
	urlScanReport, err := utils.FileOpen(config.ReportLogFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to create report file '%s'", config.ReportLogFile)
	}

	reportHeader := fmt.Sprintf("URLScan Report\n"+
		"Scanning file \"%s\" for insecure URLs\n"+
		"URLs matching the following patterns will be marked as insecure: %v \n\n"+
		"The following insecure URLs found:\n",
		config.ProxyLogFile,
		config.InsecureURLPatterns)

	if _, err := urlScanReport.WriteString(reportHeader); err != nil {
		return errors.Wrapf(err, "Failed to write to report file '%s'", config.ReportLogFile)
	}

	var patterns []*regexp.Regexp
	for _, s := range config.InsecureURLPatterns {
		patterns = append(patterns, regexp.MustCompile(s))
	}

	// Create a scunner to read proxy log file line by line
	scanner := bufio.NewScanner(strings.NewReader(string(proxyLogCont[:])))

	countInsecureURLs := 0
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()

		// check if this line matches one of insecure url patterns
		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				countInsecureURLs += 1
				log.Entry().Warnf("Insecure URL found! Pattern \"%v\" found in line \"%v\". ", pattern, line)
				_, _ = urlScanReport.WriteString(line + "\n")
			}
		}
	}

	//archive report
	reports := []piperutils.Path{{Name: config.ReportLogFile, Target: config.ReportLogFile}}
	_ = piperutils.PersistReportsAndLinks("sapURLScan", workdir, utils, reports, nil)

	if countInsecureURLs > 0 {
		log.SetErrorCategory(log.ErrorCompliance)
		return fmt.Errorf("Found %d insecure URLs in %s", countInsecureURLs, config.ProxyLogFile)
	}

	log.Entry().Info("sapURLScan step was successful.")
	return nil
}
