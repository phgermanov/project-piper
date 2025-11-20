package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy/agent"
)

type sapCollectPolicyResultsUtils interface {
	// command.ExecRunner
	Install(version string) error
	Execute(parameter []string, environmentVariables []string) error
	WaitPeriod() time.Duration
}

type sapCollectPolicyResultsUtilsBundle struct {
	*command.Command
	*piperutils.Files
	agent.Agent
}

func (s *sapCollectPolicyResultsUtilsBundle) WaitPeriod() time.Duration {
	return 20 * time.Second
}

func newSapCollectPolicyResultsUtils(config sapCollectPolicyResultsOptions) (sapCollectPolicyResultsUtils, error) {
	resolver, err := agent.NewResolverByOrchestrator(orchestrator.DetectOrchestrator(), config.GithubToken)
	if err != nil {
		return nil, err
	}

	agent := agent.NewAgent(resolver)

	utils := sapCollectPolicyResultsUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Agent:   agent,
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils, nil
}

func sapCollectPolicyResults(config sapCollectPolicyResultsOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils, err := newSapCollectPolicyResultsUtils(config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err = runSapCollectPolicyResults(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSapCollectPolicyResults(config *sapCollectPolicyResultsOptions, telemetryData *telemetry.CustomData, utils sapCollectPolicyResultsUtils) error {
	// map & validate parameters
	// done in the beginning, because without valid parameters it makes no sense to continue
	params, err := mapCollectPolicyResultsParams(config)
	if err != nil {
		return err
	}

	envVars := mapCollectPolicyResultsEnvVars(config)

	// install the agent
	err = utils.Install(config.CumulusPolicyAgentVersion)
	if err != nil {
		return err
	}

	// execute without compliance check -> no retries necessary
	if !config.ValidateCompliance {
		return utils.Execute(params, envVars)
	}

	// execute check for missing policy results with retry in case it is configured
	if config.MaxWait > 0 {
		params = append(params, "--validate-missing")
		// check is to be performed every 20 seconds
		for i := 0; i <= config.MaxWait*3; i++ {
			err = utils.Execute(params, envVars)
			// only retry in case a policy is missing not in case of policy errors
			if err == nil {
				break
			}
			log.Entry().Infof("Some policy results are missing, retrying in 20 seconds (%v/%v)", i+1, config.MaxWait*3)
			time.Sleep(utils.WaitPeriod())
		}
	}

	// check for compliance = complete and OK
	params = append(params, "--validate-compliance")
	return utils.Execute(params, envVars)
}

func mapCollectPolicyResultsParams(config *sapCollectPolicyResultsOptions) ([]string, error) {
	log.Entry().Debugf("Mapping and validation of parameters...")

	params := []string{"collectPolicyResults"}

	if len(config.PipelineRuns) == 0 {
		// pipelineRuns is mandatory
		return nil, fmt.Errorf("configuration parameter PipelineRuns is missing")
	}
	params = append(params, config.PipelineRuns...)

	if config.Token != "" {
		// if token is provided, it will be used as access token and provided to cli as environment variable (see mapCollectPolicyResultsEnvVars)
	} else if config.JSONKeyFilePath != "" {
		params = append(params, "--credentials-file="+config.JSONKeyFilePath)
	} else {
		// jsonKeyFilePath or storage token is mandatory
		return nil, fmt.Errorf("configuration parameter JSONKeyFilePath and Token is missing")
	}

	if len(config.CentralPolicyKeys) != 0 {
		params = append(params, "--central-policy-keys="+strings.Join(config.CentralPolicyKeys, ","))
	}

	// customPolicies
	if len(config.CustomPolicyKeys) != 0 {
		params = append(params, "--custom-policy-keys="+strings.Join(config.CustomPolicyKeys, ","))
	}

	// resultFile
	if config.ResultFile != "" {
		params = append(params, "--out="+config.ResultFile)
	}

	log.Entry().Debugf("Mapping and validation of parameters successful!")

	return params, nil
}

func mapCollectPolicyResultsEnvVars(config *sapCollectPolicyResultsOptions) []string {
	envVars := []string{}

	if config.Token != "" {
		envVars = append(envVars, "GCP_STORAGE_ACCESS_TOKEN="+config.Token)
	}

	return envVars
}
