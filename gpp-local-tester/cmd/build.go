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
	Short: "Build Piper binaries",
	Long:  `Build the Piper and SAP Piper binaries from local source.`,
	RunE:  buildPiper,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVarP(&force, "force", "f", false, "Force rebuild even if binary exists")
}

func buildPiper(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	printHeader("Building Piper Binaries")

	// Load configuration
	fmt.Printf("%s Loading configuration from %s\n", blue("ℹ"), configFile)
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Build OS Piper
	fmt.Printf("\n%s Building OS Piper...\n", blue("═"))
	b := builder.New(cfg.Piper)
	b.Force = force

	if err := b.Build(); err != nil {
		return fmt.Errorf("failed to build Piper: %w", err)
	}

	fmt.Printf("%s OS Piper binary built successfully\n", green("✓"))

	// Build SAP Piper
	fmt.Printf("\n%s Building SAP Piper...\n", blue("═"))
	sapB := builder.New(cfg.SapPiper)
	sapB.Force = force
	sapB.BinaryName = "sap-piper"

	if err := sapB.Build(); err != nil {
		return fmt.Errorf("failed to build SAP Piper: %w", err)
	}

	fmt.Printf("%s SAP Piper binary built successfully\n", green("✓"))
	fmt.Printf("\n%s All binaries built successfully\n", green("✓"))
	return nil
}
