package dwc

import (
	"fmt"
	"path"
	"sync"

	"github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"k8s.io/utils/strings/slices"
)

const (
	githubRepoURLMetadataEntry    = "githubRepoURL"
	commitIdMetadataEntry         = "headCommitId"
	gitBranchMetadataEntry        = "gitBranch"
	jobURLMetadataEntry           = "jobURL"
	buildURLMetadataEntry         = "buildURL"
	buildIDMetadataEntry          = "buildID"
	orchestratorTypeMetadataEntry = "orchestratorType"
)

var (
	supportedOrchestrators = []string{"Azure", "Jenkins", "GitHubActions"}
	once                   sync.Once
	collectorInitErr       error
)

// MetadataCollector is an abstraction that collects all build metadata needed in the dwc context
type MetadataCollector interface {
	GetMetadataEntry(element string) (string, error)
}

type DefaultMetadataCollector struct {
	orchestrator.ConfigProvider
	piperenv.CPEMap
}

func (collector *DefaultMetadataCollector) GetMetadataEntry(element string) (string, error) {
	if err := collector.init(); err != nil {
		return "", fmt.Errorf("failed to initialize DefaultMetadataCollector: %w", err)
	}
	switch element {
	case githubRepoURLMetadataEntry:
		return collector.resolveGitEntity("url")
	case commitIdMetadataEntry:
		return collector.resolveGitEntity(element)
	case gitBranchMetadataEntry:
		return collector.Branch(), nil
	case jobURLMetadataEntry:
		return collector.JobURL(), nil
	case buildURLMetadataEntry:
		return collector.BuildURL(), nil
	case buildIDMetadataEntry:
		return collector.BuildID(), nil
	case orchestratorTypeMetadataEntry:
		return collector.OrchestratorType(), nil
	default:
		return "", fmt.Errorf("unknown metadata entry %s could not be resolved. Please contact the dwc team", element)
	}
}

func (collector *DefaultMetadataCollector) resolveGitEntity(entity string) (string, error) {
	el, err := collector.CPEMap.ParseTemplate(fmt.Sprintf("{{git \"%s\"}}", entity))
	if err != nil {
		return "", fmt.Errorf("failed to get git entity %s from CPE: %w", entity, err)
	}
	return el.String(), nil
}

// Other than the first call every call to init() is a no-op.
// The first call is used for a lazy init of the collector.
func (collector *DefaultMetadataCollector) init() error {
	once.Do(func() {
		collector.CPEMap = piperenv.CPEMap{}
		if err := collector.CPEMap.LoadFromDisk(path.Join(cmd.GeneralConfig.EnvRootPath, "commonPipelineEnvironment")); err != nil {
			collectorInitErr = fmt.Errorf("failed to load values from commonPipelineEnvironment: %w", err)
			return
		}
		if provider, err := orchestrator.GetOrchestratorConfigProvider(nil); err != nil {
			collectorInitErr = fmt.Errorf("creation of new orchestrator specific config provider failed: %w", err)
			return
		} else {
			if !slices.Contains(supportedOrchestrators, provider.OrchestratorType()) {
				collectorInitErr = fmt.Errorf("your pipeline is running on an orchestrator of type %s which is not supported so far. Please contact the DwC team", provider.OrchestratorType())
				return
			}
			collector.ConfigProvider = provider
		}
	})
	return collectorInitErr
}
