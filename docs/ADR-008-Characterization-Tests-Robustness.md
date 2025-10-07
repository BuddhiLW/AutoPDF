# ADR-008: Characterization Tests Robustness

## Status
Accepted

## Context
The initial characterization tests proposed comparing full PDF bytes, which can be flaky across LaTeX versions and system differences. We need more robust tests that focus on essential behavior rather than exact byte matching.

## Decision

**Robust Characterization Testing Strategy**

1. **Template Rendering Verification**: Check processed template content
2. **PDF Existence and Size**: Verify output file exists with reasonable size
3. **Text Content Extraction**: Extract and verify key text content
4. **Single End-to-End Smoke Test**: One comprehensive PDF test
5. **Domain/App Layer Assertions**: Move most assertions to domain and application layers

### Rationale
- **LaTeX Version Independence**: Tests don't break with LaTeX updates
- **System Independence**: Tests work across different environments
- **Focused Assertions**: Test what matters for business logic
- **Maintainable**: Easier to maintain and update
- **Fast**: Faster test execution

## Implementation

### Template Rendering Tests
```go
// test/characterization/template_rendering_test.go
package characterization

func TestTemplateRendering_GoldenTests(t *testing.T) {
    tests := []struct {
        name           string
        templatePath   string
        configPath     string
        expectedTokens []string
    }{
        {
            name:         "basic_template_rendering",
            templatePath: "testdata/template.tex",
            configPath:   "testdata/config.yaml",
            expectedTokens: []string{
                "Test Document",
                "This is a test",
                "\\documentclass{scrartcl}",
            },
        },
        {
            name:         "template_with_variables",
            templatePath: "testdata/template_with_vars.tex",
            configPath:   "testdata/config_with_vars.yaml",
            expectedTokens: []string{
                "Hello World",
                "Custom Title",
                "\\title{Custom Title}",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Process template
            processedContent, err := processTemplate(tt.templatePath, tt.configPath)
            require.NoError(t, err)
            
            // Verify expected tokens are present
            for _, token := range tt.expectedTokens {
                assert.Contains(t, processedContent, token, "Expected token not found: %s", token)
            }
        })
    }
}
```

### PDF Existence and Size Tests
```go
// test/characterization/pdf_generation_test.go
package characterization

func TestPDFGeneration_ExistenceAndSize(t *testing.T) {
    tests := []struct {
        name           string
        templatePath   string
        configPath     string
        minSizeBytes   int64
        maxSizeBytes   int64
    }{
        {
            name:         "basic_pdf_generation",
            templatePath: "testdata/template.tex",
            configPath:   "testdata/config.yaml",
            minSizeBytes: 1000,  // At least 1KB
            maxSizeBytes: 100000, // At most 100KB
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Generate PDF
            outputPath, err := generatePDF(tt.templatePath, tt.configPath)
            require.NoError(t, err)
            
            // Check file exists
            assert.FileExists(t, outputPath)
            
            // Check file size
            stat, err := os.Stat(outputPath)
            require.NoError(t, err)
            
            assert.GreaterOrEqual(t, stat.Size(), tt.minSizeBytes, "PDF too small")
            assert.LessOrEqual(t, stat.Size(), tt.maxSizeBytes, "PDF too large")
        })
    }
}
```

### Text Content Extraction Tests
```go
// test/characterization/text_extraction_test.go
package characterization

func TestPDFTextExtraction(t *testing.T) {
    tests := []struct {
        name           string
        templatePath   string
        configPath     string
        expectedText   []string
    }{
        {
            name:         "basic_text_content",
            templatePath: "testdata/template.tex",
            configPath:   "testdata/config.yaml",
            expectedText: []string{
                "Test Document",
                "This is a test",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Generate PDF
            outputPath, err := generatePDF(tt.templatePath, tt.configPath)
            require.NoError(t, err)
            
            // Extract text content
            textContent, err := extractPDFText(outputPath)
            require.NoError(t, err)
            
            // Verify expected text is present
            for _, expected := range tt.expectedText {
                assert.Contains(t, textContent, expected, "Expected text not found: %s", expected)
            }
        })
    }
}
```

### Single End-to-End Smoke Test
```go
// test/characterization/smoke_test.go
package characterization

func TestEndToEnd_SmokeTest(t *testing.T) {
    // This is the only test that generates a full PDF and checks it
    // Keep this minimal and focused on critical path
    
    templatePath := "testdata/smoke_template.tex"
    configPath := "testdata/smoke_config.yaml"
    
    // Generate PDF
    outputPath, err := generatePDF(templatePath, configPath)
    require.NoError(t, err)
    
    // Basic checks
    assert.FileExists(t, outputPath)
    
    // Check it's a valid PDF
    assert.True(t, isValidPDF(outputPath), "Generated file is not a valid PDF")
    
    // Check basic content
    textContent, err := extractPDFText(outputPath)
    require.NoError(t, err)
    assert.Contains(t, textContent, "Smoke Test Document")
}
```

### Domain/App Layer Tests
```go
// test/domain/document_entity_test.go
package domain

func TestDocument_StateTransitions(t *testing.T) {
    doc := entities.NewDocument(
        "doc_123",
        "template.tex",
        "output.pdf",
        map[string]string{"title": "Test"},
        "pdflatex",
    )
    
    // Test state transitions
    assert.True(t, doc.CanBeProcessed())
    
    err := doc.StartProcessing()
    assert.NoError(t, err)
    assert.True(t, doc.IsProcessing())
    
    result := entities.NewCompilationResult("comp_123", "doc_123", "output.pdf", true, time.Second)
    err = doc.CompleteProcessing(*result)
    assert.NoError(t, err)
    assert.True(t, doc.IsCompleted())
}

// test/app/generate_document_service_test.go
package app

func TestGenerateDocumentService_Execute(t *testing.T) {
    // Test application service with mocked dependencies
    service := NewGenerateDocumentService(
        mockDocumentRepo,
        mockTemplateProc,
        mockLaTeXCompiler,
        mockFileSystem,
        mockEventPublisher,
        featureFlags,
    )
    
    cmd := GenerateDocumentCommand{
        TemplatePath: "template.tex",
        OutputPath:   "output.pdf",
        Variables:    map[string]string{"title": "Test"},
        Engine:       "pdflatex",
    }
    
    result, err := service.Execute(ctx, cmd)
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

## Benefits

1. **LaTeX Version Independence**: Tests don't break with LaTeX updates
2. **System Independence**: Tests work across different environments
3. **Focused Assertions**: Test what matters for business logic
4. **Maintainable**: Easier to maintain and update
5. **Fast**: Faster test execution
6. **Reliable**: Less flaky than byte comparison

## Success Criteria

- [ ] Template rendering tests verify processed content
- [ ] PDF existence and size tests pass consistently
- [ ] Text extraction tests verify content
- [ ] Single smoke test for end-to-end verification
- [ ] Domain and app layer tests cover business logic
- [ ] Tests are fast and reliable

## Consequences

- **Text Extraction Dependency**: Need PDF text extraction capability
- **Template Content Testing**: Need to verify template processing
- **Size Thresholds**: Need to set reasonable size limits

## Mitigation

- **PDF Text Extraction**: Use existing libraries or simple text extraction
- **Template Content**: Focus on key business tokens
- **Size Thresholds**: Set generous but reasonable limits
- **Fallback Tests**: Keep some basic existence tests as fallback
