# Complex Variables Test

This test demonstrates the complex variables functionality of AutoPDF, showing how the CLI can handle nested YAML structures with arrays, objects, and mixed data types.

## What This Test Demonstrates

### 1. Complex YAML Parsing
The CLI successfully parses YAML configuration files with:
- **Nested objects**: `metadata.settings.verbose`, `metadata.settings.debug`
- **Arrays**: `metadata.tags[0]`, `metadata.tags[1]`, `metadata.tags[2]`, `metadata.tags[3]`
- **Nested objects in arrays**: `items[0].name`, `items[0].enabled`, `items[0].priority`
- **Mixed data types**: strings, booleans, numbers

### 2. Variable Flattening
The complex nested structure is automatically flattened into 32 individual key-value pairs:
```
metadata.tags[0]: example
metadata.tags[1]: complex
metadata.tags[2]: variables
metadata.tags[3]: test
metadata.settings.verbose: true
metadata.settings.debug: false
metadata.settings.timeout: 30
items[0].name: Feature 1
items[0].enabled: true
items[0].priority: 1
items[0].description: Basic functionality
... and 22 more variables
```

### 3. Template Processing
The template processor successfully substitutes variables using the `delim[[.variable]]` syntax:
- Basic variables: `delim[[.title]]`, `delim[[.author]]`, `delim[[.date]]`
- Nested variables: `delim[[.metadata.version]]`, `delim[[.metadata.settings.verbose]]`
- Configuration variables: `delim[[.configuration.engine]]`, `delim[[.configuration.output_format]]`

### 4. PDF Generation
The complete workflow works:
1. Parse YAML config with complex variables
2. Flatten variables to simple key-value pairs
3. Process template with variable substitution
4. Compile LaTeX to PDF
5. Generate output.pdf (55KB)

## Files

- `config.yaml`: Complex YAML configuration with nested structures
- `simple_template.tex`: LaTeX template using complex variables
- `output.pdf`: Generated PDF demonstrating the functionality
- `README.md`: This documentation

## Running the Test

```bash
cd test/complex_variables
go run ../../cmd/autopdf/main.go build simple_template.tex config.yaml
```

## Expected Output

The CLI should successfully:
1. Parse the complex YAML configuration
2. Flatten 32 variables from the nested structure
3. Process the template with variable substitution
4. Generate a PDF file

## Key Features Demonstrated

- ✅ **YAML Parsing**: Complex nested structures
- ✅ **Variable Flattening**: Arrays and objects to key-value pairs
- ✅ **Template Processing**: Variable substitution in LaTeX
- ✅ **PDF Generation**: Complete workflow
- ✅ **CLI Integration**: Command-line interface support
- ✅ **Logging**: Detailed logging of the process

This test proves that the CLI fully supports complex nested variables from YAML configuration files, making it possible to use sophisticated data structures in AutoPDF templates.
