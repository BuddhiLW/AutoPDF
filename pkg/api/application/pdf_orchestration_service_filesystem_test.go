// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stagingFileSystem struct {
	api.FileSystem
	files     map[string]*bytes.Buffer
	removed   []string
	createErr error
}

func newStagingFileSystem() *stagingFileSystem {
	return &stagingFileSystem{files: map[string]*bytes.Buffer{}}
}

func (s *stagingFileSystem) CreateTemp(dir, pattern string) (api.TempFile, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	name := "/virtual/" + pattern
	s.files[name] = &bytes.Buffer{}
	return &stagedFile{Buffer: s.files[name], name: name}, nil
}

func (s *stagingFileSystem) Remove(path string) error {
	if _, ok := s.files[path]; !ok {
		return &fs.PathError{Op: "remove", Path: path, Err: fs.ErrNotExist}
	}
	s.removed = append(s.removed, path)
	delete(s.files, path)
	return nil
}

type stagedFile struct {
	*bytes.Buffer
	name string
}

func (f *stagedFile) Close() error { return nil }
func (f *stagedFile) Name() string { return f.name }

type stubTemplates struct{ content string }

func (s stubTemplates) Process(context.Context, string, map[string]string) (string, error) {
	return s.content, nil
}
func (stubTemplates) ValidateTemplate(string) error                 { return nil }
func (stubTemplates) GetTemplateVariables(string) ([]string, error) { return nil, nil }

type stubResolver struct{}

func (stubResolver) Resolve(*generation.TemplateVariables) (map[string]string, error) {
	return map[string]string{}, nil
}
func (stubResolver) Flatten(*generation.TemplateVariables) map[string]string { return nil }
func (stubResolver) Validate(*generation.TemplateVariables) error            { return nil }

type recordingGenerator struct {
	staging *stagingFileSystem
	path    string
	content string
}

func (g *recordingGenerator) Generate(_ context.Context, req generation.PDFGenerationRequest) (generation.PDFGenerationResult, error) {
	g.path = req.TemplatePath
	if buffer, ok := g.staging.files[req.TemplatePath]; ok {
		g.content = buffer.String()
	}
	return generation.PDFGenerationResult{Success: true}, nil
}
func (*recordingGenerator) ValidateRequest(generation.PDFGenerationRequest) error { return nil }
func (*recordingGenerator) GetSupportedEngines() []string                         { return nil }
func (*recordingGenerator) GetSupportedFormats() []string                         { return nil }

func newStubOrchestration(staging *stagingFileSystem, external generation.PDFGenerationService) *PDFOrchestrationService {
	return NewPDFOrchestrationService(
		stubTemplates{content: `\documentclass{article}`},
		stubResolver{},
		nil,
		external,
		nil,
		nil,
		nil,
	).WithFileSystem(staging)
}

func stubRequest() generation.PDFGenerationRequest {
	return generation.PDFGenerationRequest{
		TemplatePath: "/templates/letter.tex",
		Engine:       "pdflatex",
		OutputPath:   "/out/letter.pdf",
	}
}

func TestGeneratePDFStagesProcessedTemplateThroughFileSystemPort(t *testing.T) {
	staging := newStagingFileSystem()
	external := &recordingGenerator{staging: staging}

	_, err := newStubOrchestration(staging, external).GeneratePDF(context.Background(), stubRequest())

	require.NoError(t, err)
	assert.Equal(t, "/virtual/autopdf-processed-*.tex", external.path)
	assert.Equal(t, `\documentclass{article}`, external.content)
	assert.Equal(t, []string{external.path}, staging.removed)
	assert.Empty(t, staging.files)
}

func TestGeneratePDFReportsStagingFailureFromFileSystemPort(t *testing.T) {
	staging := newStagingFileSystem()
	staging.createErr = errors.New("read-only filesystem")
	external := &recordingGenerator{staging: staging}

	result, err := newStubOrchestration(staging, external).GeneratePDF(context.Background(), stubRequest())

	require.ErrorIs(t, err, staging.createErr)
	assert.False(t, result.Success)
	assert.Empty(t, external.path)
}

func TestPDFOrchestrationServiceDefaultsToOSFileSystem(t *testing.T) {
	service := NewPDFOrchestrationService(stubTemplates{}, stubResolver{}, nil, nil, nil, nil, nil)
	assert.IsType(t, api.OSFileSystem{}, service.fileSystem)
	assert.IsType(t, api.OSFileSystem{}, service.WithFileSystem(nil).fileSystem)
}
