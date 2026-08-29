package preview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BuddhiLW/AutoPDF/pkg/render/latex"
)

func (session *Session) materialize(ctx context.Context, projection latex.Projection) ([]string, error) {
	desired := make(map[string]string, len(projection.Files)+len(projection.Assets))
	dirty := make([]string, 0)
	for _, file := range projection.Files {
		hash := contentHash(file.Content)
		desired[file.Path] = hash
		if session.written[file.Path] == hash {
			continue
		}
		if err := writeWorkspaceFile(session.workspace, file.Path, file.Content); err != nil {
			return nil, err
		}
		dirty = append(dirty, file.Path)
	}
	for _, asset := range projection.Assets {
		if session.resolver == nil {
			return nil, ErrResolverRequired
		}
		if _, unchanged := session.written[asset.Path]; unchanged {
			desired[asset.Path] = session.written[asset.Path]
			continue
		}
		data, err := session.resolver.Resolve(ctx, asset)
		if err != nil {
			return nil, fmt.Errorf("resolve asset %s: %w", asset.AssetID, err)
		}
		hash := contentHash(data)
		desired[asset.Path] = hash
		if err := writeWorkspaceFile(session.workspace, asset.Path, data); err != nil {
			return nil, err
		}
		dirty = append(dirty, asset.Path)
	}
	for relative := range session.written {
		if _, exists := desired[relative]; exists {
			continue
		}
		absolute, err := safeWorkspacePath(session.workspace, relative)
		if err != nil {
			return nil, err
		}
		if err := os.Remove(absolute); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove stale preview file %s: %w", relative, err)
		}
		dirty = append(dirty, relative)
	}
	session.written = desired
	sort.Strings(dirty)
	return dirty, nil
}

func writeWorkspaceFile(root, relative string, data []byte) error {
	absolute, err := safeWorkspacePath(root, relative)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
		return fmt.Errorf("create preview directory: %w", err)
	}
	temporary := absolute + ".autopdf-new"
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return fmt.Errorf("write preview file %s: %w", relative, err)
	}
	if err := os.Rename(temporary, absolute); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("publish preview file %s: %w", relative, err)
	}
	return nil
}

func safeWorkspacePath(root, relative string) (string, error) {
	if relative == "" || filepath.IsAbs(relative) || strings.ContainsRune(relative, '\x00') {
		return "", fmt.Errorf("%w: %q", ErrUnsafePath, relative)
	}
	clean := filepath.Clean(filepath.FromSlash(relative))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %q", ErrUnsafePath, relative)
	}
	return filepath.Join(root, clean), nil
}

func contentHash(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
