package configs

import "fmt"

var DefaultConfigName = "autopdf.yaml"

var (
	ConfigFileExistsError = fmt.Errorf("config file already exists")
	ConfigFileWriteError  = fmt.Errorf("failed to write config file")
	BuildError            = fmt.Errorf("failed to build config")
	WriteError            = fmt.Errorf("failed to write config")
	ReadError             = fmt.Errorf("failed to read config")
	ParseError            = fmt.Errorf("failed to parse config")
	TemplateError         = fmt.Errorf("failed to process template")
	ConversionError       = fmt.Errorf("failed to convert PDF to images")
	CleanError            = fmt.Errorf("failed to clean up auxiliary files")
)
