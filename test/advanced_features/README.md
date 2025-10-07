# Advanced Features Example

This example demonstrates advanced AutoPDF functionality including complex variables, nested structures, and range loops.

## What This Example Shows

- **Complex nested variables** with unlimited depth
- **Range loops** over arrays and objects
- **Mixed data types** (strings, booleans, numbers)
- **Nested object access** with dot notation
- **Array indexing** with bracket notation
- **Both flattened and complex variable access**

## Files

- `config.yaml`: Complex YAML configuration with nested structures
- `advanced_document.tex`: LaTeX template with range loops and complex variables
- `README.md`: This documentation

## Running the Example

```bash
cd test/advanced_features
autopdf build advanced_document.tex config.yaml
```

## Expected Output

- `advanced_document.pdf`: Generated PDF with complex content
- `advanced_document.png`: Converted PNG image (if conversion enabled)
- `advanced_document.jpeg`: Converted JPEG image (if conversion enabled)

## Key Features Demonstrated

- ✅ **Complex Variables**: Nested objects and arrays
- ✅ **Range Loops**: Dynamic content generation
- ✅ **Mixed Data Types**: Strings, booleans, numbers
- ✅ **Nested Access**: `metadata.settings.verbose`
- ✅ **Array Access**: `items[0].name`, `items[1].enabled`
- ✅ **Template Processing**: Both `.vars.*` and `.complex.*` access
- ✅ **PDF Conversion**: Multiple output formats
- ✅ **Professional Formatting**: LaTeX document structure

## Variable Structure

The example uses a complex variable structure:
```yaml
variables:
  # Basic variables
  title: "Advanced AutoPDF Features"
  author: "AutoPDF Developer"
  
  # Nested objects
  metadata:
    version: "2.0.0"
    settings:
      verbose: true
      debug: false
      features: ["nested objects", "array access", "mixed types"]
  
  # Arrays of objects
  items:
    - name: "Feature 1"
      enabled: true
      priority: 1
    - name: "Feature 2"
      enabled: false
      priority: 2
```

## Template Syntax

The template uses both access methods:
- **Flattened variables**: `delim[[.vars.title]]`
- **Complex structure**: `delim[[range .complex.items]]`
- **Range loops**: `delim[[range .complex.metadata.tags]]`
- **Nested access**: `delim[[.vars.metadata.settings.verbose]]`
