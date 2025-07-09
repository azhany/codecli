package config

import (
	"fmt"
	
	"github.com/spf13/viper"
)

// Config holds the application configuration
var Config = struct {
	Ollama struct {
		URL          string `mapstructure:"url"`
		ChatModel    string `mapstructure:"chat_model"`
		CodeModel    string `mapstructure:"code_model"`
		EmbeddingModel string `mapstructure:"embedding_model"`
		Timeout      string `mapstructure:"timeout"`
	}
	NGT struct {
		IndexPath  string `mapstructure:"index_path"`
		Dimension  int    `mapstructure:"dimension"`
		EdgeSize   int    `mapstructure:"edge_size"`
		BatchSize  int    `mapstructure:"batch_size"`
	}
	Workspace struct {
		Root            string   `mapstructure:"root"`
		ExcludePatterns []string `mapstructure:"exclude_patterns"`
		IncludeExtensions []string `mapstructure:"include_extensions"`
	}
	Logging struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
		Output string `mapstructure:"output"`
	}
}{
	Ollama: struct {
		URL:          "http://localhost:11434",
		ChatModel:    "llama2",
		CodeModel:    "codellama",
		EmbeddingModel: "nomic-embed-text",
		Timeout:      "30s",
	},
	NGT: struct {
		IndexPath:  ".codecli/index",
		Dimension:  768,
		EdgeSize:   10,
		BatchSize:  100,
	},
	Workspace: struct {
		Root:            ".",
		ExcludePatterns: []string{"*.git*", "node_modules", "*.log", "*.tmp"},
		IncludeExtensions: []string{".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", "php"},
	},
	Logging: struct {
		Level:  "info",
		Format: "json",
		Output: "stdout",
	},
}

// LoadConfig loads the configuration from file
func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/codecli")

	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Unmarshal(&Config); err != nil {
			return fmt.Errorf("error unmarshaling config: %v", err)
		}
	}

	return nil
}
