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

	// Create .secrets file
	fmt.Printf("%s Creating .secrets file...\n", blue("ℹ"))
	secretsPath := filepath.Join(projectPath, ".secrets")
	secretsData := []string{}
	for key, value := range r.config.Secrets {
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

	// Get absolute path to project root for bind mount
	absProjectRoot, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Build act command
	args := []string{
		"workflow_dispatch",
		"-W", workflow,
		"--bind", fmt.Sprintf("%s:%s", absProjectRoot, absProjectRoot),
		"--container-options", "--add-host=host.docker.internal:host-gateway",
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
