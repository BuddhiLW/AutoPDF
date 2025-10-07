# AutoPDF Test Examples

This directory contains comprehensive examples demonstrating AutoPDF's capabilities, from basic usage to advanced features.

## 📁 **Test Examples Overview**

### 🚀 **Basic Usage** (`basic_usage/`)
- **Purpose**: Simple variable substitution and basic PDF generation
- **Features**: Basic YAML configuration, LaTeX compilation, PDF output
- **Files**: `document.tex`, `config.yaml`, `README.md`
- **Run**: `autopdf build document.tex config.yaml`

### 🔧 **Advanced Features** (`advanced_features/`)
- **Purpose**: Complex variables, nested structures, and range loops
- **Features**: Complex YAML, nested objects, arrays, mixed data types
- **Files**: `advanced_document.tex`, `config.yaml`, `README.md`
- **Run**: `autopdf build advanced_document.tex config.yaml`

### 🔄 **For Loops** (`for_loops/`)
- **Purpose**: Range loops over arrays and objects
- **Features**: Dynamic content generation, nested loops, object iteration
- **Files**: `loops_document.tex`, `config.yaml`, `README.md`
- **Run**: `autopdf build loops_document.tex config.yaml`

### ⚙️ **Persistent Settings** (`persistent_settings/`)
- **Purpose**: CLI settings management and configuration persistence
- **Features**: Persistent settings, configuration management, cross-session settings
- **Files**: `settings_document.tex`, `config.yaml`, `README.md`
- **Run**: `autopdf build settings_document.tex config.yaml`

### 📊 **Table Data Filling** (`table_example/`)
- **Purpose**: LaTeX table generation with complex data structures
- **Features**: Employee directories, department statistics, project portfolios, skills matrices
- **Files**: `table_document.tex`, `config.yaml`, `README.md`
- **Run**: `autopdf build table_document.tex config.yaml`

### 🏗️ **Legacy Examples**
- **`model_letter/`**: Letter document example
- **`model_xelatex/`**: XeLaTeX engine example
- **`complex_variables/`**: Complex variable demonstration
- **`test_migration/`**: Migration testing
- **`equivalence/`**: Equivalence testing
- **`contract/`**: Contract testing

## 🚀 **Quick Start Guide**

### 1. **Basic Example**
```bash
cd test/basic_usage
autopdf build document.tex config.yaml
```

### 2. **Advanced Features**
```bash
cd test/advanced_features
autopdf build advanced_document.tex config.yaml
```

### 3. **For Loops**
```bash
cd test/for_loops
autopdf build loops_document.tex config.yaml
```

### 4. **Persistent Settings**
```bash
cd test/persistent_settings
autopdf build settings_document.tex config.yaml
```

## 📋 **Example Features Matrix**

E.g., what is covered in each example, in `test/` subfolders.

| Example | Basic Variables | Complex Variables | Range Loops | PDF Conversion | Persistent Settings | Table Generation |
|---------|:---------------:|:-----------------:|:-----------:|:--------------:|:------------------:|:----------------:|
| `basic_usage` | ✅ | ❌ | ❌ |  ❌  | ❌ | ❌ |
| `advanced_features` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `for_loops` | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| `persistent_settings` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ |
| `table_example` | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ |

## 🔧 **Configuration Examples**

### Basic Configuration
```yaml
template: "document.tex"
output: "output.pdf"
engine: "pdflatex"
variables:
  title: "My Document"
  author: "AutoPDF User"
  date: "2025-01-07"
```

### Complex Configuration
```yaml
template: "advanced_document.tex"
output: "advanced_output.pdf"
engine: "pdflatex"
variables:
  title: "Advanced Document"
  metadata:
    version: "1.0.0"
    tags: ["example", "advanced", "complex"]
    settings:
      verbose: true
      debug: false
  items:
    - name: "Feature 1"
      enabled: true
      priority: 1
    - name: "Feature 2"
      enabled: false
      priority: 2
conversion:
  enabled: true
  formats: ["png", "jpeg"]
```

## 📝 **Template Syntax Examples**

### Basic Variables
```latex
\title{delim[[.vars.title]]}
\author{delim[[.vars.author]]}
\date{delim[[.vars.date]]}
```

### Complex Variables
```latex
% Direct access to nested properties
Version: delim[[.vars.metadata.version]]
Verbose: delim[[.vars.metadata.settings.verbose]]

% Range loops over arrays
delim[[range .complex.metadata.tags]]
delim[[.]]\par
delim[[end]]

% Range loops over objects
delim[[range .complex.items]]
\subsection{delim[[.name]]}
Enabled: delim[[.enabled]]
Priority: delim[[.priority]]
delim[[end]]
```

## 🛠️ **CLI Commands**

### Build Commands
```bash
# Basic build
autopdf build TEMPLATE [CONFIG]

# With options
autopdf build TEMPLATE [CONFIG] [OPTIONS]
```

### Setting Commands
```bash
# Verbose settings
autopdf verbose [LEVEL|on|off]

# Clean settings
autopdf clean [on|off|switch|status]

# Debug settings
autopdf debug [on|off|switch]

# Force settings
autopdf force [on|off|switch]
```

## 📊 **Expected Outputs**

Each example generates:
- **PDF Document**: Main output file
- **Image Files**: PNG/JPEG conversions (if enabled)
- **Console Logs**: Structured logging with zap
- **Auxiliary Files**: LaTeX auxiliary files (if not cleaned)

## 🔍 **Testing and Validation**

### Manual Testing
1. Run each example individually
2. Verify PDF generation
3. Check console output for errors
4. Validate image conversion (if enabled)

### Automated Testing
```bash
# Run all examples
for dir in test/*/; do
  if [ -f "$dir/config.yaml" ]; then
    echo "Testing $dir"
    cd "$dir"
    autopdf build *.tex config.yaml
    cd - > /dev/null
  fi
done
```

## 📚 **Documentation**

Each example directory contains:
- **`README.md`**: Detailed documentation
- **`config.yaml`**: Configuration file
- **`*.tex`**: LaTeX template
- **Generated files**: PDF and image outputs

## 🎯 **Learning Path**

1. **Start with `basic_usage/`** - Learn fundamental concepts
2. **Progress to `advanced_features/`** - Explore complex variables
3. **Try `for_loops/`** - Master range loops and dynamic content
4. **Experiment with `persistent_settings/`** - Understand CLI management
5. **Explore legacy examples** - See real-world applications

## 🤝 **Contributing**

To add new examples:
1. Create a new directory in `test/`
2. Add `config.yaml`, `*.tex`, and `README.md`
3. Test the example thoroughly
4. Update this index
5. Submit a pull request

## 📄 **License**

All examples are provided under the same license as AutoPDF: [Apache License 2.0](../LICENSE).
