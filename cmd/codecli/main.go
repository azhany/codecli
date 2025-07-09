package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/azhany/codecli/internal/cli"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "codecli",
		Short: "AI-Assisted Codebase Analysis Tool",
		Long: `CodeCLI is a terminal-based tool that enables local codebase analysis,
		LLM-assisted reasoning, and integrates with NGT for vector storage.`,
	}

	// Add subcommands
	cli.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
