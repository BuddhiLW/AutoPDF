// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package wiring

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/args"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/options"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestServiceBuilder_BuildDocumentService(t *testing.T) {
	builder := NewServiceBuilder()
	variables := config.NewVariables()
	variables.SetString("key", "value")

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
	variables.SetString("key", "value")

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
	assert.Equal(t, map[string]string{"key": "value"}, req.Variables)
	assert.Equal(t, "pdflatex", req.Engine)
	assert.Equal(t, "output", req.OutputPath)
	assert.True(t, req.DoConvert)
	assert.True(t, req.DoClean)
	assert.True(t, req.Conversion.Enabled)
	assert.Equal(t, []string{"jpeg"}, req.Conversion.Formats)
}
