package binary

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
)

type devopsInsightsCommand []string

type cliCommandExecutor func(pathToBinary string, executorFactory BlockingExecutorFactory, targetCmd devopsInsightsCommand) error

func DefaultCLICommandExecutor(pathToBinary string, executorFactory BlockingExecutorFactory, targetCmd devopsInsightsCommand) error {
	executor := executorFactory.CreateExecutor(log.Writer(), log.Writer())
	var execError error
	if err := executor.RunExecutable(pathToBinary, targetCmd...); err != nil {
		execError = fmt.Errorf("execution failed with exit code %d. inner execution error: %w", executor.GetExitCode(), err)
	}
	return execError
}
