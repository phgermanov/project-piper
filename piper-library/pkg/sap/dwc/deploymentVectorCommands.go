package dwc

import "fmt"

// Flag collection for vector deployment commands

var deploymentVectorBaseCommand = []string{"deployment", "vector"}

const (
	waitForVectorDeploymentSubcommand = "wait-for"
	watchVectorDeploymentSubcommand   = "watch"
	addVectorUsageSubcommand          = "add-usage"
	removeVectorUsageSubcommand       = "remove-usage"
	stageFlag                         = "--stage=%s"
	vectorIDFlag                      = "--vector=%s"
	landscapeFlag                     = "--landscape=%s"
	usageNameFlag                     = "--usage=%s"
	expiresAtFlag                     = "--expiresAt=%s"
	logsFlag                          = "--logs=%s"
	roiFlag                           = "--resource-of-interest=%s"
	timeoutFlag                       = "--timeout=%dm"
)

// Default values for vector deployment command flags
const (
	pipelineUsageName            = "sapDwCStageRelease"
	watchVectorDeploymentLogMode = "live"
)

func newWaitForDeploymentCommand(stage, vectorID string) dwcCommand {
	args := append(deploymentVectorBaseCommand, []string{waitForVectorDeploymentSubcommand, fmt.Sprintf(stageFlag, stage), fmt.Sprintf(vectorIDFlag, vectorID)}...)
	args = appendOutputFormat(args, outputFormatJSON)
	return args
}

func newAddVectorUsageCommand(landscape, vectorID, expiry string) dwcCommand {
	return append(deploymentVectorBaseCommand, []string{addVectorUsageSubcommand, fmt.Sprintf(landscapeFlag, landscape), fmt.Sprintf(vectorIDFlag, vectorID), fmt.Sprintf(expiresAtFlag, expiry), fmt.Sprintf(usageNameFlag, pipelineUsageName)}...)
}

func newRemoveVectorUsageCommand(landscape, vectorID, expiry string) dwcCommand {
	return append(deploymentVectorBaseCommand, []string{removeVectorUsageSubcommand, fmt.Sprintf(landscapeFlag, landscape), fmt.Sprintf(vectorIDFlag, vectorID), fmt.Sprintf(expiresAtFlag, expiry), fmt.Sprintf(usageNameFlag, pipelineUsageName)}...)
}

func newWatchVectorDeploymentCommand(landscape, vectorID string, artifactDescriptor ArtifactDescriptor) dwcCommand {
	cmd := append(deploymentVectorBaseCommand, []string{watchVectorDeploymentSubcommand, fmt.Sprintf(landscapeFlag, landscape), fmt.Sprintf(vectorIDFlag, vectorID), fmt.Sprintf(logsFlag, watchVectorDeploymentLogMode), fmt.Sprintf(timeoutFlag, stageWatchLockMinutes)}...)
	if artifactDescriptor.watchROIOnly() {
		cmd = append(cmd, fmt.Sprintf(roiFlag, artifactDescriptor.GetResourceName()))
	}
	return cmd
}
