package tools

import (
	"fmt"
	"os/exec"
)

// Command handles shell command execution
type Command struct {
	*Base
}

func NewCommand() *Command {
	return &Command{
		Base: NewBase("command", "Executes shell commands"),
	}
}

func (t *Command) RunCommand(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v", err)
	}
	return string(output), nil
}

func (t *Command) Execute(args map[string]interface{}) (interface{}, error) {
	cmd, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command argument is required")
	}

	workdir, _ := args["workdir"].(string)
	if workdir == "" {
		workdir = "."
	}

	command := exec.Command("sh", "-c", cmd)
	command.Dir = workdir
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %v", err)
	}

	return string(output), nil
}
