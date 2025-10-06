package domain

// TemplateRepository defines the interface for template data access
type TemplateRepository interface {
	Save(template *Template) error
	FindByID(id TemplateID) (*Template, error)
	FindByPath(path string) (*Template, error)
	Delete(id TemplateID) error
	FindAll() ([]*Template, error)
}

// DocumentRepository defines the interface for document data access
type DocumentRepository interface {
	Save(document *Document) error
	FindByID(id DocumentID) (*Document, error)
	FindByStatus(status DocumentStatus) ([]*Document, error)
	Delete(id DocumentID) error
	FindAll() ([]*Document, error)
}

// TemplateValidationService defines the interface for template validation
type TemplateValidationService interface {
	ValidateTemplate(template *Template) error
	ValidateSyntax(content string) error
	ValidateVariables(template *Template, variables *VariableCollection) error
}

// VariableResolutionService defines the interface for variable resolution
type VariableResolutionService interface {
	ResolveVariables(template *Template, inputVariables *VariableCollection) (*VariableCollection, error)
	ProcessNestedVariables(variables *VariableCollection) (*VariableCollection, error)
	ValidateVariableTypes(variables *VariableCollection) error
}

// DocumentGenerationService defines the interface for document generation
type DocumentGenerationService interface {
	GenerateDocument(templateID TemplateID, variables *VariableCollection, outputPath OutputPath) (*Document, error)
	ProcessDocument(document *Document) error
	ValidateDocument(document *Document) error
}
