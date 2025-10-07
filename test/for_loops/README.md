# For Loops Example

This example demonstrates the powerful for loop functionality in AutoPDF templates using Go template range syntax.

## What This Example Shows

- **Range loops over arrays** of strings and objects
- **Nested loop structures** with multiple levels
- **Object property access** within loops
- **Dynamic content generation** based on data structure
- **Mixed loop types** (simple arrays, object arrays, nested structures)

## Files

- `config.yaml`: Complex YAML with arrays and nested structures
- `loops_document.tex`: LaTeX template with extensive range loop usage
- `README.md`: This documentation

## Running the Example

```bash
cd test/for_loops
autopdf build loops_document.tex config.yaml
```

## Expected Output

- `loops_document.pdf`: Generated PDF with dynamic content from loops

## Key Features Demonstrated

- ✅ **Array Loops**: `delim[[range .complex.tags]]`
- ✅ **Object Loops**: `delim[[range .complex.features]]`
- ✅ **Nested Loops**: Multiple levels of iteration
- ✅ **Property Access**: `.name`, `.description`, `.enabled`
- ✅ **Dynamic Content**: Content generated from data structure
- ✅ **Mixed Data Types**: Strings, booleans, numbers in loops

## Loop Syntax Examples

### Simple Array Loop
```latex
delim[[range .complex.tags]]
delim[[.]]\par
delim[[end]]
```

### Object Array Loop
```latex
delim[[range .complex.features]]
\subsection{delim[[.name]]}
Description: delim[[.description]]
Enabled: delim[[.enabled]]
delim[[end]]
```

### Nested Structure Loop
```latex
delim[[range .complex.sections]]
\section{delim[[.title]]}
delim[[.content]]
delim[[range .subsections]]
\item delim[[.]]
delim[[end]]
delim[[end]]
```

## Data Structure

The example uses a complex structure with multiple array types:

```yaml
variables:
  # Simple array
  tags: ["templates", "loops", "variables", "automation"]
  
  # Object array
  features:
    - name: "Range Loops"
      description: "Iterate over arrays and objects"
      enabled: true
      priority: 1
  
  # Nested structure with arrays
  sections:
    - title: "Introduction"
      content: "This document demonstrates for loops"
      subsections: ["Getting Started", "Basic Syntax"]
```

## Template Access Methods

AutoPDF provides two ways to access variables:

1. **Flattened Variables** (`.vars.*`): For direct access
   - `delim[[.vars.title]]`
   - `delim[[.vars.settings.theme]]`

2. **Complex Structure** (`.complex.*`): For range loops
   - `delim[[range .complex.tags]]`
   - `delim[[range .complex.features]]`

## Best Practices

- Use `.vars.*` for simple variable substitution
- Use `.complex.*` for range loops and complex operations
- Combine both approaches for maximum flexibility
- Test templates with different data structures
- Use meaningful variable names for better readability
