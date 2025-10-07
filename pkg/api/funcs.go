package api

import (
	"log"

	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// GeneratePDF generates a PDF using the internal application layer
// This function maintains the same signature while using the adapter pattern
func GeneratePDF(cfg *config.Config, template config.Template) ([]byte, map[string]string, error) {
	log.Println("Output file:", cfg.Output)
	log.Println("Final merged config:", cfg)

	// Create the internal application adapter
	adapter := adapters.NewInternalApplicationAdapter(cfg)

	// Use the adapter to generate the PDF
	return adapter.GeneratePDF(cfg, template)
}

// func convertToFormat(file string, format string) string {
// 	dir := filepath.Dir(file)
// 	filename := filepath.Base(file)
// 	ext := filepath.Ext(filename)
// 	newFilename := strings.TrimSuffix(filename, ext) + "." + format
// 	return filepath.Join(dir, newFilename)
