package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// TemplateID represents a unique identifier for a template
type TemplateID struct {
	Value string
}

// NewTemplateID creates a new template ID
func NewTemplateID(value string) (TemplateID, error) {
	if value == "" {
		return TemplateID{}, errors.New("template ID cannot be empty")
	}
	return TemplateID{Value: value}, nil
}

// String returns the string representation of the template ID
func (tid TemplateID) String() string {
	return tid.Value
}

// Equals compares two template IDs for equality
func (tid TemplateID) Equals(other TemplateID) bool {
	return tid.Value == other.Value
}

// TemplateType represents the type of template
type TemplateType string

const (
	TemplateTypeLaTeX    TemplateType = "latex"
	TemplateTypeABNTeX   TemplateType = "abntex"
	TemplateTypeMarkdown TemplateType = "markdown"
	TemplateTypeHTML     TemplateType = "html"
)

// TemplateMetadata represents metadata for a template
type TemplateMetadata struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     string
	Author      string
	Description string
	Tags        []string
	Type        TemplateType
}

// NewTemplateMetadata creates new template metadata
func NewTemplateMetadata(author, description string, templateType TemplateType) TemplateMetadata {
	now := time.Now()
	return TemplateMetadata{
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     "1.0.0",
		Author:      author,
		Description: description,
		Tags:        []string{},
		Type:        templateType,
	}
}

// Template represents a template document
type Template struct {
	ID        TemplateID
	Path      TemplatePath
	Content   string
	Variables *VariableCollection
	Metadata  TemplateMetadata
	Type      TemplateType
}

// Template errors
var (
	ErrEmptyTemplate    = errors.New("template content cannot be empty")
	ErrTemplateTooLarge = errors.New("template content is too large")
	ErrInvalidTemplate  = errors.New("invalid template")
	ErrTemplateNotFound = errors.New("template not found")
)

// NewTemplate creates a new template
func NewTemplate(id TemplateID, path TemplatePath, content string, templateType TemplateType, metadata TemplateMetadata) (*Template, error) {
	if content == "" {
		return nil, ErrEmptyTemplate
	}
	if len(content) > 1000000 { // 1MB limit
		return nil, ErrTemplateTooLarge
	}
	if !isValidTemplateType(templateType) {
		return nil, errors.New("invalid template type")
	}

	return &Template{
		ID:        id,
		Path:      path,
		Content:   content,
		Variables: NewVariableCollection(),
		Metadata:  metadata,
		Type:      templateType,
	}, nil
}

// isValidTemplateType checks if the template type is valid
func isValidTemplateType(templateType TemplateType) bool {
	switch templateType {
	case TemplateTypeLaTeX, TemplateTypeABNTeX, TemplateTypeMarkdown, TemplateTypeHTML:
		return true
	default:
		return false
	}
}

// Validate validates the template according to business rules
func (t *Template) Validate() error {
	if t.Content == "" {
		return ErrEmptyTemplate
	}
	if len(t.Content) > 1000000 {
		return ErrTemplateTooLarge
	}
	if t.ID.Value == "" {
		return errors.New("template ID cannot be empty")
	}
	return nil
}

// HasVariables returns true if the template has variables
func (t *Template) HasVariables() bool {
	return t.Variables != nil && t.Variables.Size() > 0
}

// GetVariableNames returns the names of all variables in the template
func (t *Template) GetVariableNames() []string {
	if t.Variables == nil {
		return []string{}
	}
	return t.Variables.Keys()
}

// UpdateContent updates the template content
func (t *Template) UpdateContent(content string) error {
	if content == "" {
		return ErrEmptyTemplate
	}
	if len(content) > 1000000 {
		return ErrTemplateTooLarge
	}

	t.Content = content
	t.Metadata.UpdatedAt = time.Now()
	return nil
}

// UpdateMetadata updates the template metadata
func (t *Template) UpdateMetadata(metadata TemplateMetadata) {
	t.Metadata = metadata
	t.Metadata.UpdatedAt = time.Now()
}

// AddTag adds a tag to the template
func (t *Template) AddTag(tag string) {
	if tag == "" {
		return
	}

	// Check if tag already exists
	for _, existingTag := range t.Metadata.Tags {
		if existingTag == tag {
			return
		}
	}

	t.Metadata.Tags = append(t.Metadata.Tags, tag)
	t.Metadata.UpdatedAt = time.Now()
}

// RemoveTag removes a tag from the template
func (t *Template) RemoveTag(tag string) {
	for i, existingTag := range t.Metadata.Tags {
		if existingTag == tag {
			t.Metadata.Tags = append(t.Metadata.Tags[:i], t.Metadata.Tags[i+1:]...)
			t.Metadata.UpdatedAt = time.Now()
			return
		}
	}
}

// GetSize returns the size of the template content
func (t *Template) GetSize() int {
	return len(t.Content)
}

// Equals compares two templates for equality
func (t *Template) Equals(other *Template) bool {
	return t.ID.Equals(other.ID) &&
		t.Path.Equals(other.Path) &&
		t.Content == other.Content
}

// GetType returns the template type
func (t *Template) GetType() TemplateType {
	return t.Type
}

// SetType sets the template type
func (t *Template) SetType(templateType TemplateType) error {
	if !isValidTemplateType(templateType) {
		return errors.New("invalid template type")
	}
	t.Type = templateType
	t.Metadata.UpdatedAt = time.Now()
	return nil
}

// AddVariable adds a variable to the template
func (t *Template) AddVariable(key string, variable *Variable) {
	if t.Variables == nil {
		t.Variables = NewVariableCollection()
	}
	t.Variables.Set(key, variable)
}

// RemoveVariable removes a variable from the template
func (t *Template) RemoveVariable(key string) error {
	if t.Variables == nil {
		return nil
	}
	if !t.Variables.Has(key) {
		return fmt.Errorf("variable '%s' not found", key)
	}
	t.Variables.Delete(key)
	return nil
}

// HasVariable checks if the template has a specific variable
func (t *Template) HasVariable(key string) bool {
	if t.Variables == nil {
		return false
	}
	return t.Variables.Has(key)
}

// GetVariable returns a variable from the template
func (t *Template) GetVariable(key string) (*Variable, error) {
	if t.Variables == nil {
		return nil, errors.New("no variables in template")
	}
	variable, exists := t.Variables.Get(key)
	if !exists {
		return nil, fmt.Errorf("variable '%s' not found", key)
	}
	return variable, nil
}

// ValidateTemplate validates the template content and variables
func (t *Template) ValidateTemplate() error {
	if err := t.Validate(); err != nil {
		return err
	}
	if t.Variables != nil {
		// Validate all variables in the collection
		for key, variable := range t.Variables.GetAll() {
			if variable == nil {
				return fmt.Errorf("variable '%s' is nil", key)
			}
			if err := variable.Validate(); err != nil {
				return fmt.Errorf("variable '%s' validation failed: %w", key, err)
			}
		}
	}
	return nil
}

// CanBeProcessed checks if the template can be processed
func (t *Template) CanBeProcessed() bool {
	if t.Variables == nil {
		return false
	}
	return t.ValidateTemplate() == nil
}

// DetermineTemplateType determines the template type from content
func DetermineTemplateType(content string) TemplateType {
	if strings.Contains(content, "\\documentclass") {
		if strings.Contains(content, "\\usepackage{abntex2}") {
			return TemplateTypeABNTeX
		}
		return TemplateTypeLaTeX
	}
	if strings.Contains(content, "# ") {
		return TemplateTypeMarkdown
	}
	if strings.Contains(content, "<html>") || strings.Contains(content, "<!DOCTYPE html>") {
		return TemplateTypeHTML
	}
	return TemplateTypeLaTeX // Default
}

// TemplateFactory creates templates based on type (Factory Pattern)
type TemplateFactory interface {
	CreateTemplate(id TemplateID, path TemplatePath, content string, metadata TemplateMetadata) (*Template, error)
	CreateTemplateFromType(templateType TemplateType, id TemplateID, path TemplatePath, content string, metadata TemplateMetadata) (*Template, error)
}

// DefaultTemplateFactory implements TemplateFactory
type DefaultTemplateFactory struct{}

// NewDefaultTemplateFactory creates a new default template factory
func NewDefaultTemplateFactory() *DefaultTemplateFactory {
	return &DefaultTemplateFactory{}
}

// CreateTemplate creates a template with the specified type
func (tf *DefaultTemplateFactory) CreateTemplate(id TemplateID, path TemplatePath, content string, metadata TemplateMetadata) (*Template, error) {
	// Determine template type from content
	templateType := DetermineTemplateType(content)
	return NewTemplate(id, path, content, templateType, metadata)
}

// CreateTemplateFromType creates a template with the specified type
func (tf *DefaultTemplateFactory) CreateTemplateFromType(templateType TemplateType, id TemplateID, path TemplatePath, content string, metadata TemplateMetadata) (*Template, error) {
	return NewTemplate(id, path, content, templateType, metadata)
}

// TemplateProcessingStrategy defines the strategy for processing templates (Strategy Pattern)
type TemplateProcessingStrategy interface {
	Process(template *Template, variables *VariableCollection) (string, error)
	Supports(templateType TemplateType) bool
	GetName() string
}

// LaTeXProcessingStrategy implements LaTeX template processing
type LaTeXProcessingStrategy struct {
	delimiters DelimiterConfig
}

// DelimiterConfig represents delimiter configuration
type DelimiterConfig struct {
	Left  string
	Right string
}

// NewLaTeXProcessingStrategy creates a new LaTeX processing strategy
func NewLaTeXProcessingStrategy(delimiters DelimiterConfig) *LaTeXProcessingStrategy {
	return &LaTeXProcessingStrategy{
		delimiters: delimiters,
	}
}

// Process processes a LaTeX template
func (s *LaTeXProcessingStrategy) Process(template *Template, variables *VariableCollection) (string, error) {
	if template == nil {
		return "", errors.New("template cannot be nil")
	}
	if variables == nil {
		return "", errors.New("variables cannot be nil")
	}

	content := template.Content

	// Process variables in template content
	for key, variable := range variables.variables {
		placeholder := fmt.Sprintf("%s.%s%s", s.delimiters.Left, key, s.delimiters.Right)

		var replacement string
		switch variable.Type {
		case VariableTypeString:
			if str, err := variable.AsString(); err == nil {
				replacement = str
			}
		case VariableTypeNumber:
			if num, err := variable.AsNumber(); err == nil {
				replacement = fmt.Sprintf("%.2f", num)
			}
		case VariableTypeBoolean:
			if boolVal, err := variable.AsBoolean(); err == nil {
				replacement = fmt.Sprintf("%t", boolVal)
			}
		case VariableTypeArray:
			if arr, err := variable.AsArray(); err == nil {
				replacement = strings.Join(convertToStringSlice(arr), ", ")
			}
		case VariableTypeObject:
			if obj, err := variable.AsObject(); err == nil {
				replacement = fmt.Sprintf("%v", obj)
			}
		case VariableTypeNull:
			replacement = ""
		}

		content = strings.ReplaceAll(content, placeholder, replacement)
	}

	return content, nil
}

// Supports checks if this strategy supports the template type
func (s *LaTeXProcessingStrategy) Supports(templateType TemplateType) bool {
	return templateType == TemplateTypeLaTeX
}

// GetName returns the strategy name
func (s *LaTeXProcessingStrategy) GetName() string {
	return "LaTeX Processing Strategy"
}

// Helper function to convert interface{} slice to string slice
func convertToStringSlice(arr []interface{}) []string {
	result := make([]string, len(arr))
	for i, v := range arr {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result
}

// TemplateProcessingContext manages template processing strategies (Context Pattern)
type TemplateProcessingContext struct {
	strategies      map[TemplateType]TemplateProcessingStrategy
	defaultStrategy TemplateProcessingStrategy
}

// NewTemplateProcessingContext creates a new template processing context
func NewTemplateProcessingContext() *TemplateProcessingContext {
	delimiters := DelimiterConfig{Left: "{{", Right: "}}"}

	strategies := make(map[TemplateType]TemplateProcessingStrategy)
	strategies[TemplateTypeLaTeX] = NewLaTeXProcessingStrategy(delimiters)

	return &TemplateProcessingContext{
		strategies:      strategies,
		defaultStrategy: NewLaTeXProcessingStrategy(delimiters),
	}
}

// ProcessTemplate processes a template using the appropriate strategy
func (ctx *TemplateProcessingContext) ProcessTemplate(template *Template, variables *VariableCollection) (string, error) {
	if template == nil {
		return "", errors.New("template cannot be nil")
	}

	strategy, exists := ctx.strategies[template.Type]
	if !exists {
		strategy = ctx.defaultStrategy
	}

	return strategy.Process(template, variables)
}

// AddStrategy adds a new processing strategy
func (ctx *TemplateProcessingContext) AddStrategy(templateType TemplateType, strategy TemplateProcessingStrategy) {
	ctx.strategies[templateType] = strategy
}
