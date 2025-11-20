package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/project-piper/gpp-local-tester/internal/config"
)

type StageOutputs struct {
	ActiveStagesMap          map[string]bool            `json:"activeStagesMap"`
	ActiveStepsMap           map[string]map[string]bool `json:"activeStepsMap"`
	OnProductiveBranch       string                     `json:"onProductiveBranch"`
	PipelineOptimization     string                     `json:"pipelineOptimization"`
	IsOptimizedAndScheduled  string                     `json:"isOptimizedAndScheduled"`
	GlobalExtensionsRepo     string                     `json:"globalExtensionsRepository"`
	GlobalExtensionsRef      string                     `json:"globalExtensionsRef"`
	VaultBasePath            string                     `json:"vaultBasePath"`
	VaultPipelineName        string                     `json:"vaultPipelineName"`
	PipelineEnv              string                     `json:"pipelineEnv"`
}

type Orchestrator struct {
	config      *config.Config
	projectPath string
}

func New(cfg *config.Config) *Orchestrator {
	projectPath := cfg.ExampleProject.Path
	if !filepath.IsAbs(projectPath) {
		projectPath = filepath.Join(".", projectPath)
	}

	return &Orchestrator{
		config:      cfg,
		projectPath: projectPath,
	}
}

// RunStagesSequentially runs each stage workflow in sequence
func (o *Orchestrator) RunStagesSequentially(ctx context.Context) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Stage execution order
	stages := []string{
		"Init",
		"Build",
		"Integration",
		"Acceptance",
		"Performance",
		"Promote",
		"Release",
		"Post",
	}

	// Map stages to workflow files
	stageWorkflows := map[string]string{
		"Init":        ".github/workflows/init.yml",
		"Build":       ".github/workflows/build.yml",
		"Integration": ".github/workflows/integration.yml",
		"Acceptance":  ".github/workflows/acceptance.yml",
		"Performance": ".github/workflows/performance.yml",
		"Promote":     ".github/workflows/promote.yml",
		"Release":     ".github/workflows/release.yml",
		"Post":        ".github/workflows/post.yml",
	}

	var outputs StageOutputs

	for i, stage := range stages {
		// Skip Init if we're not on the first stage
		if i > 0 && stage == "Init" {
			continue
		}

		// After Init, check if stage is active
		if i > 0 {
			if !outputs.ActiveStagesMap[stage] {
				fmt.Printf("%s Skipping %s stage (not active)\n", yellow("⊘"), stage)
				continue
			}
		}

		fmt.Printf("\n%s\n", strings.Repeat("━", 80))
		fmt.Printf("  Running %s Stage\n", stage)
		fmt.Printf("%s\n\n", strings.Repeat("━", 80))

		// Run the stage
		workflowFile := stageWorkflows[stage]
		if err := o.runStage(ctx, stage, workflowFile, &outputs); err != nil {
			return fmt.Errorf("%s stage failed: %w", stage, err)
		}

		// Read outputs after each stage
		if err := o.readStageOutputs(&outputs); err != nil {
			return fmt.Errorf("failed to read %s stage outputs: %w", stage, err)
		}

		fmt.Printf("%s %s stage completed successfully\n", green("✓"), stage)
	}

	return nil
}

// runStage executes a single stage workflow with act
func (o *Orchestrator) runStage(ctx context.Context, stageName, workflowFile string, inputs *StageOutputs) error {
	blue := color.New(color.FgBlue).SprintFunc()

	// Build act command
	args := []string{
		"workflow_dispatch",
		"-W", workflowFile,
		"--container-options=--add-host=host.docker.internal:host-gateway",
	}

	// Add inputs as environment variables for stages after Init
	env := os.Environ()
	if stageName != "Init" && inputs != nil {
		// Convert outputs to environment variables that the workflow can access
		env = append(env,
			fmt.Sprintf("INPUT_ON-PRODUCTIVE-BRANCH=%s", inputs.OnProductiveBranch),
			fmt.Sprintf("INPUT_PIPELINE-OPTIMIZATION=%s", inputs.PipelineOptimization),
			fmt.Sprintf("INPUT_VAULT-BASE-PATH=%s", inputs.VaultBasePath),
			fmt.Sprintf("INPUT_VAULT-PIPELINE-NAME=%s", inputs.VaultPipelineName),
		)

		// Add active stages/steps as JSON
		if activeStagesJSON, err := json.Marshal(inputs.ActiveStagesMap); err == nil {
			env = append(env, fmt.Sprintf("INPUT_ACTIVE-STAGES-MAP=%s", string(activeStagesJSON)))
		}
		if activeStepsJSON, err := json.Marshal(inputs.ActiveStepsMap); err == nil {
			env = append(env, fmt.Sprintf("INPUT_ACTIVE-STEPS-MAP=%s", string(activeStepsJSON)))
		}

		if inputs.PipelineEnv != "" {
			env = append(env, fmt.Sprintf("INPUT_PIPELINE-ENV=%s", inputs.PipelineEnv))
		}
	}

	cmd := exec.CommandContext(ctx, o.config.Act.BinaryPath, args...)
	cmd.Dir = o.projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env

	fmt.Printf("%s Running: act %s\n", blue("ℹ"), strings.Join(args, " "))
	fmt.Println()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("act failed: %w", err)
	}

	return nil
}

// readStageOutputs reads the stage outputs from .pipeline files
func (o *Orchestrator) readStageOutputs(outputs *StageOutputs) error {
	// Read active stages map
	stageOutPath := filepath.Join(o.projectPath, ".pipeline", "stage_out.json")
	if data, err := os.ReadFile(stageOutPath); err == nil {
		if err := json.Unmarshal(data, &outputs.ActiveStagesMap); err != nil {
			return fmt.Errorf("failed to parse stage_out.json: %w", err)
		}
	}

	// Read active steps map
	stepOutPath := filepath.Join(o.projectPath, ".pipeline", "step_out.json")
	if data, err := os.ReadFile(stepOutPath); err == nil {
		if err := json.Unmarshal(data, &outputs.ActiveStepsMap); err != nil {
			return fmt.Errorf("failed to parse step_out.json: %w", err)
		}
	}

	// Read stage config (contains various outputs)
	stageConfigPath := filepath.Join(o.projectPath, "stage-config.json")
	if data, err := os.ReadFile(stageConfigPath); err == nil {
		var stageConfig map[string]interface{}
		if err := json.Unmarshal(data, &stageConfig); err == nil {
			// Extract values
			if val, ok := stageConfig["onProductiveBranch"].(bool); ok {
				outputs.OnProductiveBranch = fmt.Sprintf("%t", val)
			}
			if val, ok := stageConfig["pipelineOptimization"].(bool); ok {
				outputs.PipelineOptimization = fmt.Sprintf("%t", val)
			}
			if val, ok := stageConfig["globalExtensionsRepository"].(string); ok {
				outputs.GlobalExtensionsRepo = val
			}
			if val, ok := stageConfig["vaultBasePath"].(string); ok {
				outputs.VaultBasePath = val
			}
			if val, ok := stageConfig["vaultPipelineName"].(string); ok {
				outputs.VaultPipelineName = val
			}
		}
	}

	// Read pipeline environment if it exists
	pipelineEnvPath := filepath.Join(o.projectPath, ".pipeline", "commonPipelineEnvironment")
	if _, err := os.Stat(pipelineEnvPath); err == nil {
		// For now, we'll skip reading the full CPE as it's complex base64-gzipped data
		// The workflows will read it directly from disk
	}

	return nil
}
