// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package fixturegate_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/pdf"
)

const configName = "config.yaml"

var templateEngines = map[string]string{
	".tex": "pdflatex",
	".typ": "typst",
}

type fixtureConfig struct {
	Template string `yaml:"template"`
	Output   string `yaml:"output"`
	Engine   string `yaml:"engine"`
}

type buildUnit struct {
	fixture  string
	dir      string
	template string
}

func discoverUnits(t *testing.T) []buildUnit {
	t.Helper()
	configs, err := filepath.Glob(filepath.Join("..", "*", configName))
	if err != nil {
		t.Fatalf("glob fixtures: %v", err)
	}
	sort.Strings(configs)
	var units []buildUnit
	for _, cfgPath := range configs {
		dir := filepath.Dir(cfgPath)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		found := 0
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if _, ok := templateEngines[filepath.Ext(entry.Name())]; !ok {
				continue
			}
			units = append(units, buildUnit{fixture: filepath.Base(dir), dir: dir, template: entry.Name()})
			found++
		}
		if found == 0 {
			t.Errorf("fixture %s has %s but no template file", dir, configName)
		}
	}
	return units
}

func buildBinary(t *testing.T, ctx context.Context) string {
	t.Helper()
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not found on PATH: cannot build the autopdf binary")
	}
	binary := filepath.Join(t.TempDir(), "autopdf")
	build := exec.CommandContext(ctx, goTool, "build", "-o", binary, "./cmd/autopdf")
	build.Dir = filepath.Join("..", "..")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/autopdf: %v\n%s", err, out)
	}
	return binary
}

func copyFixture(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o644)
	})
	if err != nil {
		t.Fatalf("copy fixture %s: %v", src, err)
	}
}

func pointConfigAt(t *testing.T, cfgPath, template string) fixtureConfig {
	t.Helper()
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read %s: %v", cfgPath, err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", cfgPath, err)
	}
	doc["template"] = template
	rewritten, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal %s: %v", cfgPath, err)
	}
	if err := os.WriteFile(cfgPath, rewritten, 0o644); err != nil {
		t.Fatalf("write %s: %v", cfgPath, err)
	}
	var cfg fixtureConfig
	if err := yaml.Unmarshal(rewritten, &cfg); err != nil {
		t.Fatalf("reparse %s: %v", cfgPath, err)
	}
	return cfg
}

func expectedPDF(workDir string, cfg fixtureConfig) string {
	output := cfg.Output
	if output == "" {
		output = strings.TrimSuffix(cfg.Template, filepath.Ext(cfg.Template))
	}
	if !strings.HasSuffix(output, ".pdf") {
		output += ".pdf"
	}
	if filepath.IsAbs(output) {
		return output
	}
	return filepath.Join(workDir, output)
}

// TestEveryFixtureBuilds builds every template of every test/*/ fixture that
// carries a config.yaml with the real `autopdf build`, and requires a complete
// PDF. A unit is skipped only when its engine binary is not on PATH.
func TestEveryFixtureBuilds(t *testing.T) {
	units := discoverUnits(t)
	if len(units) == 0 {
		t.Fatal("no fixtures discovered under test/*/" + configName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	t.Cleanup(cancel)
	binary := buildBinary(t, ctx)

	for _, unit := range units {
		unit := unit
		t.Run(unit.fixture+"/"+unit.template, func(t *testing.T) {
			t.Parallel()
			workDir := t.TempDir()
			copyFixture(t, unit.dir, workDir)
			cfg := pointConfigAt(t, filepath.Join(workDir, configName), unit.template)

			engine := cfg.Engine
			if engine == "" {
				engine = templateEngines[filepath.Ext(unit.template)]
			}
			if _, err := exec.LookPath(engine); err != nil {
				t.Skipf("engine %s not found on PATH", engine)
			}

			pdfPath := expectedPDF(workDir, cfg)
			run := exec.CommandContext(ctx, binary, "build", unit.template, configName)
			run.Dir = workDir
			out, err := run.CombinedOutput()
			if err != nil {
				t.Fatalf("autopdf build %s %s: %v\n%s", unit.template, configName, err, out)
			}
			content, err := os.ReadFile(pdfPath)
			if err != nil {
				t.Fatalf("expected PDF %s: %v\n%s", pdfPath, err, out)
			}
			if err := pdf.Check(pdfPath, content); err != nil {
				t.Fatalf("built PDF is not complete: %v", err)
			}
		})
	}
}
