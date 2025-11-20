package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/project-piper/gpp-local-tester/internal/builder"
	"github.com/project-piper/gpp-local-tester/internal/config"
	"github.com/spf13/cobra"
)

var (
	force bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build Piper binary",
	Long:  `Build the Piper binary from local jenkins-library source.`,
	RunE:  buildPiper,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVarP(&force, "force", "f", false, "Force rebuild even if binary exists")
}

func buildPiper(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	printHeader("Building Piper Binary")

	// Load configuration
	fmt.Printf("%s Loading configuration from %s\n", blue("ℹ"), configFile)
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Build Piper
	b := builder.New(cfg.Piper)
	b.Force = force

	if err := b.Build(); err != nil {
		return fmt.Errorf("failed to build Piper: %w", err)
	}

	fmt.Printf("\n%s Piper binary built successfully\n", green("✓"))
	return nil
}
