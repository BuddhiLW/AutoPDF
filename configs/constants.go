// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package configs

// File and Directory Constants
const (
	// Config directory and file names
	ConfigDirName     = ".autopdf"
	ConfigFileName    = "config.yaml"
	FallbackConfigDir = "autopdf-config.yaml"
)

// LaTeX Auxiliary File Extensions
var AuxiliaryExtensions = []string{
	".aux", ".log", ".toc", ".lof", ".lot", ".out", ".nav", ".snm",
	".synctex.gz", ".fls", ".fdb_latexmk", ".bbl", ".blg", ".run.xml",
	".bcf", ".idx", ".ilg", ".ind", ".brf", ".vrb", ".xdv", ".dvi",
}

// Default Exclusion Patterns for File Watching
var DefaultExclusionPatterns = []string{
	"*.aux", "*.log", "*.out", "*.toc",
	"*.lof", "*.lot", "*.nav", "*.snm",
	"*.vrb", "*.fls", "*.fdb_latexmk",
	"*.synctex.gz", "*.bbl", "*.blg",
	"*.run.xml", "*.bcf", "*.idx",
	"*.ilg", "*.ind", "*.brf", "*.xdv", "*.dvi",
}

// Verbose Level Descriptions
var VerboseLevelDescriptions = map[int]string{
	0: "Silent (only errors)",
	1: "Basic information (warnings and above)",
	2: "Detailed information (info and above)",
	3: "Debug information (debug and above)",
	4: "Maximum verbosity (all logs with full introspection)",
}

// Default Debug Output
const (
	DefaultDebugOutput = "stdout"
	DefaultLogOutput   = "stdout"
)

// Error Messages
const (
	// Directory and file errors
	ErrDirectoryNotExists = "directory does not exist: %s"
	ErrConfigDirCreate    = "failed to create config directory"
	ErrConfigFileWrite    = "failed to write config file"

	// Service errors
	ErrServiceCreation  = "failed to create service"
	ErrServiceExecution = "service execution failed"
	ErrServiceCleanup   = "service cleanup failed"

	// Configuration errors
	ErrConfigLoad       = "failed to load configuration"
	ErrConfigSave       = "failed to save configuration"
	ErrConfigValidation = "configuration validation failed"

	// File operation errors
	ErrFileNotFound = "file not found: %s"
	ErrFileRead     = "failed to read file: %s"
	ErrFileWrite    = "failed to write file: %s"
	ErrFileDelete   = "failed to delete file: %s"

	// Template processing errors
	ErrTemplateProcess = "template processing failed"
	ErrTemplateCompile = "template compilation failed"
	ErrTemplateRender  = "template rendering failed"

	// LaTeX compilation errors
	ErrLaTeXCompile = "LaTeX compilation failed"
	ErrLaTeXEngine  = "LaTeX engine not found"
	ErrLaTeXOutput  = "LaTeX output generation failed"

	// Conversion errors
	ErrPDFConversion     = "PDF conversion failed"
	ErrImageGeneration   = "image generation failed"
	ErrFormatUnsupported = "unsupported format: %s"

	// Cleanup errors
	ErrAuxCleanup  = "auxiliary file cleanup failed"
	ErrTempCleanup = "temporary file cleanup failed"
)

// Default Values
const (
	// Logging defaults
	DefaultVerboseLevel = 2 // Detailed information
	DefaultLogLevel     = "info"

	// File watching defaults
	DefaultWatchInterval = "1s"
	DefaultDebounceTime  = "500ms"

	// Parallel processing defaults
	DefaultMaxConcurrency = 4
	DefaultTimeout        = "30s"

	// Cleanup defaults
	DefaultCleanEnabled = false
	DefaultForceEnabled = false
	DefaultDebugEnabled = false
)

// Path Constants
const (
	// Home directory fallback
	TempDirFallback = "/tmp"

	// Config file permissions
	ConfigDirPerms  = 0755
	ConfigFilePerms = 0644
)

// Service Names
const (
	ServiceDocument   = "document"
	ServiceOptions    = "options"
	ServiceParallel   = "parallel"
	ServicePersistent = "persistent"
	ServiceWatch      = "watch"
	ServiceClean      = "clean"
	ServiceVerbose    = "verbose"
	ServiceDebug      = "debug"
	ServiceForce      = "force"
	ServiceConvert    = "convert"
	ServiceConfig     = "config"
)

// Log Level Names
const (
	LogLevelSilent   = "silent"
	LogLevelBasic    = "basic"
	LogLevelDetailed = "detailed"
	LogLevelDebug    = "debug"
	LogLevelMaximum  = "maximum"
)

// Output Format Names
const (
	OutputFormatStdout = "stdout"
	OutputFormatStderr = "stderr"
	OutputFormatBoth   = "both"
	OutputFormatFile   = "file"
)

// File Type Extensions
const (
	ExtensionTex   = ".tex"
	ExtensionPDF   = ".pdf"
	ExtensionPNG   = ".png"
	ExtensionJPEG  = ".jpg"
	ExtensionJPEG2 = ".jpeg"
	ExtensionSVG   = ".svg"
	ExtensionYAML  = ".yaml"
	ExtensionYML   = ".yml"
	ExtensionJSON  = ".json"
)

// Template Delimiters
const (
	TemplateStartDelim = "delim[["
	TemplateEndDelim   = "]]"
)

// Default Configuration Values
const (
	DefaultEngine       = "pdflatex"
	DefaultOutputFormat = "pdf"
	DefaultTemplateName = "template.tex"
	DefaultConfigName   = "config.yaml"
	DefaultOutputName   = "output.pdf"
)

// Error Categories
const (
	ErrorCategoryConfig     = "configuration"
	ErrorCategoryFile       = "file_operation"
	ErrorCategoryTemplate   = "template"
	ErrorCategoryLaTeX      = "latex"
	ErrorCategoryConversion = "conversion"
	ErrorCategoryCleanup    = "cleanup"
	ErrorCategoryService    = "service"
	ErrorCategoryNetwork    = "network"
	ErrorCategorySystem     = "system"
)

// Status Messages
const (
	StatusSuccess    = "success"
	StatusError      = "error"
	StatusWarning    = "warning"
	StatusInfo       = "info"
	StatusDebug      = "debug"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusSkipped    = "skipped"
	StatusPending    = "pending"
)

// Default Messages
const (
	MsgConfigLoaded       = "Configuration loaded successfully"
	MsgConfigSaved        = "Configuration saved successfully"
	MsgServiceStarted     = "Service started successfully"
	MsgServiceStopped     = "Service stopped successfully"
	MsgFileProcessed      = "File processed successfully"
	MsgFileSkipped        = "File skipped"
	MsgFileError          = "File processing error"
	MsgCleanupComplete    = "Cleanup completed successfully"
	MsgBuildComplete      = "Build completed successfully"
	MsgBuildFailed        = "Build failed"
	MsgConversionComplete = "Conversion completed successfully"
	MsgConversionFailed   = "Conversion failed"
)
