package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code-execution-mcp/config"
)

// WorkspaceManager handles session-based workspace directories
type WorkspaceManager struct {
	basePath string
}

// NewWorkspaceManager creates a new workspace manager
func NewWorkspaceManager(cfg *config.ExecutionConfig) (*WorkspaceManager, error) {
	basePath := cfg.WorkspacePath
	if basePath == "" {
		basePath = "./workspace"
	}

	// Create base workspace directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	return &WorkspaceManager{
		basePath: basePath,
	}, nil
}

// CreateSession creates a new session workspace directory
func (wm *WorkspaceManager) CreateSession() (string, error) {
	// Use timestamp for unique session ID
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	sessionPath := filepath.Join(wm.basePath, sessionID)

	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	return sessionPath, nil
}

// Cleanup removes a session workspace directory
func (wm *WorkspaceManager) Cleanup(sessionPath string) error {
	// Validate that the path is within our workspace (prevent directory traversal)
	absSession, err := filepath.Abs(sessionPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	absBase, err := filepath.Abs(wm.basePath)
	if err != nil {
		return fmt.Errorf("failed to get base path: %w", err)
	}

	if !strings.HasPrefix(absSession, absBase) {
		return fmt.Errorf("invalid session path: outside workspace")
	}

	return os.RemoveAll(sessionPath)
}

// CleanupAll removes all session directories (useful for startup cleanup)
func (wm *WorkspaceManager) CleanupAll() error {
	entries, err := os.ReadDir(wm.basePath)
	if err != nil {
		return fmt.Errorf("failed to read workspace directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "session_") {
			sessionPath := filepath.Join(wm.basePath, entry.Name())
			if err := os.RemoveAll(sessionPath); err != nil {
				return fmt.Errorf("failed to cleanup session %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}
