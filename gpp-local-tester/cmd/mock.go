package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/project-piper/gpp-local-tester/internal/config"
	"github.com/project-piper/gpp-local-tester/internal/mock"
	"github.com/spf13/cobra"
)

var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Run mock server only",
	Long:  `Start the mock server for external services (Vault, Cumulus, SonarQube, etc.)`,
	RunE:  runMockServer,
}

func init() {
	rootCmd.AddCommand(mockCmd)
}

func runMockServer(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nShutting down mock server...")
		cancel()
	}()

	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	printHeader("GPP Mock Server")

	// Load configuration
	fmt.Printf("%s Loading configuration from %s\n", blue("ℹ"), configFile)
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.MockServices.Enabled {
		return fmt.Errorf("mock services are disabled in config")
	}

	// Start mock server
	mockServer := mock.NewServer(cfg.MockServices)
	if err := mockServer.Start(); err != nil {
		return fmt.Errorf("failed to start mock server: %w", err)
	}
	defer mockServer.Stop()

	fmt.Printf("%s Mock server running at http://%s:%d\n", green("✓"), cfg.MockServices.Host, cfg.MockServices.Port)
	fmt.Printf("%s Configured endpoints: %d\n\n", blue("ℹ"), len(cfg.MockServices.Endpoints))
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Wait for context cancellation
	<-ctx.Done()

	fmt.Println()
	printHeader("Request Statistics")
	stats := mockServer.GetStats()
	fmt.Printf("Total requests: %d\n", stats.TotalRequests)
	if stats.TotalRequests > 0 {
		fmt.Printf("\nBy method:\n")
		for method, count := range stats.MethodCounts {
			fmt.Printf("  %s: %d\n", method, count)
		}
		fmt.Printf("\nTop paths:\n")
		for path, count := range stats.PathCounts {
			fmt.Printf("  %s: %d\n", path, count)
		}
	}
	fmt.Println()

	return nil
}
