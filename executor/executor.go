package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"code-execution-mcp/config"
)

// ExecutionResult contains the result of a Python script execution
type ExecutionResult struct {
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exit_code"`
}

// PythonExecutor handles Python script execution in isolated workspaces
type PythonExecutor struct {
	workspace      *WorkspaceManager
	defaultTimeout time.Duration
	generator      *ModuleGenerator
}

// NewPythonExecutor creates a new Python executor
func NewPythonExecutor(cfg *config.ExecutionConfig) (*PythonExecutor, error) {
	workspace, err := NewWorkspaceManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace manager: %w", err)
	}

	// Cleanup any existing sessions from previous runs
	if err := workspace.CleanupAll(); err != nil {
		return nil, fmt.Errorf("failed to cleanup old sessions: %w", err)
	}

	defaultTimeout := time.Duration(cfg.DefaultTimeout) * time.Second
	if defaultTimeout == 0 {
		defaultTimeout = 30 * time.Second
	}

	generator, err := NewModuleGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create module generator: %w", err)
	}

	return &PythonExecutor{
		workspace:      workspace,
		defaultTimeout: defaultTimeout,
		generator:      generator,
	}, nil
}

// Execute runs Python code with a specified timeout
func (pe *PythonExecutor) Execute(ctx context.Context, code string, timeoutSeconds int, servers map[string][]config.Tool, serverConfigs []config.MCPServerConfig) (*ExecutionResult, error) {
	// Create session workspace
	sessionPath, err := pe.workspace.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer pe.workspace.Cleanup(sessionPath)

	// Generate modules
	if err := pe.generator.Generate(sessionPath, servers, serverConfigs); err != nil {
		return nil, fmt.Errorf("failed to generate modules: %w", err)
	}

	// Write code to a temporary file
	scriptPath := filepath.Join(sessionPath, "script.py")
	if err := os.WriteFile(scriptPath, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	// Set timeout
	timeout := pe.defaultTimeout
	if timeoutSeconds > 0 {
		timeout = time.Duration(timeoutSeconds) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute Python script
	cmd := exec.CommandContext(ctx, "python3", "script.py")
	cmd.Dir = sessionPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	result := &ExecutionResult{
		Output:   stdout.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}

		// Include stderr in error if present
		if stderr.Len() > 0 {
			result.Error = stderr.String()
		} else {
			result.Error = err.Error()
		}
	}

	return result, nil
}
