// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoRoot walks up from this test file until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir, "go.mod not found walking up from test dir")
		dir = parent
	}
}

// loggerKeyDecl is one `const <name> <type> = "logger"` declaration found in the tree.
type loggerKeyDecl struct {
	File string
	Name string
	Type string
}

// findLoggerContextKeys collects every constant whose value is the string
// "logger" and whose type is a named string type — the shape used for a
// context.Context key.
func findLoggerContextKeys(t *testing.T, root string) []loggerKeyDecl {
	t.Helper()
	var found []loggerKeyDecl

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == ".git" || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return nil // not our concern here; the compiler reports syntax errors
		}

		ast.Inspect(file, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST {
				return true
			}
			for _, spec := range decl.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok || value.Type == nil {
					continue
				}
				typeIdent, ok := value.Type.(*ast.Ident)
				if !ok {
					continue
				}
				for i, val := range value.Values {
					lit, ok := val.(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING || lit.Value != `"logger"` {
						continue
					}
					rel, _ := filepath.Rel(root, path)
					found = append(found, loggerKeyDecl{
						File: rel,
						Name: value.Names[i].Name,
						Type: typeIdent.Name,
					})
				}
			}
			return true
		})
		return nil
	})
	require.NoError(t, err)
	return found
}

// TestSingleLoggerContextKey guards the Single-Source rule for the logger
// context key.
//
// context.Value compares keys by dynamic TYPE as well as value, so two
// constants that both read "logger" but carry different named types never
// match each other. A second declaration therefore does not merely duplicate
// the first — it silently reads nothing, and every lookup through it falls
// back to a freshly built default logger, discarding whatever the caller
// configured. The failure is invisible at runtime: logging still happens,
// just at the wrong level and to the wrong sink.
func TestSingleLoggerContextKey(t *testing.T) {
	root := repoRoot(t)
	keys := findLoggerContextKeys(t, root)

	require.NotEmpty(t, keys, "expected to find the logger context key declaration")

	assert.Len(t, keys, 1,
		"exactly one logger context key must exist; found %d: %+v. "+
			"Readers must go through configs.GetLoggerFromContext rather than "+
			"declaring a package-private key of their own.", len(keys), keys)

	assert.Equal(t, "configs/defaults.go", filepath.ToSlash(keys[0].File),
		"the single logger context key belongs in configs")
}
