package tex

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/configs"
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
	Usage:   `TEMPLATE [CONFIG] [CLEAN]`,
	MinArgs: 1,
	MaxArgs: 3,
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
		log.Println("Building PDF...")
		templateFile := args[0]
		configFile := configs.DefaultConfigName
		log.Println("Template file:", templateFile)
		log.Println("Config file:", configFile)
		if len(args) > 1 {
			log.Println("Using provided config file:", args[1])
			configFile = args[1]
		} else {
			log.Println("No config file provided, creating default config file...")
			err := Default(templateFile)
			if err != nil {
				return configs.BuildError
			}
			log.Println("Default config file written to:", configs.DefaultConfigName)
			configFile = configs.DefaultConfigName
		}

		yamlData, err := os.ReadFile(configFile)
		if err != nil {
			return configs.ReadError
		}

		// Parse the YAML config
		cfg, err := config.NewConfigFromYAML(yamlData)
		if err != nil {
			return configs.ParseError
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
			return configs.BuildError
		}
		processedTexFile := filepath.Join(tempDir, "autopdf_"+filepath.Base(cfg.Template.String()))

		result, err := engine.Process(cfg.Template.String())
		if err != nil {
			return configs.TemplateError
		}

		// Write processed template to temp file
		if err := os.WriteFile(processedTexFile, []byte(result), 0644); err != nil {
			return configs.WriteError
		}

		// Compile the LaTeX to PDF
		compiler := NewCompiler(cfg)
		outputPDF, err := compiler.Compile(processedTexFile)
		if err != nil {
			log.Printf("Error compiling: %s", err)
			return configs.BuildError
		}

		// If output path is specified in config, move the PDF there
		if cfg.Output != "" && outputPDF != cfg.Output.String() {
			outputDir := filepath.Dir(cfg.Output.String())
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return configs.BuildError
			}

			pdfData, err := os.ReadFile(outputPDF)
			if err != nil {
				return configs.ReadError
			}

			if err := os.WriteFile(cfg.Output.String(), pdfData, 0644); err != nil {
				return configs.WriteError
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
		if len(args) > 2 && args[2] == "clean" {

			// Get the directory of the input file
			dir := filepath.Dir(outputPDF)
			baseName := filepath.Base(outputPDF)

			// Determine output PDF path
			outputPDF := filepath.Join(dir, replaceExt(baseName, ".pdf"))
			if cfg.Output.String() != "" {
				outputPDF = cfg.Output.String()
			}

			// Create output directory, if it doesn't exist
			dirOutput := filepath.Dir(outputPDF)
			if err := CleanCmd.Do(cmd, dirOutput); err != nil {
				return configs.CleanError
			}
		}
		return nil
	},
}
