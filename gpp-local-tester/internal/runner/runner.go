package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/project-piper/gpp-local-tester/internal/config"
	"gopkg.in/yaml.v3"
)

type Runner struct {
	config *config.Config
}

func New(cfg *config.Config) *Runner {
	return &Runner{
		config: cfg,
	}
}

// CheckPrerequisites verifies required tools are installed
func CheckPrerequisites() error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	required := []struct {
		name    string
		command string
		args    []string
	}{
		{"act", "act", []string{"--version"}},
		{"docker", "docker", []string{"--version"}},
		{"go", "go", []string{"version"}},
	}

	missing := []string{}

	for _, req := range required {
		cmd := exec.Command(req.command, req.args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("%s %s is not installed\n", red("✗"), req.name)
			missing = append(missing, req.name)
		} else {
			version := strings.Split(string(output), "\n")[0]
			fmt.Printf("%s %s found: %s\n", green("✓"), req.name, version)
		}
	}

	// Check if Docker daemon is running
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s Docker daemon is not running\n", red("✗"))
		missing = append(missing, "docker (daemon)")
	} else {
		fmt.Printf("%s Docker daemon is running\n", green("✓"))
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing prerequisites: %v", missing)
	}

	return nil
}

// PrepareEnvironment sets up the act environment
func (r *Runner) PrepareEnvironment() error {
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	projectPath := r.config.ExampleProject.Path
	if !filepath.IsAbs(projectPath) {
		projectPath = filepath.Join(".", projectPath)
	}

	// Copy local Piper binary to example project
	fmt.Printf("%s Copying local Piper binary to project...\n", blue("ℹ"))
	if err := r.copyPiperBinary(); err != nil {
		return fmt.Errorf("failed to copy Piper binary: %w", err)
	}
	fmt.Printf("%s Local Piper binary copied\n", green("✓"))

	// Copy and adapt GPP workflows for local execution
	fmt.Printf("%s Preparing GPP workflows for local execution...\n", blue("ℹ"))
	if err := r.prepareGPPWorkflows(); err != nil {
		return fmt.Errorf("failed to prepare GPP workflows: %w", err)
	}
	fmt.Printf("%s GPP workflows prepared\n", green("✓"))

	// Create .secrets file
	fmt.Printf("%s Creating .secrets file...\n", blue("ℹ"))
	secretsPath := filepath.Join(projectPath, ".secrets")
	secretsData := []string{}
	for key, value := range r.config.Secrets {
		// Use system GITHUB_TOKEN if available for act to download actions
		if key == "GITHUB_TOKEN" {
			if systemToken := os.Getenv("GITHUB_TOKEN"); systemToken != "" {
				value = systemToken
				fmt.Printf("%s Using system GITHUB_TOKEN for action downloads\n", blue("ℹ"))
			}
		}
		secretsData = append(secretsData, fmt.Sprintf("%s=%s", key, value))
	}
	if err := os.WriteFile(secretsPath, []byte(strings.Join(secretsData, "\n")), 0600); err != nil {
		return fmt.Errorf("failed to create .secrets: %w", err)
	}
	fmt.Printf("%s .secrets created\n", green("✓"))

	// Create .env file
	fmt.Printf("%s Creating .env file...\n", blue("ℹ"))
	envPath := filepath.Join(projectPath, ".env")
	envData := []string{}
	for key, value := range r.config.Act.Env {
		envData = append(envData, fmt.Sprintf("%s=%s", key, value))
	}
	for key, value := range r.config.Variables {
		envData = append(envData, fmt.Sprintf("%s=%s", key, value))
	}
	if err := os.WriteFile(envPath, []byte(strings.Join(envData, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to create .env: %w", err)
	}
	fmt.Printf("%s .env created\n", green("✓"))

	// Create .actrc file
	fmt.Printf("%s Creating .actrc file...\n", blue("ℹ"))
	actrcPath := filepath.Join(projectPath, ".actrc")
	actrcData := []string{
		fmt.Sprintf("-P %s", r.config.Act.Platform),
		fmt.Sprintf("--container-architecture %s", r.config.Act.ContainerArchitecture),
		"--env-file .env",
		"--secret-file .secrets",
	}
	actrcData = append(actrcData, r.config.Act.Flags...)
	if err := os.WriteFile(actrcPath, []byte(strings.Join(actrcData, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to create .actrc: %w", err)
	}
	fmt.Printf("%s .actrc created\n", green("✓"))

	// Update pipeline config with overrides
	fmt.Printf("%s Updating pipeline configuration...\n", blue("ℹ"))
	if err := r.updatePipelineConfig(); err != nil {
		return fmt.Errorf("failed to update pipeline config: %w", err)
	}
	fmt.Printf("%s Pipeline config updated\n", green("✓"))

	return nil
}

func (r *Runner) updatePipelineConfig() error {
	projectPath := r.config.ExampleProject.Path
	if !filepath.IsAbs(projectPath) {
		projectPath = filepath.Join(".", projectPath)
	}

	configPath := filepath.Join(projectPath, ".pipeline", "config.yml")

	// Read existing config
	var existingConfig map[string]any
	if data, err := os.ReadFile(configPath); err == nil {
		yaml.Unmarshal(data, &existingConfig)
	} else {
		existingConfig = make(map[string]any)
	}

	// Merge with overrides
	for key, value := range r.config.PipelineConfig {
		existingConfig[key] = value
	}

	// Write back
	data, err := yaml.Marshal(existingConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// Run executes the act command
func (r *Runner) Run(ctx context.Context, workflow string) error {
	projectPath := r.config.ExampleProject.Path
	if !filepath.IsAbs(projectPath) {
		projectPath = filepath.Join(".", projectPath)
	}

	// Build act command
	// Note: Volume mounting for local Piper binary needs to be handled differently
	// Act doesn't support direct volume mounting like Docker CLI
	args := []string{
		"workflow_dispatch",
		"-W", workflow,
		"--container-options=--add-host=host.docker.internal:host-gateway",
	}

	cmd := exec.CommandContext(ctx, r.config.Act.BinaryPath, args...)
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	blue := color.New(color.FgBlue).SprintFunc()
	fmt.Printf("%s Running: act %s\n", blue("ℹ"), strings.Join(args, " "))
	fmt.Println()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("act failed: %w", err)
	}

	return nil
}

// copyPiperBinary copies the local SAP Piper binary to the example project
func (r *Runner) copyPiperBinary() error {
	// Use SAP Piper binary which contains SAP-specific commands like sapPipelineInit
	sourceBinary := r.config.SapPiper.BinaryPath
	if !filepath.IsAbs(sourceBinary) {
		absPath, err := filepath.Abs(sourceBinary)
		if err != nil {
			return fmt.Errorf("failed to resolve source binary path: %w", err)
		}
		sourceBinary = absPath
	}

	// Check if source binary exists
	if _, err := os.Stat(sourceBinary); err != nil {
		return fmt.Errorf("source SAP Piper binary not found at %s: %w", sourceBinary, err)
	}

	// Get project path
	projectPath := r.config.ExampleProject.Path
	if !filepath.IsAbs(projectPath) {
		projectPath = filepath.Join(".", projectPath)
	}

	// Create destination directory
	destDir := filepath.Join(projectPath, ".piper")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create .piper directory: %w", err)
	}

	// Copy SAP Piper binary as "piper" (project-piper-action expects it at this name)
	destBinary := filepath.Join(destDir, "piper")
	sourceData, err := os.ReadFile(sourceBinary)
	if err != nil {
		return fmt.Errorf("failed to read source binary: %w", err)
	}

	if err := os.WriteFile(destBinary, sourceData, 0755); err != nil {
		return fmt.Errorf("failed to write destination binary: %w", err)
	}

	return nil
}

// prepareGPPWorkflows copies and adapts GPP workflows for local execution
func (r *Runner) prepareGPPWorkflows() error {
	// Find piper-pipeline-github directory
	gppPath := "../piper-pipeline-github"
	absGPPPath, err := filepath.Abs(gppPath)
	if err != nil {
		return fmt.Errorf("failed to resolve GPP path: %w", err)
	}

	// Check if GPP workflows exist
	gppWorkflowsPath := filepath.Join(absGPPPath, ".github", "workflows")
	if _, err := os.Stat(gppWorkflowsPath); err != nil {
		return fmt.Errorf("GPP workflows not found at %s: %w", gppWorkflowsPath, err)
	}

	// Get project path
	projectPath := r.config.ExampleProject.Path
	if !filepath.IsAbs(projectPath) {
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			return fmt.Errorf("failed to resolve project path: %w", err)
		}
		projectPath = absPath
	}

	// Create workflows directory
	destWorkflowsPath := filepath.Join(projectPath, ".github", "workflows")
	if err := os.MkdirAll(destWorkflowsPath, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	// List of workflow files to copy and adapt
	workflowFiles := []string{
		"sap-piper-workflow.yml",
		"init.yml",
		"build.yml",
		"integration.yml",
		"acceptance.yml",
		"performance.yml",
		"promote.yml",
		"release.yml",
		"post.yml",
	}

	for _, workflowFile := range workflowFiles {
		sourceFile := filepath.Join(gppWorkflowsPath, workflowFile)
		destFile := filepath.Join(destWorkflowsPath, workflowFile)

		// Read source workflow
		content, err := os.ReadFile(sourceFile)
		if err != nil {
			// Skip if file doesn't exist
			continue
		}

		// Adapt workflow for local execution
		adaptedContent := r.adaptWorkflowForLocal(string(content))

		// Write adapted workflow
		if err := os.WriteFile(destFile, []byte(adaptedContent), 0644); err != nil {
			return fmt.Errorf("failed to write workflow %s: %w", workflowFile, err)
		}
	}

	return nil
}

// adaptWorkflowForLocal modifies a workflow to use local actions and remove external dependencies
func (r *Runner) adaptWorkflowForLocal(content string) string {
	// Replace SAP/project-piper-action with local action
	content = strings.ReplaceAll(content, "uses: SAP/project-piper-action@v1.22", "uses: ./.github/actions/local-piper-action")

	// Replace actions/checkout with specific commit (act may have issues with latest)
	content = strings.ReplaceAll(content, "uses: actions/checkout@08eba0b27e820071cde6df949e0beb9ba4906955", "uses: actions/checkout@v4")
	content = strings.ReplaceAll(content, "uses: actions/checkout@v4", "uses: actions/checkout@v3")

	// Replace actions/setup-go with v3 (more compatible with act)
	content = strings.ReplaceAll(content, "uses: actions/setup-go@44694675825211faa026b3c33043df3e48a5fa00", "uses: actions/setup-go@v3")
	content = strings.ReplaceAll(content, "uses: actions/setup-go@v6.0.0", "uses: actions/setup-go@v3")

	// Replace actions/upload-artifact with v3 (already present)
	// No change needed

	// Comment out system trust action (not needed for local testing)
	content = strings.ReplaceAll(content,
		"- name: Retrieve System Trust session token",
		"# - name: Retrieve System Trust session token (disabled for local testing)")
	content = strings.ReplaceAll(content,
		"id: system_trust",
		"# id: system_trust")
	content = strings.ReplaceAll(content,
		"uses: project-piper/system-trust-composite-action@",
		"# uses: project-piper/system-trust-composite-action@")

	// Comment out extensibility actions (not needed for basic testing)
	content = strings.ReplaceAll(content,
		"uses: project-piper/piper-pipeline-github/.github/workflows@main",
		"# uses: project-piper/piper-pipeline-github/.github/workflows@main (disabled for local testing)")

	// Change runs-on from self-hosted to ubuntu-latest for act
	content = strings.ReplaceAll(content,
		"runs-on: ${{ fromJSON(inputs.runs-on) }}",
		"runs-on: ubuntu-latest")

	return content
}
