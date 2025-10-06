package examples

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/internal/template"
)

func main() {
	// Create a temporary directory for the example
	tempDir, err := os.MkdirTemp("", "autopdf-enhanced-example")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an enhanced template that demonstrates complex data structures
	templateContent := `
\documentclass[12pt,a4paper]{article}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{graphicx}
\usepackage{booktabs}
\usepackage{longtable}
\usepackage{enumitem}
\usepackage{hyperref}
\usepackage{geometry}

\geometry{margin=1in}

\title{delim[[.title]]}
\author{delim[[.author]]}
\date{delim[[.date]]}

\begin{document}

\maketitle

\section{Company Information}

\textbf{Company Name:} delim[[.company.name]]

\textbf{Address:}
\begin{itemize}
    \item Street: delim[[.company.address.street]]
    \item City: delim[[.company.address.city]]
    \item State: delim[[.company.address.state]]
    \item ZIP: delim[[.company.address.zip]]
    \item Country: delim[[.company.address.country]]
\end{itemize}

\textbf{Contact Information:}
\begin{itemize}
    \item Phone: delim[[.company.contact.phone]]
    \item Email: delim[[.company.contact.email]]
    \item Website: delim[[.company.contact.website]]
\end{itemize}

\section{Team Members}

delim[[range .team]]
\subsection{delim[[.name]]}
\begin{itemize}
    \item \textbf{Role:} delim[[.role]]
    \item \textbf{Experience:} delim[[.experience]] years
    \item \textbf{Skills:} delim[[join ", " .skills]]
\end{itemize}
delim[[end]]

\section{Projects}

delim[[range .projects]]
\subsection{delim[[.name]]}
\textbf{Description:} delim[[.description]]

\textbf{Technologies:} delim[[join ", " .technologies]]

\textbf{Status:} delim[[.status]]

\textbf{Team Members:}
\begin{itemize}
delim[[range .team_members]]
    \item delim[[.]]
delim[[end]]
\end{itemize}
delim[[end]]

\section{Financial Information}

\textbf{Total Budget:} \$delim[[.financial.total_budget]]

\textbf{Breakdown:}
\begin{itemize}
delim[[range .financial.categories]]
    \item delim[[.name]]: \$delim[[.amount]] (delim[[.percentage]]\%)
delim[[end]]
\end{itemize}

\end{document}
`

	// Write the template to a file
	templatePath := filepath.Join(tempDir, "enhanced-template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		log.Fatalf("Failed to write template: %v", err)
	}

	// Create enhanced engine configuration
	config := &template.EnhancedConfig{
		TemplatePath: templatePath,
		OutputPath:   filepath.Join(tempDir, "output.tex"),
		Engine:       "xelatex",
		Delimiters: template.DelimiterConfig{
			Left:  "delim[[",
			Right: "]]",
		},
		Functions: make(map[string]interface{}),
	}

	// Create the enhanced engine
	engine := template.NewEnhancedEngine(config)

	// Set complex variables with nested structures
	variables := map[string]interface{}{
		"title":  "Enhanced AutoPDF Template Example",
		"author": "AutoPDF Team",
		"date":   "2024-01-15",

		"company": map[string]interface{}{
			"name": "AutoPDF Solutions Inc.",
			"address": map[string]interface{}{
				"street":  "123 Technology Drive",
				"city":    "San Francisco",
				"state":   "CA",
				"zip":     "94105",
				"country": "USA",
			},
			"contact": map[string]interface{}{
				"phone":   "+1-555-0123",
				"email":   "info@autopdf.com",
				"website": "https://autopdf.com",
			},
		},

		"team": []interface{}{
			map[string]interface{}{
				"name":       "John Doe",
				"role":       "Lead Developer",
				"experience": 5,
				"skills":     []interface{}{"Go", "LaTeX", "Templates", "Docker"},
			},
			map[string]interface{}{
				"name":       "Jane Smith",
				"role":       "Technical Writer",
				"experience": 3,
				"skills":     []interface{}{"Documentation", "LaTeX", "Markdown", "Git"},
			},
			map[string]interface{}{
				"name":       "Bob Johnson",
				"role":       "DevOps Engineer",
				"experience": 4,
				"skills":     []interface{}{"Docker", "Kubernetes", "CI/CD", "AWS"},
			},
		},

		"projects": []interface{}{
			map[string]interface{}{
				"name":         "AutoPDF Core",
				"description":  "Core PDF generation library with LaTeX support",
				"technologies": []interface{}{"Go", "LaTeX", "Templates"},
				"status":       "Active",
				"team_members": []interface{}{"John Doe", "Jane Smith"},
			},
			map[string]interface{}{
				"name":         "AutoPDF Web",
				"description":  "Web interface for AutoPDF",
				"technologies": []interface{}{"React", "Node.js", "PostgreSQL"},
				"status":       "Planning",
				"team_members": []interface{}{"Bob Johnson", "Jane Smith"},
			},
			map[string]interface{}{
				"name":         "AutoPDF CLI",
				"description":  "Command-line interface for AutoPDF",
				"technologies": []interface{}{"Go", "Cobra", "Viper"},
				"status":       "Beta",
				"team_members": []interface{}{"John Doe"},
			},
		},

		"financial": map[string]interface{}{
			"total_budget": 1000000,
			"categories": []interface{}{
				map[string]interface{}{
					"name":       "Development",
					"amount":     600000,
					"percentage": 60,
				},
				map[string]interface{}{
					"name":       "Infrastructure",
					"amount":     200000,
					"percentage": 20,
				},
				map[string]interface{}{
					"name":       "Marketing",
					"amount":     100000,
					"percentage": 10,
				},
				map[string]interface{}{
					"name":       "Operations",
					"amount":     100000,
					"percentage": 10,
				},
			},
		},
	}

	// Set variables in the engine
	if err := engine.SetVariablesFromMap(variables); err != nil {
		log.Fatalf("Failed to set variables: %v", err)
	}

	// Process the template
	result, err := engine.Process(templatePath)
	if err != nil {
		log.Fatalf("Template processing failed: %v", err)
	}

	// Write the result to a file
	outputPath := filepath.Join(tempDir, "enhanced-output.tex")
	if err := os.WriteFile(outputPath, []byte(result), 0644); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	fmt.Printf("Enhanced template processed successfully!\n")
	fmt.Printf("Template: %s\n", templatePath)
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Printf("Result length: %d characters\n", len(result))

	// Show a preview of the result
	fmt.Printf("\nPreview of the processed template:\n")
	fmt.Printf("=====================================\n")
	lines := []string{}
	for i, line := range []byte(result) {
		if i < 500 { // Show first 500 characters
			lines = append(lines, string(line))
		} else {
			break
		}
	}
	fmt.Printf("%s...\n", string([]byte(result)[:500]))
}
