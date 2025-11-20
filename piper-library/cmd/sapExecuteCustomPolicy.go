package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy/agent"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy/report"
)

const paramOut = "--out="

type sapExecuteCustomPolicyUtils interface {
	// command.ExecRunner
	Install(version string) error
	Execute(parameter []string, environmentVariables []string) error
	FileExists(filename string) (bool, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte, perm os.FileMode) error
}

type sapExecuteCustomPolicyUtilsBundle struct {
	*command.Command
	*piperutils.Files
	agent.Agent
}

func newSapExecuteCustomPolicyUtils(config sapExecuteCustomPolicyOptions) (sapExecuteCustomPolicyUtils, error) {
	resolver, err := agent.NewResolverByOrchestrator(orchestrator.DetectOrchestrator(), config.GithubToken)
	if err != nil {
		return nil, err
	}

	agent := agent.NewAgent(resolver)

	utils := sapExecuteCustomPolicyUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Agent:   agent,
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils, nil
}

func sapExecuteCustomPolicy(config sapExecuteCustomPolicyOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils, err := newSapExecuteCustomPolicyUtils(config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err = runSapExecuteCustomPolicy(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSapExecuteCustomPolicy(config *sapExecuteCustomPolicyOptions, telemetryData *telemetry.CustomData, utils sapExecuteCustomPolicyUtils) error {
	// map & validate parameters
	// done in the beginning, because without valid parameters it makes no sense to continue
	params, err := mapExecuteCustomPolicyParams(config)
	if err != nil {
		return err
	}

	// install the agent
	err = utils.Install(config.CumulusPolicyAgentVersion)
	if err != nil {
		return err
	}

	// execute
	err = utils.Execute(params, nil)
	if err != nil {
		return err
	}

	if config.GenerateJunitReport == true {
		err = generateJunitReport(params, utils)
	}

	return err
}

func mapExecuteCustomPolicyParams(config *sapExecuteCustomPolicyOptions) ([]string, error) {
	log.Entry().Debugf("Mapping and validation of parameters...")

	params := []string{"execute", "--type=custom"}

	if config.PolicyKey == "" {
		// policyKey is mandatory
		return nil, fmt.Errorf("configuration parameter policyKey is missing")
	}
	params = append(params, config.PolicyKey)

	if config.EvidenceFile == "" {
		// evidenceFile is mandatory
		return nil, fmt.Errorf("configuration parameter evidenceFile is missing")
	}
	params = append(params, config.EvidenceFile)

	if config.PolicyPath != "" {
		params = append(params, "--directory="+config.PolicyPath)
	}

	if !config.FailOnPolicyViolation {
		params = append(params, "--ignore-compliance=true")
	}

	var resultFile string
	if config.ResultFile != "" {
		resultFile = config.ResultFile
	} else {
		resultFile = path.Join("./custom-policy-result", config.PolicyKey, "result.json")
	}
	resultFile = strings.ReplaceAll(resultFile, "<policyKey>", config.PolicyKey)
	params = append(params, paramOut+resultFile)

	log.Entry().Debugf("Mapping and validation of parameters successful!")

	return params, nil
}

func generateJunitReport(params []string, utils sapExecuteCustomPolicyUtils) error {
	resultFilePath, err := getResultFilePath(params)
	if err != nil {
		return err
	}

	reporter := report.NewPolicyJunitReporter(policy.Custom, resultFilePath, utils)

	return reporter.ReportPolicyResult()
}

func getResultFilePath(params []string) (string, error) {
	var resultFile string
	for _, param := range params {
		if strings.HasPrefix(param, paramOut) {
			resultFile = strings.TrimPrefix(param, paramOut)
		}
	}
	if resultFile == "" {
		return "", fmt.Errorf("resultFile not found")
	}
	return resultFile, nil
}
