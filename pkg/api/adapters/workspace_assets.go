// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"fmt"
	"os"
	"path/filepath"
)

// linkWorkspaceAssets is the Collect step of a request workspace: it links
// every entry of assetDir into workspace, except the template, its config and
// the staged output.
func linkWorkspaceAssets(assetDir, workspace, templatePath, stageName string) error {
	if assetDir == "" {
		return nil
	}
	assetDir, err := filepath.Abs(assetDir)
	if err != nil {
		return fmt.Errorf("resolve asset directory: %w", err)
	}
	workspace, err = filepath.Abs(workspace)
	if err != nil {
		return fmt.Errorf("resolve workspace: %w", err)
	}
	if assetDir == workspace {
		return nil
	}

	entries, err := os.ReadDir(assetDir)
	if err != nil {
		return fmt.Errorf("read asset directory: %w", err)
	}
	templatePath, err = filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("resolve template path: %w", err)
	}
	for _, entry := range entries {
		if entry.Name() == "config.yaml" || entry.Name() == stageName {
			continue
		}
		source := filepath.Join(assetDir, entry.Name())
		if source == templatePath {
			continue
		}
		// Link what the entry resolves to now, not the entry itself. The
		// caller owns assetDir and may replace or remove its links while this
		// request still compiles; the workspace must not follow them.
		if resolved, err := filepath.EvalSymlinks(source); err == nil {
			source = resolved
		}
		target := filepath.Join(workspace, entry.Name())
		if err := os.Symlink(source, target); err != nil {
			return fmt.Errorf("link workspace asset %q: %w", entry.Name(), err)
		}
	}
	return nil
}
