# AutoPDF Struct Converter

The AutoPDF Struct Converter provides automatic conversion from Go structs to AutoPDF `config.Variables` using reflection and struct tags. This enables seamless integration of complex Go objects into AutoPDF templates.

## Features

- **Automatic Conversion**: Convert any Go struct to AutoPDF Variables using reflection
- **Struct Tags**: Fine-grained control using `autopdf` struct tags
- **Nested Structures**: Support for nested structs with flattening and inlining options
- **Built-in Type Support**: Automatic handling of `time.Time`, `time.Duration`, `url.URL`, and more
- **Custom Converters**: Register custom converters for specific types
- **Extensible**: Implement `AutoPDFFormattable` interface for custom conversion logic
- **Performance**: Efficient reflection-based conversion with caching

## Basic Usage

```go
package main

import (
    "time"
    "github.com/BuddhiLW/AutoPDF/pkg/converter"
)

type Document struct {
    Title     string    `autopdf:"title"`
    Author    string    `autopdf:"author"`
    CreatedAt time.Time `autopdf:"created_at"`
    Pages     int       `autopdf:"pages"`
}

func main() {
    doc := Document{
        Title:     "My Document",
        Author:    "John Doe",
        CreatedAt: time.Now(),
        Pages:     42,
    }

    // Create converter with built-in type support
    converter := converter.BuildWithDefaults()
    
    // Convert struct to Variables
    variables, err := converter.ConvertStruct(doc)
    if err != nil {
        panic(err)
    }

    // Use variables in AutoPDF template
    // Template can access: {{.title}}, {{.author}}, {{.created_at}}, {{.pages}}
}
```

## Struct Tags

The `autopdf` struct tag provides fine-grained control over conversion behavior:

### Basic Tag Usage

```go
type User struct {
    Name  string `autopdf:"name"`           // Field name in template
    Email string `autopdf:"email"`          // Field name in template
    Age   int    `autopdf:"age"`            // Field name in template
    ID    int    `autopdf:"-"`              // Skip this field
}
```

### Tag Options

```go
type Advanced struct {
    // Omit empty values
    Name string `autopdf:"name,omitempty"`
    
    // Flatten nested structures (use dot notation)
    User User `autopdf:"user,flatten"`
    
    // Inline nested struct fields at parent level
    Profile User `autopdf:"profile,inline"`
    
    // Flatten slices to comma-separated strings
    Tags []string `autopdf:"tags,flatten"`
    
    // Multiple options
    Data User `autopdf:"data,flatten,omitempty"`
}
```

### Tag Format

- `autopdf:"field_name"` - Set field name in template
- `autopdf:"field_name,omitempty"` - Skip field if empty
- `autopdf:"field_name,flatten"` - Flatten nested structures
- `autopdf:"field_name,inline"` - Inline nested struct fields
- `autopdf:"-"` - Skip field entirely

## Advanced Features

### Nested Structures

#### Default Behavior (Preserves Structure)
```go
type Address struct {
    Street string `autopdf:"street"`
    City   string `autopdf:"city"`
}

type User struct {
    Name    string  `autopdf:"name"`
    Address Address `autopdf:"address"`
}

// Results in nested structure:
// address.street, address.city
```

#### Flattening
```go
type User struct {
    Name    string  `autopdf:"name"`
    Address Address `autopdf:"address,flatten"`
}

// Results in flattened structure:
// name, address.street, address.city
```

#### Inlining
```go
type User struct {
    Name    string  `autopdf:"name"`
    Address Address `autopdf:"address,inline"`
}

// Results in inlined structure:
// name, street, city
```

### Array and Slice Handling

#### Default Behavior (SliceVariable)
```go
type Document struct {
    Tags []string `autopdf:"tags"`
}

// Results in SliceVariable for template loops
```

#### Flattening to String
```go
type Document struct {
    Tags []string `autopdf:"tags,flatten"`
}

// Results in comma-separated string: "tag1, tag2, tag3"
```

### Built-in Type Support

The converter automatically handles common Go types:

```go
type Data struct {
    CreatedAt time.Time      `autopdf:"created_at"`     // RFC3339 format
    Duration  time.Duration  `autopdf:"duration"`       // String format
    Homepage  url.URL        `autopdf:"homepage"`       // URL string
    UpdatedAt *time.Time     `autopdf:"updated_at"`     // Pointer support
}
```

### Custom Type Conversion

#### Using AutoPDFFormattable Interface
```go
type CustomType struct {
    Value string
}

func (ct CustomType) ToAutoPDFVariable() (config.Variable, error) {
    return &config.StringVariable{Value: "custom:" + ct.Value}, nil
}

type Document struct {
    Custom CustomType `autopdf:"custom"`
}
```

#### Using Custom Converters
```go
type CustomConverter struct{}

func (cc CustomConverter) Convert(value interface{}) (config.Variable, error) {
    // Custom conversion logic
    return &config.StringVariable{Value: "converted"}, nil
}

func (cc CustomConverter) CanConvert(value interface{}) bool {
    _, ok := value.(MyCustomType)
    return ok
}

// Register converter
registry := converter.NewConverterRegistry()
registry.Register(reflect.TypeOf(MyCustomType{}), CustomConverter{})
```

## Converter Builders

### BuildWithDefaults
```go
converter := converter.BuildWithDefaults()
// Includes built-in converters with sensible defaults
```

### BuildForTemplates
```go
converter := converter.BuildForTemplates()
// Optimized for template usage with omitempty enabled
```

### BuildForFlattened
```go
converter := converter.BuildForFlattened()
// Flattens nested structures by default
```

### Custom Builder
```go
converter := converter.NewConverterBuilder().
    WithBuiltinConverters().
    WithTimeFormat("2006-01-02").
    WithDurationFormat("seconds").
    WithSliceSeparator(" | ").
    WithDefaultFlatten(true).
    WithOmitEmpty(true).
    Build()
```

## Integration with AutoPDF

### Using with Variable Resolver
```go
import (
    "github.com/BuddhiLW/AutoPDF/pkg/api/adapters/variable_resolver"
    "github.com/BuddhiLW/AutoPDF/pkg/converter"
)

// Create converter
converter := converter.BuildWithDefaults()

// Convert struct to Variables
variables, err := converter.ConvertStruct(myStruct)
if err != nil {
    return err
}

// Use with existing AutoPDF pipeline
resolver := variable_resolver.NewVariableResolverAdapter(config, logger)
resolved, err := resolver.Resolve(variables.GetVariables())
```

### Template Usage
```latex
% LaTeX template
\documentclass{article}
\begin{document}

\title{My Document}
\author{John Doe}
\date{2025-01-07}

\maketitle

% Access converted variables
Hello, my name is \textbf{John Doe} and I'm 30 years old.

% Access nested structures
My address is: \textbf{123 Main St, Anytown, USA}

% Access arrays (if not flattened)
\begin{itemize}
\foreach \tag in \tags{
    \item \tag
}
\end{itemize}

\end{document}
```

## Performance Considerations

- **Reflection Overhead**: The converter uses reflection, which has some overhead
- **Caching**: Consider caching converted results for frequently used structs
- **Memory**: Large structs with many nested fields may consume more memory
- **Type Safety**: Use custom converters for performance-critical types

## Error Handling

The converter provides detailed error messages for common issues:

```go
variables, err := converter.ConvertStruct(myStruct)
if err != nil {
    // Handle conversion errors
    log.Printf("Conversion failed: %v", err)
}
```

Common error scenarios:
- Invalid struct tags
- Unsupported types
- Circular references
- Conversion failures

## Best Practices

1. **Use Struct Tags**: Always use `autopdf` tags for better control
2. **Implement AutoPDFFormattable**: For custom types that need special handling
3. **Register Custom Converters**: For performance-critical or complex types
4. **Test Conversion**: Always test struct conversion with your actual data
5. **Use Appropriate Builders**: Choose the right builder for your use case
6. **Handle Errors**: Always check for conversion errors
7. **Consider Performance**: Cache results for frequently converted structs

## Examples

See `pkg/api/examples/struct_conversion_example.go` for comprehensive examples demonstrating all features.

## License

Copyright 2025 AutoPDF BuddhiLW
SPDX-License-Identifier: Apache-2.0

