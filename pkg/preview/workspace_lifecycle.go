package preview

import (
	"errors"
	"fmt"
	"os"
)

var ErrWorkspaceRequired = errors.New("preview workspace factory returned no workspace")

// Workspace is a private session directory lease. Close owns its cleanup.
type Workspace interface {
	Path() string
	Close() error
}

// WorkspaceFactory is the injectable lifecycle port for preview storage.
type WorkspaceFactory interface {
	Create() (Workspace, error)
}

type localWorkspaceFactory struct {
	root string
	keep bool
}

type localWorkspace struct {
	path string
	keep bool
}

func (factory localWorkspaceFactory) Create() (Workspace, error) {
	directory, err := os.MkdirTemp(factory.root, "autopdf-preview-")
	if err != nil {
		return nil, fmt.Errorf("create preview workspace: %w", err)
	}
	return &localWorkspace{path: directory, keep: factory.keep}, nil
}

func (workspace *localWorkspace) Path() string { return workspace.path }

func (workspace *localWorkspace) Close() error {
	if workspace.keep {
		return nil
	}
	return os.RemoveAll(workspace.path)
}
