package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/project-piper/gpp-local-tester/internal/builder"
	"github.com/project-piper/gpp-local-tester/internal/config"
	"github.com/project-piper/gpp-local-tester/internal/mock"
	"github.com/project-piper/gpp-local-tester/internal/orchestrator"
	"github.com/project-piper/gpp-local-tester/internal/runner"
	"github.com/spf13/cobra"
)

var (
	skipBuild      bool
	skipMockServer bool
	sequential     bool
	workflow       string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run GPP pipeline locally",
	Long:  `Run the GPP pipeline locally using act with mock services and local Piper binary.`,
	RunE:  runPipeline,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&skipBuild, "skip-build", false, "Skip building Piper binary")
	runCmd.Flags().BoolVar(&skipMockServer, "skip-mock", false, "Skip starting mock server")
	runCmd.Flags().BoolVar(&sequential, "sequential", true, "Run stages sequentially (workaround for act limitation)")
	runCmd.Flags().StringVarP(&workflow, "workflow", "w", "", "Workflow file to run (relative to example project)")
}

func runPipeline(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nInterrupt received, cleaning up...")
		cancel()
	}()

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	printHeader("GPP Local Testing Tool")

	// Load configuration
	fmt.Printf("%s Loading configuration from %s\n", blue("ℹ"), configFile)
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	fmt.Printf("%s Configuration loaded\n\n", green("✓"))

	// Check prerequisites
	printHeader("Checking Prerequisites")
	if err := runner.CheckPrerequisites(); err != nil {
		return fmt.Errorf("prerequisites check failed: %w", err)
	}
	fmt.Printf("%s All prerequisites met\n\n", green("✓"))

	// Build Piper binaries
	if !skipBuild {
		printHeader("Building Piper Binaries")

		// Build OS Piper
		fmt.Printf("%s Building OS Piper...\n", blue("ℹ"))
		b := builder.New(cfg.Piper)
		if err := b.Build(); err != nil {
			return fmt.Errorf("failed to build OS Piper: %w", err)
		}
		fmt.Printf("%s OS Piper binary ready\n", green("✓"))

		// Build SAP Piper
		fmt.Printf("%s Building SAP Piper...\n", blue("ℹ"))
		sapB := builder.New(cfg.SapPiper)
		sapB.BinaryName = "sap-piper"
		if err := sapB.Build(); err != nil {
			return fmt.Errorf("failed to build SAP Piper: %w", err)
		}
		fmt.Printf("%s SAP Piper binary ready\n\n", green("✓"))
	} else {
		fmt.Printf("%s Skipping Piper builds\n\n", yellow("⚠"))
	}

	// Start mock server
	var mockServer *mock.Server
	if cfg.MockServices.Enabled && !skipMockServer {
		printHeader("Starting Mock Server")
		mockServer = mock.NewServer(cfg.MockServices)
		if err := mockServer.Start(); err != nil {
			return fmt.Errorf("failed to start mock server: %w", err)
		}
		defer mockServer.Stop()
		fmt.Printf("%s Mock server running at http://%s:%d\n\n", green("✓"), cfg.MockServices.Host, cfg.MockServices.Port)
	} else {
		fmt.Printf("%s Mock server disabled\n\n", yellow("⚠"))
	}

	// Prepare act environment
	printHeader("Preparing Act Environment")
	r := runner.New(cfg)
	if err := r.PrepareEnvironment(); err != nil {
		return fmt.Errorf("failed to prepare environment: %w", err)
	}
	fmt.Printf("%s Act environment prepared\n\n", green("✓"))

	// Run act
	printHeader("Running GPP Pipeline with Act")

	if sequential {
		// Run stages sequentially (workaround for act's limitation with multiple reusable workflows)
		fmt.Printf("%s Running stages sequentially to work around act limitation\n\n", blue("ℹ"))
		orch := orchestrator.New(cfg)
		if err := orch.RunStagesSequentially(ctx); err != nil {
			fmt.Printf("\n%s Pipeline failed\n", red("✗"))
			return err
		}
	} else {
		// Run the main workflow (may fail with act's reusable workflow limitation)
		workflowPath := workflow
		if workflowPath == "" {
			workflowPath = cfg.ExampleProject.WorkflowFile
		}

		if err := r.Run(ctx, workflowPath); err != nil {
			fmt.Printf("\n%s Pipeline failed\n", red("✗"))
			return err
		}
	}

	fmt.Printf("\n%s Pipeline completed successfully!\n\n", green("✓"))

	// Print summary
	printHeader("Test Summary")
	if mockServer != nil {
		stats := mockServer.GetStats()
		fmt.Printf("Mock Server Statistics:\n")
		fmt.Printf("  Total requests: %d\n", stats.TotalRequests)
		fmt.Printf("  By method:\n")
		for method, count := range stats.MethodCounts {
			fmt.Printf("    %s: %d\n", method, count)
		}
	}

	return nil
}
