package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	configFile string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "gpp-test",
	Short: "GPP Local Testing Tool",
	Long: `GPP Local Testing Tool - Test GitHub Pipeline Platform changes locally using act.

This tool allows you to test changes to jenkins-library, piper-library, and
github-actions locally without pushing to GitHub. It provides mock external
services and uses your local Piper binary.`,
	Version: "1.0.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "Config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func printHeader(title string) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  %s\n", title)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}
