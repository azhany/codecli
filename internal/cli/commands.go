package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/azhany/codecli/internal/tools"
	"github.com/azhany/codecli/internal/types"
	"github.com/azhany/codecli/internal/vector"
	"github.com/spf13/cobra"
)

// AddCommands adds all CLI commands to the root command
func AddCommands(rootCmd *cobra.Command) {
	// Initialize core components
	vectorStore, err := vector.NewStore()
	if err != nil {
		fmt.Printf("Error initializing vector store: %v\n", err)
		os.Exit(1)
	}

	toolManager := tools.NewManager()

	// Register tools
	searchTool := tools.NewSearch(vectorStore)
	toolManager.RegisterTool(searchTool)

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
			tool, err := toolManager.GetTool("search")
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			_, err = tool.Execute(map[string]interface{}{
				"operation": "index",
			})
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			fmt.Println("Successfully indexed codebase")
		},
	}
	rootCmd.AddCommand(indexCmd)

	// Search command
	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search codebase",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tool, err := toolManager.GetTool("search")
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			query := strings.Join(args, " ")
			results, err := tool.Execute(map[string]interface{}{
				"operation": "search",
				"query":     query,
			})
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			searchResults, ok := results.([]types.SearchResult)
			if !ok {
				fmt.Println("Error: Invalid search results")
				os.Exit(1)
			}

			if len(searchResults) == 0 {
				fmt.Println("No results found")
				return
			}

			for _, result := range searchResults {
				fmt.Printf("%s:%d: %s\n", result.Path, result.Line, result.Content)
			}
		},
	}
	rootCmd.AddCommand(searchCmd)

	// Chat command (placeholder - will be implemented with chat tool)
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start interactive chat mode",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Chat functionality will be implemented soon")
		},
	}
	rootCmd.AddCommand(chatCmd)
}
