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

type sapExecuteCentralPolicyUtils interface {
	command.ExecRunner

	Install(version string) error
	Execute(parameter []string, environmentVariables []string) error
	FileExists(filename string) (bool, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte, perm os.FileMode) error

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The sapExecuteCentralPolicyUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type sapExecuteCentralPolicyUtilsBundle struct {
	*command.Command
	*piperutils.Files
	agent.Agent
}

func newSapExecuteCentralPolicyUtils(config sapExecuteCentralPolicyOptions) (sapExecuteCentralPolicyUtils, error) {
	resolver, err := agent.NewResolverByOrchestrator(orchestrator.DetectOrchestrator(), config.GithubToken)
	if err != nil {
		return nil, err
	}

	agent := agent.NewAgent(resolver)
	utils := sapExecuteCentralPolicyUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Agent:   agent,
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils, nil
}

func sapExecuteCentralPolicy(config sapExecuteCentralPolicyOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils, err := newSapExecuteCentralPolicyUtils(config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err = runSapExecuteCentralPolicy(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSapExecuteCentralPolicy(config *sapExecuteCentralPolicyOptions, telemetryData *telemetry.CustomData, utils sapExecuteCentralPolicyUtils) error {
	_ = telemetryData
	// map & validate parameters
	// done in the beginning, because without valid parameters it makes no sense to continue
	params, err := mapExecuteCentralPolicyParams(config)
	if err != nil {
		return err
	}

	synchronizeParams, err := mapSynchronizeCentralPolicyBundlesParams(config)
	if err != nil {
		return err
	}

	synchronizeEnvVars := mapSynchronizeCentralPolicyBundlesEnvVars(config)

	// install the agent
	err = utils.Install(config.CumulusPolicyAgentVersion)
	if err != nil {
		return err
	}

	// download bundles
	err = utils.Execute(synchronizeParams, synchronizeEnvVars)
	if err != nil {
		return err
	}

	// execute
	err = utils.Execute(params, nil)
	if err != nil {
		return err
	}

	if config.GenerateJunitReport {
		err = generateJunitReportOfCentralPolicy(params, utils)
	}

	return err
}

func mapExecuteCentralPolicyParams(config *sapExecuteCentralPolicyOptions) ([]string, error) {
	log.Entry().Debugf("Mapping and validation of parameters...")

	params := []string{"execute", "--type=central"}

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

	if config.CentralPolicyPath != "" {
		params = append(params, "--directory="+config.CentralPolicyPath)
	}

	if !config.FailOnPolicyViolation {
		params = append(params, "--ignore-compliance=true")
	}

	var resultFile string
	if config.ResultFile != "" {
		resultFile = config.ResultFile
	} else {
		resultFile = path.Join("./policy-result", config.PolicyKey, "result.json")
	}
	resultFile = strings.ReplaceAll(resultFile, "<policyKey>", config.PolicyKey)
	params = append(params, paramOut+resultFile)

	log.Entry().Debugf("Mapping and validation of parameters successful!")

	return params, nil
}

func mapSynchronizeCentralPolicyBundlesParams(config *sapExecuteCentralPolicyOptions) ([]string, error) {
	params := []string{"synchronizePolicyBundles"}

	if config.CentralPolicyPath != "" {
		params = append(params, "--directory="+config.CentralPolicyPath)
	}

	if config.CumulusPolicyBucket != "" {
		params = append(params, "--policy-bucket="+config.CumulusPolicyBucket)
	}

	if config.Token != "" {
		// if token is provided, it will be used as access token and provided to cli as environment variable (see mapSynchronizeCentralPolicyBundlesEnvVars)
	} else if config.JSONKeyFilePath != "" {
		params = append(params, "--credentials-file="+config.JSONKeyFilePath)
	} else {
		// jsonKeyFilePath or storage token is mandatory
		return nil, fmt.Errorf("configuration parameter JSONKeyFilePath and Token is missing")
	}

	return params, nil
}

func mapSynchronizeCentralPolicyBundlesEnvVars(config *sapExecuteCentralPolicyOptions) []string {
	envVars := []string{}

	if config.Token != "" {
		envVars = append(envVars, "GCP_STORAGE_ACCESS_TOKEN="+config.Token)
	}

	return envVars
}

func generateJunitReportOfCentralPolicy(params []string, utils sapExecuteCentralPolicyUtils) error {
	resultFilePath, err := getResultFilePath(params)
	if err != nil {
		return err
	}

	reporter := report.NewPolicyJunitReporter(policy.Central, resultFilePath, utils)

	return reporter.ReportPolicyResult()
}
