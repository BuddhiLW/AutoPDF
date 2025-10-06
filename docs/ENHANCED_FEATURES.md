# AutoPDF Enhanced Features

## Overview

AutoPDF has been enhanced to support complex nested data structures, making it suitable for generating sophisticated documents like legal contracts, technical reports, and business documents.

## Key Enhancements

### 1. Complex Variable Types

AutoPDF now supports the following variable types:

- **String**: Text values
- **Number**: Numeric values (int, float64)
- **Boolean**: True/false values
- **Array**: Lists of values
- **Object**: Nested key-value structures
- **Null**: Empty values

### 2. Nested Data Structures

Support for deeply nested objects using dot notation:

```yaml
variables:
  company:
    name: "AutoPDF Solutions"
    address:
      street: "123 Technology Drive"
      city: "San Francisco"
      state: "CA"
      zip: "94105"
    contact:
      phone: "+1-555-0123"
      email: "info@autopdf.com"
```

Template access:
```latex
Company: delim[[.company.name]]
Address: delim[[.company.address.street]], delim[[.company.address.city]]
Phone: delim[[.company.contact.phone]]
```

### 3. Array Processing

Support for arrays with loops and indexing:

```yaml
variables:
  team:
    - name: "John Doe"
      role: "Developer"
      skills: ["Go", "LaTeX"]
    - name: "Jane Smith"
      role: "Writer"
      skills: ["Documentation", "Markdown"]
```

Template processing:
```latex
delim[[range .team]]
\subsection{delim[[.name]]}
Role: delim[[.role]]
Skills: delim[[join ", " .skills]]
delim[[end]]
```

### 4. Enhanced Template Functions

Built-in functions for template processing:

- `len(arr)`: Get array length
- `upper(str)`: Convert to uppercase
- `lower(str)`: Convert to lowercase
- `title(str)`: Title case
- `join(sep, arr)`: Join array with separator
- `range(arr)`: Iterate over array
- `index(arr, i)`: Get array element by index
- `keys(obj)`: Get object keys
- `values(obj)`: Get object values

### 5. Conditional Processing

Support for conditional sections:

```latex
delim[[if .has_team]]
\section{Team Members}
delim[[range .team]]
- delim[[.name]]
delim[[end]]
delim[[end]]
```

## Usage Examples

### Basic Enhanced Template

```go
package main

import (
    "github.com/BuddhiLW/AutoPDF/internal/template"
)

func main() {
    config := &template.EnhancedConfig{
        TemplatePath: "document.tex",
        OutputPath:     "output.tex",
        Engine:         "xelatex",
        Delimiters: template.DelimiterConfig{
            Left:  "delim[[",
            Right: "]]",
        },
    }
    
    engine := template.NewEnhancedEngine(config)
    
    variables := map[string]interface{}{
        "title": "My Document",
        "author": "John Doe",
        "sections": []interface{}{
            map[string]interface{}{
                "title": "Introduction",
                "content": "This is the introduction...",
            },
            map[string]interface{}{
                "title": "Conclusion",
                "content": "This is the conclusion...",
            },
        },
    }
    
    engine.SetVariablesFromMap(variables)
    result, err := engine.Process("document.tex")
    // Handle result...
}
```

### Complex Legal Document

```yaml
# config.yaml
template: "legal-contract.tex"
output: "contract.pdf"
engine: "xelatex"
variables:
  contract:
    title: "Software License Agreement"
    parties:
      licensor:
        name: "AutoPDF Solutions Inc."
        address: "123 Tech Drive, San Francisco, CA"
      licensee:
        name: "Client Corporation"
        address: "456 Business Ave, New York, NY"
    terms:
      duration: "5 years"
      territory: "United States"
      restrictions:
        - "No reverse engineering"
        - "No redistribution"
        - "No modification"
    financial:
      license_fee: 50000
      currency: "USD"
      payment_terms: "Annual"
```

Template:
```latex
\documentclass[12pt]{article}
\title{delim[[.contract.title]]}

\begin{document}
\maketitle

\section{Parties}
\textbf{Licensor:} delim[[.contract.parties.licensor.name]]
Address: delim[[.contract.parties.licensor.address]]

\textbf{Licensee:} delim[[.contract.parties.licensee.name]]
Address: delim[[.contract.parties.licensee.address]]

\section{Terms}
Duration: delim[[.contract.terms.duration]]
Territory: delim[[.contract.terms.territory]]

\subsection{Restrictions}
delim[[range .contract.terms.restrictions]]
\item delim[[.]]
delim[[end]]

\section{Financial Terms}
License Fee: \$delim[[.contract.financial.license_fee]] delim[[.contract.financial.currency]]
Payment Terms: delim[[.contract.financial.payment_terms]]

\end{document}
```

## Architecture

### Domain-Driven Design

The enhanced AutoPDF follows DDD principles:

- **Domain Layer**: Core business logic (`pkg/domain/`)
- **Application Layer**: Use cases and services (`internal/template/`)
- **Infrastructure Layer**: External dependencies and adapters

### SOLID Principles

1. **Single Responsibility**: Each component has one clear purpose
2. **Open/Closed**: Extensible through interfaces and composition
3. **Liskov Substitution**: Implementations are interchangeable
4. **Interface Segregation**: Small, focused interfaces
5. **Dependency Inversion**: Depend on abstractions, not concretions

### Key Components

#### Variable System
- `Variable`: Represents a typed value
- `VariableCollection`: Manages variable storage and retrieval
- `TemplateContext`: Provides template processing context

#### Enhanced Engine
- `EnhancedEngine`: Main template processing engine
- `EnhancedConfig`: Configuration for the enhanced engine
- Support for complex data structures and nested access

## Testing

Comprehensive test coverage following TDD principles:

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/domain/... -v
go test ./internal/template/... -v

# Run with coverage
go test ./... -cover
```

## Migration Guide

### From Simple Variables to Complex Structures

**Before:**
```yaml
variables:
  title: "Document Title"
  author: "John Doe"
  content: "Document content..."
```

**After:**
```yaml
variables:
  document:
    title: "Document Title"
    author: "John Doe"
    content: "Document content..."
    metadata:
      created: "2024-01-15"
      version: "1.0"
      tags: ["important", "draft"]
```

### Template Updates

**Before:**
```latex
\title{delim[[.title]]}
\author{delim[[.author]]}
```

**After:**
```latex
\title{delim[[.document.title]]}
\author{delim[[.document.author]]}
\date{delim[[.document.metadata.created]]}
```

## Performance Considerations

- **Memory Usage**: Complex structures use more memory
- **Processing Time**: Nested access has slight overhead
- **Template Size**: Larger templates may take longer to process

## Best Practices

1. **Structure Data Logically**: Group related data in objects
2. **Use Meaningful Names**: Clear, descriptive variable names
3. **Avoid Deep Nesting**: Keep nesting levels reasonable (3-4 levels max)
4. **Validate Input**: Check required variables before processing
5. **Handle Errors**: Proper error handling for missing variables

## Future Enhancements

- **Custom Functions**: User-defined template functions
- **Template Inheritance**: Base templates with extensions
- **Conditional Logic**: More sophisticated conditional processing
- **Data Validation**: Schema validation for variables
- **Performance Optimization**: Caching and optimization strategies

## Contributing

When contributing to the enhanced AutoPDF features:

1. Follow TDD principles
2. Maintain SOLID design principles
3. Add comprehensive tests
4. Update documentation
5. Consider backward compatibility

## Examples

See the `examples/` directory for complete working examples:

- `enhanced_template_example.go`: Basic enhanced template usage
- `legal_document_example.go`: Complex legal document generation
- `business_report_example.go`: Business report with charts and tables
