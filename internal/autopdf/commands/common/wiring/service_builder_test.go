// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package wiring

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common/args"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/domain/options"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceBuilder_BuildDocumentService(t *testing.T) {
	builder := NewServiceBuilder()
	variables := config.NewVariables()
	require.NoError(t, variables.SetString("key", "value"))

	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Variables: *variables,
		Engine:    config.Engine("pdflatex"),
		Output:    config.Output("output"),
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"jpeg"},
		},
	}

	service := builder.BuildDocumentService(cfg)

	assert.NotNil(t, service)
	assert.NotNil(t, service.TemplateProcessor)
	assert.NotNil(t, service.LaTeXCompiler)
	assert.NotNil(t, service.Converter)
	assert.NotNil(t, service.Cleaner)
}

func TestServiceBuilder_BuildRequest(t *testing.T) {
	builder := NewServiceBuilder()
	variables := config.NewVariables()
	require.NoError(t, variables.SetString("key", "value"))

	cfg := &config.Config{
		Template:   config.Template("template.tex"),
		Variables:  *variables,
		Engine:     config.Engine("pdflatex"),
		Output:     config.Output("output"),
		Passes:     2,
		UseLatexmk: true,
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"jpeg"},
		},
	}
	args := &args.BuildArgs{
		TemplateFile: "template.tex",
		ConfigFile:   "config.yaml",
		Options: func() options.BuildOptions {
			opts := options.NewBuildOptions()
			opts.EnableClean(".")
			return opts
		}(),
	}

	req := builder.BuildRequest(args, cfg)

	assert.Equal(t, "template.tex", req.TemplatePath)
	assert.Equal(t, "config.yaml", req.ConfigPath)
	assert.NotNil(t, req.Variables)
	assert.Equal(t, map[string]string{"key": "value"}, req.Variables.Flatten())
	assert.Equal(t, "pdflatex", req.Engine)
	assert.Equal(t, "output", req.OutputPath)
	assert.Equal(t, ".", req.WorkingDir)
	assert.Equal(t, 2, req.Passes)
	assert.True(t, req.UseLatexmk)
	assert.True(t, req.DoClean)
	assert.True(t, req.Conversion.Enabled)
	assert.Equal(t, []string{"jpeg"}, req.Conversion.Formats)
}
