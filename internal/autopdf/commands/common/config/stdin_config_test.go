// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/configs"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common/args"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_StdinReadsJSONFromInjectedReader(t *testing.T) {
	body := `{"template":"doc.tex","output":"out\/doc.pdf","engine":"xelatex","passes":2,
"variables":{"title":"A \"quoted\" title","n":12,"items":["x","y"],"meta":{"author":"Ann"}}}`
	resolver := NewConfigResolver().WithStdin(strings.NewReader(body))

	cfg, err := resolver.LoadConfig(args.StdinConfig)
	require.NoError(t, err)

	assert.Equal(t, "doc.tex", cfg.Template.String())
	assert.Equal(t, "out/doc.pdf", cfg.Output.String())
	assert.Equal(t, "xelatex", cfg.Engine.String())
	assert.Equal(t, 2, cfg.Passes)
	flat := cfg.Variables.Flatten()
	assert.Equal(t, `A "quoted" title`, flat["title"])
	assert.Equal(t, "12", flat["n"])
	assert.Equal(t, "x", flat["items[0]"])
	assert.Equal(t, "y", flat["items[1]"])
	assert.Equal(t, "Ann", flat["meta.author"])
}

func TestLoadConfig_StdinReadsYAMLFromInjectedReader(t *testing.T) {
	body := "output: out.pdf\nvariables:\n  title: Hello\n  items:\n    - a\n    - b\n"
	resolver := NewConfigResolver().WithStdin(strings.NewReader(body))

	cfg, err := resolver.LoadConfig(args.StdinConfig)
	require.NoError(t, err)

	assert.Equal(t, "out.pdf", cfg.Output.String())
	assert.Equal(t, "", cfg.Engine.String())
	flat := cfg.Variables.Flatten()
	assert.Equal(t, "Hello", flat["title"])
	assert.Equal(t, "b", flat["items[1]"])
}

func TestLoadConfig_StdinOmittedEngineStaysEmptyForJSON(t *testing.T) {
	resolver := NewConfigResolver().WithStdin(strings.NewReader(`{"variables":{"a":"b"}}`))

	cfg, err := resolver.LoadConfig(args.StdinConfig)
	require.NoError(t, err)
	assert.Equal(t, "", cfg.Engine.String())
}

func TestLoadConfig_StdinRefusesMalformedInput(t *testing.T) {
	for name, body := range map[string]string{
		"broken json": `{"variables": {"a": }`,
		"broken yaml": "variables: [unclosed\n",
	} {
		t.Run(name, func(t *testing.T) {
			resolver := NewConfigResolver().WithStdin(strings.NewReader(body))
			_, err := resolver.LoadConfig(args.StdinConfig)
			assert.True(t, errors.Is(err, configs.ParseError), "got %v", err)
		})
	}
}

func TestLoadConfig_StdinWithoutReaderIsAReadError(t *testing.T) {
	_, err := (&ConfigResolver{}).LoadConfig(args.StdinConfig)
	assert.True(t, errors.Is(err, configs.ReadError), "got %v", err)
}

func TestParseConfigData_JSONEscapesYAMLCannotRead(t *testing.T) {
	cfg, err := ParseConfigData([]byte(`{"variables":{"path":"a\/b","e":"caf\u00e9"}}`))
	require.NoError(t, err)
	flat := cfg.Variables.Flatten()
	assert.Equal(t, "a/b", flat["path"])
	assert.Equal(t, "café", flat["e"])
}

func TestParseConfigData_YAMLFlowMappingIsStillYAML(t *testing.T) {
	cfg, err := ParseConfigData([]byte("{variables: {title: Hi}}"))
	require.NoError(t, err)
	assert.Equal(t, "Hi", cfg.Variables.Flatten()["title"])
}

func TestResolveTemplatePath_StdinConfigResolvesAgainstTemplateDir(t *testing.T) {
	resolver := NewConfigResolver()
	templateFile := filepath.Join("/work", "docs", "main.tex")

	cfg := &config.Config{Template: config.Template("parts/body.tex")}
	require.NoError(t, resolver.ResolveTemplatePath(cfg, templateFile, args.StdinConfig))
	assert.Equal(t, "/work/docs/parts/body.tex", cfg.Template.String())

	cfg = &config.Config{}
	require.NoError(t, resolver.ResolveTemplatePath(cfg, templateFile, args.StdinConfig))
	assert.Equal(t, templateFile, cfg.Template.String())
}
