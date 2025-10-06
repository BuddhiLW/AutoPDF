package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentID_NewDocumentID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DocumentID
		wantErr error
	}{
		{
			name:    "valid document ID",
			input:   "doc-123",
			want:    DocumentID{Value: "doc-123"},
			wantErr: nil,
		},
		{
			name:    "empty document ID",
			input:   "",
			want:    DocumentID{},
			wantErr: errors.New("document ID cannot be empty"),
		},
		{
			name:    "document ID with spaces",
			input:   "  doc-123  ",
			want:    DocumentID{Value: "  doc-123  "},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDocumentID(tt.input)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("NewDocumentID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("NewDocumentID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewDocumentID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocumentID_String(t *testing.T) {
	id, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	assert.Equal(t, "doc-123", id.String())
}

func TestDocumentID_Equals(t *testing.T) {
	id1, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	id2, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	id3, err := NewDocumentID("doc-456")
	require.NoError(t, err)

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}

func TestDocumentMetadata_NewDocumentMetadata(t *testing.T) {
	metadata := NewDocumentMetadata("John Doe", "Test document")

	assert.Equal(t, "John Doe", metadata.Author)
	assert.Equal(t, "Test document", metadata.Description)
	assert.Empty(t, metadata.Tags)
	assert.NotNil(t, metadata.Properties)
}

func TestDocument_NewDocument(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		templateID string
		outputPath string
		wantErr    error
	}{
		{
			name:       "valid document",
			id:         "doc-123",
			templateID: "template-123",
			outputPath: "output.pdf",
			wantErr:    nil,
		},
		{
			name:       "empty document ID",
			id:         "",
			templateID: "template-123",
			outputPath: "output.pdf",
			wantErr:    ErrEmptyDocumentID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docID, err := NewDocumentID(tt.id)
			if tt.wantErr == nil {
				require.NoError(t, err)
			}

			templateID, err := NewTemplateID(tt.templateID)
			require.NoError(t, err)

			outputPath, err := NewOutputPath(tt.outputPath)
			require.NoError(t, err)

			document, err := NewDocument(docID, templateID, outputPath)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				assert.NotNil(t, document)
				assert.Equal(t, DocumentStatusPending, document.Status)
				assert.NotZero(t, document.CreatedAt)
				assert.NotZero(t, document.UpdatedAt)
			}
		})
	}
}

func TestDocument_Validate(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	// Valid document
	assert.NoError(t, document.Validate())

	// Invalid document ID
	document.ID = DocumentID{}
	assert.Error(t, document.Validate())

	// Reset for next test
	document.ID = docID

	// Invalid template ID
	document.TemplateID = TemplateID{}
	assert.Error(t, document.Validate())

	// Reset for next test
	document.TemplateID = templateID

	// Invalid output path
	document.OutputPath = OutputPath{}
	assert.Error(t, document.Validate())
}

func TestDocument_CanBeProcessed(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	// Pending document can be processed
	assert.True(t, document.CanBeProcessed())

	// Processing document cannot be processed
	document.Status = DocumentStatusProcessing
	assert.False(t, document.CanBeProcessed())

	// Completed document cannot be processed
	document.Status = DocumentStatusCompleted
	assert.False(t, document.CanBeProcessed())

	// Failed document cannot be processed
	document.Status = DocumentStatusFailed
	assert.False(t, document.CanBeProcessed())

	// Cancelled document cannot be processed
	document.Status = DocumentStatusCancelled
	assert.False(t, document.CanBeProcessed())
}

func TestDocument_StatusChecks(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	// Initially pending
	assert.False(t, document.IsCompleted())
	assert.False(t, document.IsFailed())
	assert.False(t, document.IsProcessing())

	// Test processing status
	document.Status = DocumentStatusProcessing
	assert.False(t, document.IsCompleted())
	assert.False(t, document.IsFailed())
	assert.True(t, document.IsProcessing())

	// Test completed status
	document.Status = DocumentStatusCompleted
	assert.True(t, document.IsCompleted())
	assert.False(t, document.IsFailed())
	assert.False(t, document.IsProcessing())

	// Test failed status
	document.Status = DocumentStatusFailed
	assert.False(t, document.IsCompleted())
	assert.True(t, document.IsFailed())
	assert.False(t, document.IsProcessing())
}

func TestDocument_StartProcessing(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	// Start processing from pending status
	err = document.StartProcessing()
	require.NoError(t, err)
	assert.Equal(t, DocumentStatusProcessing, document.Status)

	// Try to start processing from processing status (should fail)
	err = document.StartProcessing()
	assert.Error(t, err)
	assert.Equal(t, ErrDocumentNotReady, err)
}

func TestDocument_CompleteProcessing(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	// Complete processing from processing status
	document.Status = DocumentStatusProcessing
	err = document.CompleteProcessing()
	require.NoError(t, err)
	assert.Equal(t, DocumentStatusCompleted, document.Status)
	assert.NotNil(t, document.ProcessedAt)

	// Try to complete processing from pending status (should fail)
	document.Status = DocumentStatusPending
	err = document.CompleteProcessing()
	assert.Error(t, err)
	assert.Equal(t, ErrStatusTransition, err)
}

func TestDocument_FailProcessing(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	// Fail processing from processing status
	document.Status = DocumentStatusProcessing
	errorMsg := "Template compilation failed"
	err = document.FailProcessing(errorMsg)
	require.NoError(t, err)
	assert.Equal(t, DocumentStatusFailed, document.Status)
	assert.Equal(t, errorMsg, document.Error)

	// Try to fail processing from pending status (should fail)
	document.Status = DocumentStatusPending
	err = document.FailProcessing("Another error")
	assert.Error(t, err)
	assert.Equal(t, ErrStatusTransition, err)
}

func TestDocument_Cancel(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	// Cancel from pending status
	err = document.Cancel()
	require.NoError(t, err)
	assert.Equal(t, DocumentStatusCancelled, document.Status)

	// Reset status
	document.Status = DocumentStatusProcessing

	// Cancel from processing status
	err = document.Cancel()
	require.NoError(t, err)
	assert.Equal(t, DocumentStatusCancelled, document.Status)

	// Try to cancel from completed status (should fail)
	document.Status = DocumentStatusCompleted
	err = document.Cancel()
	assert.Error(t, err)
	assert.Equal(t, ErrStatusTransition, err)
}

func TestDocument_GetProcessingTime(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	// Initially no processing time
	assert.Nil(t, document.GetProcessingTime())

	// Complete processing
	document.Status = DocumentStatusProcessing
	err = document.CompleteProcessing()
	require.NoError(t, err)

	// Now should have processing time
	processingTime := document.GetProcessingTime()
	assert.NotNil(t, processingTime)
	assert.True(t, *processingTime >= 0)
}

func TestDocument_AddProperty(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	originalUpdatedAt := document.UpdatedAt

	// Add property
	document.AddProperty("author", "John Doe")
	document.AddProperty("version", "1.0.0")

	// Check properties
	author, exists := document.GetProperty("author")
	assert.True(t, exists)
	assert.Equal(t, "John Doe", author)

	version, exists := document.GetProperty("version")
	assert.True(t, exists)
	assert.Equal(t, "1.0.0", version)

	// Check that updatedAt was updated
	assert.True(t, document.UpdatedAt.After(originalUpdatedAt))
}

func TestDocument_AddTag(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	originalUpdatedAt := document.UpdatedAt

	// Add tags
	document.AddTag("important")
	document.AddTag("draft")

	assert.Contains(t, document.Metadata.Tags, "important")
	assert.Contains(t, document.Metadata.Tags, "draft")
	assert.Len(t, document.Metadata.Tags, 2)

	// Add duplicate tag (should not be added)
	document.AddTag("important")
	assert.Len(t, document.Metadata.Tags, 2)

	// Add empty tag (should not be added)
	document.AddTag("")
	assert.Len(t, document.Metadata.Tags, 2)

	// Check that updatedAt was updated
	assert.True(t, document.UpdatedAt.After(originalUpdatedAt))
}

func TestDocument_RemoveTag(t *testing.T) {
	docID, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document, err := NewDocument(docID, templateID, outputPath)
	require.NoError(t, err)

	// Add tags
	document.AddTag("important")
	document.AddTag("draft")

	originalUpdatedAt := document.UpdatedAt

	// Remove existing tag
	document.RemoveTag("important")
	assert.NotContains(t, document.Metadata.Tags, "important")
	assert.Contains(t, document.Metadata.Tags, "draft")
	assert.Len(t, document.Metadata.Tags, 1)

	// Remove non-existing tag
	document.RemoveTag("nonexistent")
	assert.Len(t, document.Metadata.Tags, 1)

	// Check that updatedAt was updated
	assert.True(t, document.UpdatedAt.After(originalUpdatedAt))
}

func TestDocument_Equals(t *testing.T) {
	docID1, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID1, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath1, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document1, err := NewDocument(docID1, templateID1, outputPath1)
	require.NoError(t, err)

	docID2, err := NewDocumentID("doc-123")
	require.NoError(t, err)

	templateID2, err := NewTemplateID("template-123")
	require.NoError(t, err)

	outputPath2, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	document2, err := NewDocument(docID2, templateID2, outputPath2)
	require.NoError(t, err)

	docID3, err := NewDocumentID("doc-456")
	require.NoError(t, err)

	document3, err := NewDocument(docID3, templateID1, outputPath1)
	require.NoError(t, err)

	assert.True(t, document1.Equals(document2))
	assert.False(t, document1.Equals(document3))
}
