package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/project-piper/gpp-local-tester/internal/config"
)

type Builder struct {
	config config.PiperConfig
	Force  bool
}

func New(cfg config.PiperConfig) *Builder {
	return &Builder{
		config: cfg,
		Force:  false,
	}
}

func (b *Builder) Build() error {
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Get absolute path for binary
	binaryPath := b.config.BinaryPath
	if !filepath.IsAbs(binaryPath) {
		absPath, err := filepath.Abs(binaryPath)
		if err != nil {
			return fmt.Errorf("failed to resolve binary path: %w", err)
		}
		binaryPath = absPath
	}

	if !b.Force && fileExists(binaryPath) {
		fmt.Printf("%s Piper binary already exists at: %s\n", blue("ℹ"), binaryPath)
		fmt.Printf("%s Using existing binary (use --force to rebuild)\n", yellow("⚠"))
		return nil
	}

	// Get absolute path for source
	sourcePath := b.config.SourcePath
	if !filepath.IsAbs(sourcePath) {
		absPath, err := filepath.Abs(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to resolve source path: %w", err)
		}
		sourcePath = absPath
	}

	if !dirExists(sourcePath) {
		return fmt.Errorf("piper source not found at: %s", sourcePath)
	}

	fmt.Printf("%s Building Piper from source...\n", blue("ℹ"))
	fmt.Printf("  Source: %s\n", sourcePath)
	fmt.Printf("  Output: %s\n", binaryPath)
	fmt.Println()

	// Get git commit hash
	gitCommit := getGitCommit(sourcePath)
	if gitCommit != "" {
		fmt.Printf("%s Git commit: %s\n", blue("ℹ"), gitCommit)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build command
	ldflags := fmt.Sprintf("-w -s -X github.com/SAP/jenkins-library/cmd.GitCommit=%s", gitCommit)

	cmd := exec.Command("go", "build",
		"-tags", "release",
		"-ldflags", ldflags,
		"-o", binaryPath,
		".")

	cmd.Dir = sourcePath
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS=linux",
		"GOARCH=amd64",
	)

	// Run build
	fmt.Printf("%s Running: go build...\n", blue("ℹ"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("\nBuild output:\n%s\n", string(output))
		return fmt.Errorf("failed to build Piper: %w", err)
	}

	// Make binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Get binary info
	info, err := os.Stat(binaryPath)
	if err == nil {
		fmt.Printf("\n%s Binary size: %.2f MB\n", green("✓"), float64(info.Size())/(1024*1024))
	}

	// Try to get version
	versionCmd := exec.Command(binaryPath, "version")
	if versionOutput, err := versionCmd.CombinedOutput(); err == nil {
		lines := strings.Split(string(versionOutput), "\n")
		if len(lines) > 0 && lines[0] != "" {
			fmt.Printf("%s Version: %s\n", green("✓"), strings.TrimSpace(lines[0]))
		}
	}

	return nil
}

func getGitCommit(dir string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
