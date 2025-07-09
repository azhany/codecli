package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/azhany/codecli/internal/config"
	"github.com/azhany/codecli/internal/tools"
)

// AddCommands adds all CLI commands to the root command
func AddCommands(rootCmd *cobra.Command) {
	// Config commands
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}
	rootCmd.AddCommand(configCmd)

	// Index command
	indexCmd := &cobra.Command{
		Use:   "index",
		Short: "Index codebase for semantic search",
		Run: func(cmd *cobra.Command, args []string) {
			if err := tools.IndexCodebase(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(indexCmd)

	// Search command
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search codebase",
		Run: func(cmd *cobra.Command, args []string) {
			if err := tools.SearchCodebase(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(searchCmd)

	// Chat command
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start interactive chat mode",
		Run: func(cmd *cobra.Command, args []string) {
			if err := tools.StartChat(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(chatCmd)
}
