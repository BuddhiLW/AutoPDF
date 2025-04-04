package tex

import (
	"log"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/configs"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

func Default(texFilePath string) error {
	// create default config.yaml for autopdf
	config := config.Config{
		Template:  config.Template(texFilePath),
		Output:    config.Output(texFilePath),
		Engine:    config.Engine("pdflatex"),
		Variables: map[string]string{},
		Conversion: config.Conversion{
			Enabled: false,
			Formats: []string{},
		},
	}

	// write autopdf.yaml to current directory
	defaultPath, err := os.Getwd()
	if err != nil {
		return configs.BuildError
	}
	defaultPath = filepath.Join(defaultPath, configs.DefaultConfigName)
	err = os.WriteFile(defaultPath, []byte(config.String()), 0644)
	if err != nil {
		return configs.WriteError
	}
	log.Println("Default config file written to:", defaultPath)
	return nil
}
