package domain

import (
	"errors"
	"time"
)

// DocumentID represents a unique identifier for a document
type DocumentID struct {
	Value string
}

// NewDocumentID creates a new document ID
func NewDocumentID(value string) (DocumentID, error) {
	if value == "" {
		return DocumentID{}, errors.New("document ID cannot be empty")
	}
	return DocumentID{Value: value}, nil
}

// String returns the string representation of the document ID
func (did DocumentID) String() string {
	return did.Value
}

// Equals compares two document IDs for equality
func (did DocumentID) Equals(other DocumentID) bool {
	return did.Value == other.Value
}

// DocumentStatus represents the status of a document
type DocumentStatus string

const (
	DocumentStatusPending    DocumentStatus = "pending"
	DocumentStatusProcessing DocumentStatus = "processing"
	DocumentStatusCompleted  DocumentStatus = "completed"
	DocumentStatusFailed     DocumentStatus = "failed"
	DocumentStatusCancelled  DocumentStatus = "cancelled"
)

// Document represents a generated document
type Document struct {
	ID          DocumentID
	TemplateID  TemplateID
	Variables   *VariableCollection
	OutputPath  OutputPath
	Status      DocumentStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ProcessedAt *time.Time
	Error       string
	Metadata    DocumentMetadata
}

// DocumentMetadata represents metadata for a document
type DocumentMetadata struct {
	Author      string
	Description string
	Tags        []string
	Properties  map[string]interface{}
}

// NewDocumentMetadata creates new document metadata
func NewDocumentMetadata(author, description string) DocumentMetadata {
	return DocumentMetadata{
		Author:      author,
		Description: description,
		Tags:        []string{},
		Properties:  make(map[string]interface{}),
	}
}

// Document errors
var (
	ErrEmptyDocumentID  = errors.New("document ID cannot be empty")
	ErrInvalidStatus    = errors.New("invalid document status")
	ErrStatusTransition = errors.New("invalid status transition")
	ErrDocumentNotFound = errors.New("document not found")
	ErrDocumentNotReady = errors.New("document is not ready for processing")
)

// NewDocument creates a new document
func NewDocument(id DocumentID, templateID TemplateID, outputPath OutputPath) (*Document, error) {
	if id.Value == "" {
		return nil, ErrEmptyDocumentID
	}

	return &Document{
		ID:         id,
		TemplateID: templateID,
		OutputPath: outputPath,
		Status:     DocumentStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Variables:  NewVariableCollection(),
		Metadata:   NewDocumentMetadata("", ""),
	}, nil
}

// Validate validates the document according to business rules
func (d *Document) Validate() error {
	if d.ID.Value == "" {
		return ErrEmptyDocumentID
	}
	if d.TemplateID.Value == "" {
		return errors.New("template ID cannot be empty")
	}
	if d.OutputPath.String() == "" {
		return errors.New("output path cannot be empty")
	}
	return nil
}

// CanBeProcessed returns true if the document can be processed
func (d *Document) CanBeProcessed() bool {
	return d.Status == DocumentStatusPending &&
		d.TemplateID.Value != "" &&
		d.OutputPath.String() != ""
}

// IsCompleted returns true if the document is completed
func (d *Document) IsCompleted() bool {
	return d.Status == DocumentStatusCompleted
}

// IsFailed returns true if the document processing failed
func (d *Document) IsFailed() bool {
	return d.Status == DocumentStatusFailed
}

// IsProcessing returns true if the document is being processed
func (d *Document) IsProcessing() bool {
	return d.Status == DocumentStatusProcessing
}

// StartProcessing starts processing the document
func (d *Document) StartProcessing() error {
	if !d.CanBeProcessed() {
		return ErrDocumentNotReady
	}

	d.Status = DocumentStatusProcessing
	d.UpdatedAt = time.Now()
	return nil
}

// CompleteProcessing completes processing the document
func (d *Document) CompleteProcessing() error {
	if d.Status != DocumentStatusProcessing {
		return ErrStatusTransition
	}

	d.Status = DocumentStatusCompleted
	now := time.Now()
	d.ProcessedAt = &now
	d.UpdatedAt = now
	return nil
}

// FailProcessing fails processing the document
func (d *Document) FailProcessing(errorMsg string) error {
	if d.Status != DocumentStatusProcessing {
		return ErrStatusTransition
	}

	d.Status = DocumentStatusFailed
	d.Error = errorMsg
	d.UpdatedAt = time.Now()
	return nil
}

// Cancel cancels the document processing
func (d *Document) Cancel() error {
	if d.Status == DocumentStatusCompleted {
		return ErrStatusTransition
	}

	d.Status = DocumentStatusCancelled
	d.UpdatedAt = time.Now()
	return nil
}

// GetProcessingTime returns the processing time if completed
func (d *Document) GetProcessingTime() *time.Duration {
	if d.ProcessedAt == nil {
		return nil
	}

	duration := d.ProcessedAt.Sub(d.CreatedAt)
	return &duration
}

// AddProperty adds a property to the document metadata
func (d *Document) AddProperty(key string, value interface{}) {
	if d.Metadata.Properties == nil {
		d.Metadata.Properties = make(map[string]interface{})
	}
	d.Metadata.Properties[key] = value
	d.UpdatedAt = time.Now()
}

// GetProperty gets a property from the document metadata
func (d *Document) GetProperty(key string) (interface{}, bool) {
	if d.Metadata.Properties == nil {
		return nil, false
	}
	value, exists := d.Metadata.Properties[key]
	return value, exists
}

// AddTag adds a tag to the document metadata
func (d *Document) AddTag(tag string) {
	if tag == "" {
		return
	}

	// Check if tag already exists
	for _, existingTag := range d.Metadata.Tags {
		if existingTag == tag {
			return
		}
	}

	d.Metadata.Tags = append(d.Metadata.Tags, tag)
	d.UpdatedAt = time.Now()
}

// RemoveTag removes a tag from the document metadata
func (d *Document) RemoveTag(tag string) {
	for i, existingTag := range d.Metadata.Tags {
		if existingTag == tag {
			d.Metadata.Tags = append(d.Metadata.Tags[:i], d.Metadata.Tags[i+1:]...)
			d.UpdatedAt = time.Now()
			return
		}
	}
}

// Equals compares two documents for equality
func (d *Document) Equals(other *Document) bool {
	return d.ID.Equals(other.ID) &&
		d.TemplateID.Equals(other.TemplateID) &&
		d.OutputPath.Equals(other.OutputPath)
}
