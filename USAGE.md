# CodeCLI Usage Examples

This document provides comprehensive examples of how to use the CodeCLI tool.

## Prerequisites

1. **Install Ollama**:
   ```bash
   curl -fsSL https://ollama.ai/install.sh | sh
   ```

2. **Start Ollama service**:
   ```bash
   ollama serve
   ```

3. **Pull required models**:
   ```bash
   ollama pull llama2
   ollama pull codellama
   ollama pull nomic-embed-text
   ```

4. **Build CodeCLI**:
   ```bash
   make build
   # or
   go build -o codecli .
   ```

## Basic Commands

### 1. File Operations

#### Read a file
```bash
./codecli read --file example.go
```

#### Write to a file
```bash
./codecli write --file test.txt --content "Hello, World!"
```

#### Append to a file
```bash
./codecli write --file test.txt --content "\nNew line" --append
```

### 2. List Operations

#### List all files in workspace
```bash
./codecli list --type files
```

#### List code definitions in a file
```bash
./codecli list --type definitions --file example.go
```

### 3. Search Operations

#### Keyword search
```bash
./codecli search --query "User" --type keyword
./codecli search --query "database connection" --type keyword
```

#### Semantic search (requires indexing first)
```bash
./codecli search --query "user management functionality" --type semantic
```

#### Combined search
```bash
./codecli search --query "error handling" --type both
```

### 4. Execute Commands

#### Run shell commands
```bash
./codecli run --command "ls -la"
./codecli run --command "go test ./..."
./codecli run --command "git status"
```

## Advanced Features

### 1. Codebase Indexing

#### Index current directory
```bash
./codecli index
```

#### Index specific directory
```bash
./codecli index --path /path/to/project
```

#### Rebuild index
```bash
./codecli index --rebuild
```

#### Index with custom workers
```bash
./codecli index --workers 8 --batch-size 50
```

### 2. Code Completion

#### Basic completion
```bash
./codecli complete --file example.go --line 50
```

#### Completion with context
```bash
./codecli complete --file example.go --line 50 --context 10
```

#### Completion with cursor position
```bash
./codecli complete --file example.go --line 50 --column 15
```

### 3. Interactive Chat Mode

#### Start basic chat
```bash
./codecli chat
```

#### Chat with context files
```bash
./codecli chat --context-files example.go,main.go
```

#### Chat with memory
```bash
./codecli chat --use-memory
```

## Chat Mode Examples

Once in chat mode, you can ask various questions:

### Code Analysis Questions
```
> What functions are defined in example.go?
> How does the user creation work?
> Show me all the error handling in the codebase
> What are the main data structures used?
```

### File Operations via Chat
```
> Read the main.go file
> List all Go files in the project
> Search for database-related code
> Show me the configuration structure
```

### Code Generation
```
> Write a function to validate email addresses
> Create a test for the UserService
> Generate documentation for the User struct
> Add error handling to the CreateUser function
```

## Tool Calling Examples

The LLM can automatically call tools based on your requests:

### Example 1: File Analysis
```
User: "What's in the example.go file?"
Assistant: I'll read that file for you.
[Tool: read_file({"path": "example.go"})]
[Result: file contents...]
The example.go file contains a user management service with the following components:
- User struct with ID, Name, Email, and Created fields
- UserService for managing users
- HTTP handlers for REST API
- Database connection helpers
- Configuration management
```

### Example 2: Code Search
```
User: "Find all functions that handle errors"
Assistant: I'll search for error handling patterns in your codebase.
[Tool: search_files({"query": "error", "type": "keyword"})]
[Result: list of files...]
I found error handling in the following locations:
- example.go: UpdateUser and DeleteUser functions
- main.go: Configuration loading and database connection
```

### Example 3: Code Execution
```
User: "Run the tests"
Assistant: I'll run the tests for you.
[Tool: execute_command({"command": "go test ./..."})]
[Result: test output...]
The tests completed successfully with the following results: ...
```

## Configuration Examples

### Custom Configuration File
Create a `custom-config.yaml`:
```yaml
ollama:
  url: "http://localhost:11434"
  chat_model: "llama2:13b"
  code_model: "codellama:7b"
  embedding_model: "nomic-embed-text"

workspace:
  root: "/path/to/your/project"
  exclude_patterns:
    - "*.git*"
    - "node_modules"
    - "vendor"
    - "*.log"
  include_extensions:
    - ".go"
    - ".py"
    - ".js"
    - ".ts"

logging:
  level: "debug"
  format: "json"
```

Use custom config:
```bash
./codecli --config custom-config.yaml chat
```

## Workflow Examples

### 1. New Project Analysis
```bash
# 1. Index the codebase
./codecli index --path /path/to/new/project

# 2. Get overview
./codecli chat
> "Give me an overview of this codebase"

# 3. Find entry points
> "What are the main functions or entry points?"

# 4. Understand architecture
> "How is the code organized? What are the main components?"
```

### 2. Bug Investigation
```bash
# 1. Search for error-related code
./codecli search --query "error" --type both

# 2. Analyze specific files
./codecli read --file problematic_file.go

# 3. Get AI assistance
./codecli chat --context-files problematic_file.go
> "I'm getting an error in this file. Can you help me understand what might be wrong?"
```

### 3. Code Review Assistance
```bash
# 1. List recent changes
./codecli run --command "git diff --name-only HEAD~1"

# 2. Review each file
./codecli chat
> "Review the changes in file.go and suggest improvements"

# 3. Check for patterns
> "Are there any code quality issues or patterns I should be aware of?"
```

### 4. Documentation Generation
```bash
# 1. Analyze code structure
./codecli list --type definitions --file main.go

# 2. Generate documentation
./codecli chat
> "Generate documentation for the main functions in this file"

# 3. Create README sections
> "Create a usage section for the README based on the available functions"
```

## Troubleshooting

### Common Issues

#### Ollama Connection Error
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Start Ollama if not running
ollama serve
```

#### Model Not Found
```bash
# Pull required models
ollama pull llama2
ollama pull codellama
ollama pull nomic-embed-text
```

#### Index Issues
```bash
# Clear and rebuild index
./codecli index --rebuild

# Check index status
./codecli search --query "test" --type semantic
```

#### Performance Issues
```bash
# Reduce batch size
./codecli index --batch-size 25

# Use fewer workers
./codecli index --workers 2

# Enable debug logging
./codecli --log-level debug index
```

## Tips and Best Practices

1. **Index First**: Always index your codebase before using semantic search
2. **Use Context**: Provide context files when starting chat sessions
3. **Specific Queries**: Be specific in your search queries for better results
4. **Combine Tools**: Use different commands together for comprehensive analysis
5. **Regular Updates**: Re-index when your codebase changes significantly
6. **Configuration**: Customize configuration for your specific needs
7. **Memory Management**: Clear chat history periodically for better performance

## Integration with IDEs

### VS Code Integration
You can integrate CodeCLI with VS Code by creating tasks:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "CodeCLI: Index Codebase",
            "type": "shell",
            "command": "./codecli",
            "args": ["index"],
            "group": "build"
        },
        {
            "label": "CodeCLI: Search Code",
            "type": "shell",
            "command": "./codecli",
            "args": ["search", "--query", "${input:searchQuery}", "--type", "both"],
            "group": "build"
        }
    ],
    "inputs": [
        {
            "id": "searchQuery",
            "description": "Enter search query",
            "default": "",
            "type": "promptString"
        }
    ]
}
```

This comprehensive guide should help you get started with CodeCLI and explore its full potential for codebase analysis and AI-assisted development.