# Table Data Filling Example

This example demonstrates how to fill LaTeX tables with data using AutoPDF's complex variables and for loops, showcasing various table types and data structures.

## What This Example Shows

- **Employee Directory**: Long table with employee information
- **Department Statistics**: Simple table with department data
- **Project Portfolio**: Complex table with project details
- **Skills Matrix**: Employee skills mapping
- **Department Skills**: Skills by department
- **Project Technologies**: Technology stack per project
- **Skills Inventory**: Complete skills list
- **Salary Analysis**: Financial data tables
- **Status Summaries**: Statistical tables

## Files

- `config.yaml`: Complex YAML with employee, department, and project data
- `table_document.tex`: LaTeX template with various table types
- `README.md`: This documentation

## Running the Example

```bash
cd test/table_example
autopdf build table_document.tex config.yaml
```

## Expected Output

- `table_document.pdf`: Generated PDF with multiple table types
- Console output showing successful compilation

## Key Features Demonstrated

- ✅ **Long Tables**: Multi-page tables with `longtable` package
- ✅ **Simple Tables**: Basic tabular data with headers
- ✅ **Complex Data**: Nested objects and arrays
- ✅ **Conditional Formatting**: Status indicators and conditional content
- ✅ **Number Formatting**: Currency and number formatting
- ✅ **Range Loops**: Dynamic content generation from arrays
- ✅ **Nested Loops**: Complex data structures with multiple levels
- ✅ **Professional Formatting**: Booktabs and color formatting
- ✅ **YAML Data Integration**: All tables filled with YAML configuration data
- ✅ **Dynamic Table Generation**: Tables adapt to data structure changes

## Table Types Demonstrated

### 1. Employee Directory (Long Table)
```latex
\begin{longtable}{|p{3cm}|p{3cm}|p{2.5cm}|p{2cm}|p{2cm}|p{2cm}|}
\hline
\textbf{Name} & \textbf{Position} & \textbf{Department} & \textbf{Salary} & \textbf{Experience} & \textbf{Status} \\
\hline
\endhead

delim[[range .complex.employees]]
delim[[.name]] & delim[[.position]] & delim[[.department]] & \$${delim[[.salary]]:,} & delim[[.experience]] years & delim[[if .active]]Active\delim[[else]]Inactive\delim[[end]] \\
\hline
delim[[end]]
```

### 2. Department Statistics (Simple Table)
```latex
\begin{tabular}{|l|c|c|}
\hline
\textbf{Department} & \textbf{Head Count} & \textbf{Average Salary} \\
\hline
delim[[range .complex.departments]]
delim[[.name]] & delim[[.head_count]] & \$${delim[[.avg_salary]]:,} \\
\hline
delim[[end]]
\end{tabular}
```

### 3. Skills Matrix (Nested Loops)
```latex
delim[[range .complex.employees]]
delim[[.name]] & delim[[range .skills]]delim[[.]]\delim[[if not @last]], \delim[[end]]delim[[end]] \\
\hline
delim[[end]]
```

## Data Structure

The example uses a comprehensive data structure with all tables filled from YAML data:

```yaml
variables:
  # Employee data
  employees:
    - name: "John Doe"
      position: "Software Engineer"
      department: "Engineering"
      salary: 75000
      experience: 3
      skills: ["Go", "Python", "JavaScript"]
      active: true
  
  # Department data with salary analysis
  departments:
    - name: "Engineering"
      head_count: 2
      avg_salary: 77500
      min_salary: 75000
      max_salary: 80000
      skills: ["Go", "Python", "JavaScript"]
  
  # Project status summary
  project_status:
    - status: "Active"
      count: 1
      percentage: 33.3
  
  # Experience level distribution
  experience_levels:
    - range: "1-2 years"
      count: 1
```

## Template Syntax Examples

### Basic Range Loop
```latex
delim[[range .complex.employees]]
delim[[.name]] & delim[[.position]] \\
\hline
delim[[end]]
```

### Conditional Content
```latex
delim[[if .active]]Active\delim[[else]]Inactive\delim[[end]]
```

### Nested Loops
```latex
delim[[range .complex.employees]]
delim[[.name]]: delim[[range .skills]]
delim[[.]]\delim[[if not @last]], \delim[[end]]
delim[[end]]
delim[[end]]
```

### Number Formatting
```latex
\$${delim[[.salary]]:,}  % Currency formatting
delim[[.experience]] years  % Text concatenation
```

## LaTeX Packages Used

- `booktabs`: Professional table formatting
- `longtable`: Multi-page tables
- `array`: Enhanced column specifications
- `colortbl`: Table coloring
- `xcolor`: Color definitions

## Table Features

### Column Specifications
- `|p{3cm}|`: Fixed-width paragraph columns
- `|l|`: Left-aligned columns
- `|c|`: Center-aligned columns
- `|r|`: Right-aligned columns

### Table Types
- **Simple Tables**: `tabular` environment
- **Long Tables**: `longtable` environment for multi-page content
- **Floating Tables**: `table` environment with captions

### Formatting Options
- **Headers**: Bold text with `\textbf{}`
- **Borders**: Horizontal and vertical lines with `\hline` and `|`
- **Spacing**: Column padding and row spacing
- **Alignment**: Text alignment within cells

## Best Practices

### Data Organization
- Structure data hierarchically for better template readability
- Use meaningful variable names
- Group related data together
- Consider data size for table formatting

### Template Design
- Use appropriate table environments for content length
- Plan column widths based on content
- Use consistent formatting across tables
- Test with different data sizes

### LaTeX Considerations
- Use `longtable` for tables that might span multiple pages
- Consider column width constraints
- Use `booktabs` for professional appearance
- Format numbers and currency appropriately

## Advanced Features

### Conditional Formatting
```latex
delim[[if .active]]
\textcolor{green}{Active}
\delim[[else]]
\textcolor{red}{Inactive}
\delim[[end]]
```

### Complex Data Access
```latex
delim[[.vars.company.total_employees]]  % Direct variable access
delim[[range .complex.employees]]      % Range loop access
```

### Number Formatting
```latex
\$${delim[[.salary]]:,}  % Currency with thousands separator
delim[[.experience]] years  % Text concatenation
```

This example provides a comprehensive demonstration of table data filling with AutoPDF, covering various table types, data structures, and formatting options.
