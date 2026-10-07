// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"text/template"

	"github.com/BuddhiLW/AutoPDF/v2/configs"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

var errTemplateRequired = errors.New("template is required")

func capabilities(context.Context, json.RawMessage) (any, error) {
	engine, err := api.NewEngine()
	if err != nil {
		return nil, err
	}
	caps := engine.Capabilities()
	return map[string]any{
		"engines":            caps.Engines,
		"conversion_formats": caps.ConversionFormats,
	}, nil
}

type auditRequest struct {
	Name      string         `json:"name"`
	Template  string         `json:"template"`
	Variables map[string]any `json:"variables"`
}

type reference struct {
	Key    string `json:"key"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Waived bool   `json:"waived"`
	Reason string `json:"reason,omitempty"`
}

type offence struct {
	reference
	Cause string `json:"cause"`
}

// audit parses a template exactly as the renderer does (same delimiters,
// same strict functions) and reports every unsatisfied read at once.
func audit(_ context.Context, raw json.RawMessage) (any, error) {
	var req auditRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Template) == "" {
		return nil, errTemplateRequired
	}
	if req.Name == "" {
		req.Name = "template"
	}

	tmpl, err := template.New(req.Name).
		Funcs(strict.Funcs()).
		Delims(configs.TemplateStartDelim, configs.TemplateEndDelim).
		Option(strict.MissingKeyOption).
		Parse(req.Template)
	if err != nil {
		return nil, err
	}

	refs := strict.References(tmpl, req.Template)
	verdict := strict.Default().Audit(req.Name, refs, strict.Bindings(req.Variables))

	references := make([]reference, 0, len(refs))
	for _, r := range refs {
		references = append(references, toReference(r))
	}
	offences := make([]offence, 0, len(verdict.Offences))
	for _, o := range verdict.Offences {
		offences = append(offences, offence{reference: toReference(o.Reference), Cause: string(o.Cause)})
	}
	return map[string]any{
		"satisfied":  verdict.OK(),
		"references": references,
		"offences":   offences,
	}, nil
}

func toReference(r strict.Reference) reference {
	return reference{Key: r.Key, Line: r.Line, Column: r.Column, Waived: r.Waived, Reason: r.Reason}
}

type generateRequest struct {
	TemplatePath string            `json:"template_path"`
	OutputPath   string            `json:"output_path"`
	Variables    map[string]string `json:"variables"`
	Engine       string            `json:"engine"`
	WorkingDir   string            `json:"working_dir"`
	Passes       int               `json:"passes"`
}

// generate renders through api.Engine, the supported embedding surface. The
// PDF stays on disk; only its path and metadata cross the boundary.
func generate(ctx context.Context, raw json.RawMessage) (any, error) {
	var req generateRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}
	engine, err := api.NewEngine()
	if err != nil {
		return nil, err
	}
	result, err := engine.Generate(ctx, api.Request{
		TemplatePath: req.TemplatePath,
		OutputPath:   req.OutputPath,
		Variables:    req.Variables,
		LaTeXEngine:  req.Engine,
		WorkingDir:   req.WorkingDir,
		Passes:       req.Passes,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"pdf_path":  result.PDFPath,
		"file_size": result.Metadata.FileSize,
		"pages":     result.Metadata.PageCount,
		"engine":    result.Metadata.Engine,
	}, nil
}
