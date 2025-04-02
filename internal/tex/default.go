package tex

import (
	"log"
	"os"

	"github.com/BuddhiLW/AutoPDF/internal/config"
)

func Default(texFile string) {
	// create default config.yaml for autopdf
	config := config.Config{
		Template: config.Template(texFile),
		Output:   config.Output(texFile),
		Engine:   config.Engine("pdflatex"),
		Conversion: config.Conversion{
			Enabled: false,
			Formats: []string{},
		},
	}

	// write autopdf.yaml to current directory
	if err := os.WriteFile("autopdf.yaml", []byte(config.String()), 0644); err != nil {
		log.Fatalf("Failed to write autopdf.yaml: %v", err)
	}
}
