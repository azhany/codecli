# CodeCLI - AI-Assisted Codebase Analysis Tool

A terminal-based Golang CLI tool for software engineers that enables local codebase analysis, LLM-assisted reasoning using Ollama, and integrates with NGT for vector storage.

## Features

- **Codebase Analysis**: Parse and analyze code structures, extract definitions, and navigate codebases
- **LLM Integration**: Leverage Ollama for intelligent code understanding and generation
- **Semantic Search**: Use vector embeddings for semantic code search and similarity matching
- **Tool Calling**: LLM can invoke tools to read files, execute commands, and analyze code
- **Conversational Interface**: Interactive chat mode with memory and context awareness
- **Code Completion**: AI-powered code completion and suggestions
- **Multi-language Support**: Extensible architecture for different programming languages

## Prerequisites

### System Requirements
- Go 1.21 or higher
- Ollama installed and running locally
- NGT (Neighborhood Graph and Tree) library

### Dependencies Installation

#### 1. Install Ollama
```bash
# On macOS
brew install ollama

# On Linux
curl -fsSL https://ollama.ai/install.sh | sh

# On Windows
# Download from https://ollama.ai/download
```

#### 2. Install Required Models
```bash
# Install a general chat model
ollama pull llama2

# Install a code-specialized model
ollama pull codellama

# Install an embedding model
ollama pull nomic-embed-text
```

#### 3. Install NGT
```bash
# On Ubuntu/Debian
sudo apt-get install libngt-dev

# On macOS
brew install ngt

# On Windows
# Follow NGT installation guide for Windows
```

## Installation

### From Source
```bash
git clone <repository-url>
cd codecli
go mod tidy
go build -o codecli ./cmd/codecli
```

### Using Go Install
```bash
go install github.com/your-org/codecli/cmd/codecli@latest
```

## Configuration

Create a configuration file `config.yaml` in your home directory or project root:

```yaml
# Ollama Configuration
ollama:
  url: "http://localhost:11434"
  chat_model: "llama2"
  code_model: "codellama"
  embedding_model: "nomic-embed-text"
  timeout: "30s"

# NGT Configuration
ngt:
  index_path: ".codecli/index"
  dimension: 768
  edge_size: 10
  batch_size: 100

# Workspace Configuration
workspace:
  root: "."
  exclude_patterns:
    - "*.git*"
    - "node_modules"
    - "*.log"
    - "*.tmp"
  include_extensions:
    - ".go"
    - ".py"
    - ".js"
    - ".ts"
    - ".java"
    - ".cpp"
    - ".c"
    - ".h"

# Logging Configuration
logging:
  level: "info"
  format: "json"
  output: "stdout"
```

## Usage

### Basic Commands

#### Initialize and Index Codebase
```bash
# Index current directory
codecli index

# Index specific directory
codecli index --path /path/to/project

# Index with custom config
codecli index --config custom-config.yaml
```

#### File Operations
```bash
# Read a file
codecli read --file src/main.go

# Write to a file
codecli write --file output.txt --content "Hello, World!"

# List files in workspace
codecli list --type files

# List code definitions
codecli list --type definitions --file src/main.go
```

#### Search Operations
```bash
# Keyword search
codecli search --query "function main" --type keyword

# Semantic search
codecli search --query "database connection logic" --type semantic

# Combined search
codecli search --query "error handling" --type both
```

#### Code Completion
```bash
# Complete code at cursor position
codecli complete --file src/main.go --line 42 --column 15

# Complete with context
codecli complete --file src/main.go --context 10
```

#### Interactive Chat Mode
```bash
# Start interactive session
codecli chat

# Chat with specific context
codecli chat --context-files src/main.go,src/utils.go

# Chat with memory from previous sessions
codecli chat --use-memory
```

#### Execute Commands
```bash
# Run shell command
codecli run --command "go test ./..."

# Run with output capture
codecli run --command "git status" --capture
```

### Advanced Usage

#### Tool Calling in Chat Mode
In chat mode, the LLM can automatically invoke tools based on your queries:

```
User: "Show me the main function in src/main.go"
Assistant: I'll read that file for you.
[Tool: read_file(src/main.go)]
[Result: file contents...]
Here's the main function from src/main.go: ...

User: "What tests are available?"
Assistant: Let me search for test files.
[Tool: search_files(pattern="*_test.go")]
[Result: list of test files...]
I found the following test files: ...
```

#### Batch Processing
```bash
# Process multiple files
codecli batch --files "src/*.go" --operation analyze

# Batch index with concurrency
codecli index --workers 8 --batch-size 50
```

#### Configuration Management
```bash
# Validate configuration
codecli config validate

# Show current configuration
codecli config show

# Set configuration values
codecli config set ollama.chat_model llama2:13b
```

## API Reference

### Core Tools Available to LLM

1. **execute_command(command: string)**: Execute shell commands
2. **read_file(path: string)**: Read file contents
3. **write_to_file(path: string, content: string, append: bool)**: Write to files
4. **list_files(root: string, pattern: string)**: List files recursively
5. **list_code_definition_names(file: string, language: string)**: Extract code definitions
6. **search_files(query: string, type: string, limit: int)**: Search codebase
7. **ask_followup_question(question: string, context: string)**: Handle conversational queries

### Configuration Options

#### Ollama Settings
- `ollama.url`: Ollama server URL (default: http://localhost:11434)
- `ollama.chat_model`: Model for chat interactions
- `ollama.code_model`: Model for code completion
- `ollama.embedding_model`: Model for embeddings
- `ollama.timeout`: Request timeout

#### NGT Settings
- `ngt.index_path`: Path to store vector index
- `ngt.dimension`: Vector dimension (must match embedding model)
- `ngt.edge_size`: NGT edge size parameter
- `ngt.batch_size`: Batch size for indexing

#### Workspace Settings
- `workspace.root`: Root directory for analysis
- `workspace.exclude_patterns`: Patterns to exclude
- `workspace.include_extensions`: File extensions to include

## Architecture

### Project Structure
```
codecli/
├── cmd/codecli/           # CLI entry point
├── internal/
│   ├── cli/              # CLI commands and handlers
│   ├── config/           # Configuration management
│   ├── llm/              # Ollama integration
│   ├── vector/           # NGT vector storage
│   ├── tools/            # Tool implementations
│   ├── parser/           # Code parsing utilities
│   └── logger/           # Logging utilities
├── pkg/                  # Public packages
├── configs/              # Configuration files
├── docs/                 # Documentation
└── tests/                # Test files
```

### Key Components

1. **CLI Layer**: Cobra-based command interface
2. **Configuration**: Viper-based config management
3. **LLM Client**: Ollama API integration
4. **Vector Store**: NGT-based semantic search
5. **Tool System**: Extensible tool calling framework
6. **Parser**: Multi-language code analysis
7. **Logger**: Structured logging with context

## Development

### Building from Source
```bash
# Clone repository
git clone <repository-url>
cd codecli

# Install dependencies
go mod tidy

# Build
go build -o codecli ./cmd/codecli

# Run tests
go test ./...

# Run with race detection
go test -race ./...
```

### Adding New Tools
1. Implement tool interface in `internal/tools/`
2. Register tool in tool registry
3. Update LLM system prompt
4. Add tests and documentation

### Adding Language Support
1. Implement parser in `internal/parser/`
2. Add language-specific patterns
3. Update configuration schema
4. Add tests for new language

## Troubleshooting

### Common Issues

#### Ollama Connection Issues
```bash
# Check if Ollama is running
ollama list

# Start Ollama service
ollama serve

# Test connection
curl http://localhost:11434/api/tags
```

#### NGT Index Issues
```bash
# Remove corrupted index
rm -rf .codecli/index

# Rebuild index
codecli index --rebuild
```

#### Performance Issues
```bash
# Reduce batch size
codecli config set ngt.batch_size 25

# Increase workers
codecli index --workers 4

# Exclude large directories
codecli config set workspace.exclude_patterns "node_modules,*.git*,build"
```

### Debug Mode
```bash
# Enable debug logging
codecli --log-level debug chat

# Verbose output
codecli --verbose index
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- GitHub Issues: Report bugs and feature requests
- Documentation: Check docs/ directory
- Examples: See examples/ directory