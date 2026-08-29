package api

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

// GeneratePDF generates a PDF using the internal application layer
// This function maintains the same signature while using the adapter pattern
func GeneratePDF(cfg *config.Config, template config.Template) ([]byte, map[string]string, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("api: config must not be nil")
	}
	templatePath := cfg.Template.String()
	if templatePath == "" {
		templatePath = template.String()
	}
	engine, err := NewEngine()
	if err != nil {
		return nil, nil, err
	}
	result, err := engine.Generate(context.Background(), Request{
		TemplatePath: templatePath,
		OutputPath:   cfg.Output.String(),
		Variables:    cfg.Variables.Flatten(),
		LaTeXEngine:  cfg.Engine.String(),
		Passes:       cfg.Passes,
		UseLatexmk:   cfg.UseLatexmk,
		Conversion: ConversionOptions{
			Enabled: cfg.Conversion.Enabled,
			Formats: append([]string(nil), cfg.Conversion.Formats...),
		},
	})
	if err != nil {
		return nil, nil, err
	}
	paths := make(map[string]string, len(result.ImagePaths))
	for index, imagePath := range result.ImagePaths {
		paths[fmt.Sprintf("image_%d", index)] = imagePath
	}
	return result.PDF, paths, nil
}

// func convertToFormat(file string, format string) string {
// 	dir := filepath.Dir(file)
// 	filename := filepath.Base(file)
// 	ext := filepath.Ext(filename)
// 	newFilename := strings.TrimSuffix(filename, ext) + "." + format
// 	return filepath.Join(dir, newFilename)
