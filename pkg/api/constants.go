// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api

// PDF Validation Error Messages
const (
	// PDF Path validation
	ErrPDFPathRequired = "PDF path is required"
	ErrPDFPathEmpty    = "PDF path cannot be empty"
	ErrPDFPathInvalid  = "Invalid PDF path format"

	// PDF File validation
	ErrPDFFileNotFound         = "PDF file does not exist: %s"
	ErrPDFFileNotReadable      = "Cannot read PDF file: %s"
	ErrPDFFileEmpty            = "PDF file is empty: %s"
	ErrPDFFileInvalid          = "File is not a valid PDF: %s"
	ErrPDFFileCorrupted        = "PDF file appears to be corrupted: %s"
	ErrPDFFileExtensionInvalid = "File does not have .pdf extension: %s"

	// PDF Structure validation
	ErrPDFHeaderInvalid    = "Invalid PDF header signature"
	ErrPDFStructureInvalid = "Invalid PDF structure"
	ErrPDFVersionInvalid   = "Unsupported PDF version"
	ErrPDFValidationFailed = "PDF validation failed"

	// PDF Metadata validation
	ErrPDFMetadataInvalid  = "Cannot extract PDF metadata: %s"
	ErrPDFPageCountInvalid = "Cannot determine page count: %s"
	ErrPDFFileSizeInvalid  = "Cannot determine file size: %s"

	// PDF Content validation
	ErrPDFContentInvalid  = "PDF content validation failed: %s"
	ErrPDFSecurityInvalid = "PDF security validation failed: %s"
)

// Template Processing Error Messages
const (
	// Template Path validation
	ErrTemplatePathRequired = "Template path is required"
	ErrTemplatePathEmpty    = "Template path cannot be empty"
	ErrTemplatePathInvalid  = "Invalid template path format"

	// Template File validation
	ErrTemplateFileNotFound    = "Template file does not exist: %s"
	ErrTemplateFileNotReadable = "Cannot read template file: %s"
	ErrTemplateFileEmpty       = "Template file is empty: %s"
	ErrTemplateFileInvalid     = "Template file is not a valid LaTeX file: %s"

	// Template Content validation
	ErrTemplateContentInvalid    = "Template content validation failed: %s"
	ErrTemplateSyntaxInvalid     = "Template contains invalid LaTeX syntax: %s"
	ErrTemplateVariablesInvalid  = "Template variables validation failed: %s"
	ErrTemplateDelimitersInvalid = "Template delimiters validation failed: %s"

	// Template Processing validation
	ErrTemplateProcessingFailed  = "Template processing failed: %s"
	ErrTemplateRenderingFailed   = "Template rendering failed: %s"
	ErrTemplateCompilationFailed = "Template compilation failed: %s"
	ErrTemplateValidationFailed  = "Template validation failed: %s"
)

// Variable Resolution Error Messages
const (
	// Variable validation
	ErrVariableKeyRequired  = "Variable key is required"
	ErrVariableKeyEmpty     = "Variable key cannot be empty"
	ErrVariableKeyInvalid   = "Invalid variable key format: %s"
	ErrVariableValueInvalid = "Invalid variable value: %s"
	ErrVariableTypeInvalid  = "Unsupported variable type: %s"

	// Variable resolution
	ErrVariableResolutionFailed = "Variable resolution failed: %s"
	ErrVariableFlatteningFailed = "Variable flattening failed: %s"
	ErrVariableValidationFailed = "Variable validation failed: %s"
)

// PDF Generation Error Messages
const (
	// Generation process
	ErrPDFGenerationFailed    = "PDF generation failed: %s"
	ErrPDFGenerationTimeout   = "PDF generation timeout: %s"
	ErrPDFGenerationCancelled = "PDF generation cancelled: %s"

	// Engine validation
	ErrEngineRequired        = "LaTeX engine is required"
	ErrEngineNotFound        = "LaTeX engine not found: %s"
	ErrEngineNotSupported    = "Unsupported LaTeX engine: %s"
	ErrEngineExecutionFailed = "LaTeX engine execution failed: %s"

	// Output validation
	ErrOutputPathRequired    = "Output path is required"
	ErrOutputPathInvalid     = "Invalid output path: %s"
	ErrOutputPathNotWritable = "Cannot write to output path: %s"
	ErrOutputPathExists      = "Output file already exists: %s"

	// Conversion validation
	ErrConversionFailed        = "PDF conversion failed: %s"
	ErrConversionFormatInvalid = "Unsupported conversion format: %s"
	ErrConversionTimeout       = "Conversion timeout: %s"
)

// System Error Messages
const (
	// File system errors
	ErrFileSystemAccess     = "File system access error: %s"
	ErrFileSystemPermission = "File system permission error: %s"
	ErrFileSystemSpace      = "Insufficient disk space: %s"

	// Memory errors
	ErrMemoryAllocation   = "Memory allocation failed: %s"
	ErrMemoryInsufficient = "Insufficient memory: %s"

	// Network errors
	ErrNetworkConnection  = "Network connection error: %s"
	ErrNetworkTimeout     = "Network timeout: %s"
	ErrNetworkUnavailable = "Network unavailable: %s"
)

// Error Categories
const (
	ErrorCategoryPDF           = "pdf_validation"
	ErrorCategoryTemplate      = "template_processing"
	ErrorCategoryVariable      = "variable_resolution"
	ErrorCategoryGeneration    = "pdf_generation"
	ErrorCategoryConfiguration = "configuration"
	ErrorCategorySystem        = "system"
	ErrorCategoryNetwork       = "network"
	ErrorCategoryFileSystem    = "file_system"
)

// Error Severity Levels
const (
	ErrorSeverityLow      = "low"
	ErrorSeverityMedium   = "medium"
	ErrorSeverityHigh     = "high"
	ErrorSeverityCritical = "critical"
)

// Error Context Keys
const (
	ContextKeyPDFPath      = "pdf_path"
	ContextKeyTemplatePath = "template_path"
	ContextKeyOutputPath   = "output_path"
	ContextKeyEngine       = "engine"
	ContextKeyFormat       = "format"
	ContextKeyError        = "error"
	ContextKeyTimestamp    = "timestamp"
	ContextKeyDuration     = "duration"
	ContextKeyFileSize     = "file_size"
	ContextKeyPageCount    = "page_count"
)

// Validation Rules
const (
	// PDF validation rules
	MinPDFFileSize     = 100               // Minimum PDF file size in bytes
	MaxPDFFileSize     = 100 * 1024 * 1024 // Maximum PDF file size (100MB)
	MinPDFHeaderLength = 4                 // Minimum PDF header length

	// Template validation rules
	MinTemplateFileSize = 50               // Minimum template file size in bytes
	MaxTemplateFileSize = 10 * 1024 * 1024 // Maximum template file size (10MB)

	// Variable validation rules
	MaxVariableKeyLength   = 100   // Maximum variable key length
	MaxVariableValueLength = 10000 // Maximum variable value length
	MaxVariableCount       = 1000  // Maximum number of variables
)

// Default Values
const (
	// PDF defaults
	DefaultPDFVersion   = "1.4"
	DefaultPDFPageCount = 1
	DefaultPDFFileSize  = 1024

	// Template defaults
	DefaultTemplateEngine = "pdflatex"
	DefaultTemplateFormat = "tex"

	// Validation defaults
	DefaultValidationTimeout = "30s"
	DefaultValidationRetries = 3
)
