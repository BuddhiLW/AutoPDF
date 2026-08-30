package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const renderTargetRoot = "pkg/render/"

// TestCoreDoesNotImportARenderTarget asserts that pkg/api names no concrete
// rendering target except the default one, so adding a target is a new package
// rather than an edit to the core.
//
// pkg/render/latex is exempt: it supplies the shared projection value type and
// the default projector.
func TestCoreDoesNotImportARenderTarget(t *testing.T) {
	root := repoRoot(t)
	offenders := map[string][]string{}

	err := filepath.Walk(filepath.Join(root, "pkg", "api"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				continue
			}
			index := strings.Index(imported, renderTargetRoot)
			if index < 0 {
				continue
			}
			target := imported[index+len(renderTargetRoot):]
			if target == "latex" {
				continue
			}
			relative, _ := filepath.Rel(root, path)
			offenders[target] = append(offenders[target], filepath.ToSlash(relative))
		}
		return nil
	})
	require.NoError(t, err)

	assert.Empty(t, offenders,
		"pkg/api must not import a concrete render target; a target supplies its own "+
			"catalog and an api.ManifestProjector, and the engine takes both through "+
			"interfaces. Naming one here closes an open set: the next target needs an "+
			"edit to the core. Offenders: %+v", offenders)
}
