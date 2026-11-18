// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/pdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublication(t *testing.T) {
	buildErr := errors.New("xelatex exited 1")
	incomplete := &pdf.IncompleteError{Path: "out.pdf", Size: 9, Defect: pdf.DefectMissingTrailer}

	t.Run("a finished PDF crosses, with or without a build error", func(t *testing.T) {
		assert.NoError(t, publication(nil, nil))
		assert.NoError(t, publication(buildErr, nil))
	})
	t.Run("an unfinished PDF never crosses", func(t *testing.T) {
		assert.Same(t, incomplete, publication(nil, incomplete))
	})
	t.Run("when both failed, the build error leads and the defect stays matchable", func(t *testing.T) {
		err := publication(buildErr, incomplete)
		assert.ErrorIs(t, err, buildErr)
		var defect *pdf.IncompleteError
		require.ErrorAs(t, err, &defect)
		assert.Equal(t, pdf.DefectMissingTrailer, defect.Defect)
	})
}

// The production race: the caller's asset directory holds a symlink that
// another request removes while this one compiles. The workspace links the
// resolved target, so it keeps working.
func TestLinkWorkspaceAssetsSurvivesTheCallerRemovingItsLinks(t *testing.T) {
	root := t.TempDir()
	realAssets := filepath.Join(root, "template-assets")
	require.NoError(t, os.MkdirAll(realAssets, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(realAssets, "logo.png"), []byte("png"), 0644))

	callerDir := filepath.Join(root, "caller")
	require.NoError(t, os.MkdirAll(callerDir, 0755))
	require.NoError(t, os.Symlink(realAssets, filepath.Join(callerDir, "assets")))
	template := filepath.Join(callerDir, "letter.tex")
	require.NoError(t, os.WriteFile(template, []byte("tex"), 0644))

	workspace := filepath.Join(root, "workspace")
	require.NoError(t, os.MkdirAll(workspace, 0755))
	require.NoError(t, linkWorkspaceAssets(callerDir, workspace, template, "output.pdf"))

	require.NoError(t, os.Remove(filepath.Join(callerDir, "assets")))

	content, err := os.ReadFile(filepath.Join(workspace, "assets", "logo.png"))
	require.NoError(t, err)
	assert.Equal(t, "png", string(content))
	assert.NoFileExists(t, filepath.Join(workspace, "letter.tex"), "the template is never linked")
}
