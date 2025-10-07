# SOLID Template Path Solution

## Problem

The initial fix of using `templateFile` directly was a hack that violated SOLID principles. The proper approach is to ensure `cfg.Template` is set correctly and use it throughout the system.

## Root Cause Analysis

The issue was **path resolution context**:

1. **Config files** can contain relative template paths (e.g., `template: "./main.tex"`)
2. **Relative paths** in config files should be resolved relative to the **config file's directory**, not the current working directory
3. **Template adapter** should use the resolved path from the config, not the raw command-line argument

## SOLID Solution

### 1. Proper Path Resolution in Build Service

**File**: `internal/autopdf/commands/build_service.go`

```go
// If template not set in config, use the provided one
if cfg.Template == "" {
    // Ensure we use the absolute path to avoid path resolution issues
    absTemplatePath, err := filepath.Abs(templateFile)
    if err != nil {
        return fmt.Errorf("failed to resolve template path: %w", err)
    }
    cfg.Template = config.Template(absTemplatePath)
} else {
    // Template is set in config, but it might be relative
    // Resolve it relative to the config file's directory
    configDir := filepath.Dir(configFile)
    templatePath := cfg.Template.String()
    
    // If it's not already absolute, make it relative to config directory
    if !filepath.IsAbs(templatePath) {
        absTemplatePath := filepath.Join(configDir, templatePath)
        absTemplatePath, err = filepath.Abs(absTemplatePath)
        if err != nil {
            return fmt.Errorf("failed to resolve template path: %w", err)
        }
        cfg.Template = config.Template(absTemplatePath)
    }
}
```

### 2. Template Adapter Uses Config Properly

**File**: `internal/application/adapters/template_adapter.go`

```go
// Process processes a template with variables
func (tpa *TemplateProcessorAdapter) Process(ctx context.Context, templatePath string, variables map[string]string) (string, error) {
    // Create a config with the provided variables, but preserve the template from the original config
    cfg := &config.Config{
        Template:  tpa.config.Template, // Use the template from the original config
        Variables: config.Variables(variables),
        Engine:    tpa.config.Engine,
        Output:    tpa.config.Output,
    }

    // If no template is set in the config, use the provided templatePath
    if cfg.Template == "" {
        cfg.Template = config.Template(templatePath)
    }

    // Create the template engine
    engine := template.NewEngine(cfg)

    // Process the template using the template path from config
    configTemplatePath := cfg.Template.String()
    result, err := engine.Process(configTemplatePath)
    if err != nil {
        return "", err
    }

    return result, nil
}
```

### 3. Build Service Uses Config Template

**File**: `internal/autopdf/commands/build_service.go`

```go
// Build the request
req := application.BuildRequest{
    TemplatePath: cfg.Template.String(), // Use the template from config (properly resolved)
    ConfigPath:   configFile,
    Variables:    map[string]string(cfg.Variables),
    Engine:       cfg.Engine.String(),
    OutputPath:   cfg.Output.String(),
    DoConvert:    cfg.Conversion.Enabled,
    DoClean:      doClean,
    Conversion: application.ConversionSettings{
        Enabled: cfg.Conversion.Enabled,
        Formats: cfg.Conversion.Formats,
    },
}
```

## SOLID Principles Applied

### Single Responsibility Principle (SRP)
- **Build Service**: Handles path resolution and config management
- **Template Adapter**: Handles template processing with resolved paths
- **Template Engine**: Handles template parsing and variable substitution

### Open/Closed Principle (OCP)
- **Template Adapter**: Can be extended with new template processors without modification
- **Path Resolution**: Can be extended with new path resolution strategies

### Liskov Substitution Principle (LSP)
- **Template Adapter**: Implements `TemplateProcessor` interface correctly
- **All adapters**: Can be substituted without breaking functionality

### Interface Segregation Principle (ISP)
- **Port interfaces**: Small, focused interfaces for specific concerns
- **Template Processor**: Only handles template processing, not path resolution

### Dependency Inversion Principle (DIP)
- **Build Service**: Depends on abstractions (application service)
- **Application Service**: Depends on abstractions (port interfaces)
- **Adapters**: Implement abstractions, depend on concrete implementations

## Benefits of SOLID Approach

1. **Proper Separation of Concerns**: Path resolution is handled in the right place
2. **Config-Driven**: Template paths are resolved according to config file context
3. **Testable**: Each component can be tested in isolation
4. **Maintainable**: Clear responsibilities and dependencies
5. **Extensible**: Easy to add new path resolution strategies or template processors

## Test Results

### ✅ Template Processing with Relative Paths
```bash
$ ./autopdf build ./test/model_xelatex/main.tex ./test/model_xelatex/config.yaml
Building PDF using application service...
Successfully built PDF: ./out/output
Generated image files:
  - out/output.jpeg
```

### ✅ All Tests Pass
```bash
$ go test ./...
ok  	github.com/BuddhiLW/AutoPDF/internal/application	(cached)
ok  	github.com/BuddhiLW/AutoPDF/internal/autopdf	0.716s
ok  	github.com/BuddhiLW/AutoPDF/internal/converter	(cached)
ok  	github.com/BuddhiLW/AutoPDF/internal/template	(cached)
ok  	github.com/BuddhiLW/AutoPDF/internal/tex	(cached)
ok  	github.com/BuddhiLW/AutoPDF/pkg/api	(cached)
ok  	github.com/BuddhiLW/AutoPDF/pkg/config	(cached)
```

## Key Insights

1. **Context Matters**: Relative paths in config files should be resolved relative to the config file's directory
2. **SOLID Principles**: Proper separation of concerns leads to maintainable, testable code
3. **Config-Driven**: The system should respect the configuration hierarchy and context
4. **Path Resolution**: Should be handled at the appropriate layer (build service) not in adapters

## Status: ✅ SOLID SOLUTION IMPLEMENTED

The template path issue is now resolved using proper SOLID principles, with clear separation of concerns and maintainable code structure.
