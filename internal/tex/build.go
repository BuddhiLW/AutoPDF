package tex

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/internal/config"
	"github.com/BuddhiLW/AutoPDF/internal/converter"
	"github.com/BuddhiLW/AutoPDF/internal/template"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

var BuildCmd = &bonzai.Cmd{
	Name:    `build`,
	Alias:   `b`,
	Short:   `process template and compile to PDF`,
	Usage:   `TEMPLATE [CONFIG]`,
	MinArgs: 1,
	MaxArgs: 2,
	Long: `
The build command processes a template file using variables from a configuration,
compiles the processed template to LaTeX, and produces a PDF output.

If no configuration file is provided, it will look for autopdf.yaml in the current directory.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		templateFile := args[0]
		configFile := "autopdf.yaml"
		if len(args) > 1 {
			configFile = args[1]
		}

		// Read the YAML config
		yamlData, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		// Parse the YAML config
		cfg, err := config.NewConfigFromYAML(yamlData)
		if err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		// If template not set in config, use the provided one
		if cfg.Template == "" {
			cfg.Template = config.Template(templateFile)
		}

		// Process the template
		engine := template.NewEngine(cfg)
		// write temporarly in the working directory
		tempDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		processedTexFile := filepath.Join(tempDir, "autopdf_"+filepath.Base(cfg.Template.String()))

		result, err := engine.Process(cfg.Template.String())
		if err != nil {
			return fmt.Errorf("template processing failed: %w", err)
		}

		// Write processed template to temp file
		if err := os.WriteFile(processedTexFile, []byte(result), 0644); err != nil {
			return fmt.Errorf("failed to write processed template: %w", err)
		}

		// Compile the LaTeX to PDF
		compiler := NewCompiler(cfg)
		outputPDF, err := compiler.Compile(processedTexFile)
		if err != nil {
			return fmt.Errorf("LaTeX compilation failed: %w", err)
		}

		// If output path is specified in config, move the PDF there
		if cfg.Output != "" && outputPDF != cfg.Output.String() {
			outputDir := filepath.Dir(cfg.Output.String())
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			pdfData, err := os.ReadFile(outputPDF)
			if err != nil {
				return fmt.Errorf("failed to read compiled PDF: %w", err)
			}

			if err := os.WriteFile(cfg.Output.String(), pdfData, 0644); err != nil {
				return fmt.Errorf("failed to write to output location: %w", err)
			}

			outputPDF = cfg.Output.String()
		}

		// Clean up the temp file
		os.Remove(processedTexFile)

		// Handle conversion if enabled
		if cfg.Conversion.Enabled {
			conv := converter.NewConverter(cfg)
			imageFiles, err := conv.ConvertPDFToImages(outputPDF)
			if err != nil {
				return fmt.Errorf("PDF to image conversion failed: %w", err)
			}

			if len(imageFiles) > 0 {
				fmt.Println("Generated image files:")
				for _, file := range imageFiles {
					fmt.Printf("  - %s\n", file)
				}
			}
		}

		fmt.Printf("Successfully built PDF: %s\n", outputPDF)
		return nil
	},
}
