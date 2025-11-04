// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package variable_resolver

import (
	"fmt"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
)

// VariableResolverAdapter implements domain.VariableResolver
type VariableResolverAdapter struct {
	config    *config.Config
	logger    *logger.LoggerAdapter
	converter *converter.StructConverter
}

// NewVariableResolverAdapter creates a new variable resolver adapter
func NewVariableResolverAdapter(cfg *config.Config, logger *logger.LoggerAdapter) *VariableResolverAdapter {
	return &VariableResolverAdapter{
		config:    cfg,
		logger:    logger,
		converter: converter.BuildWithDefaults(),
	}
}

// NewVariableResolverAdapterWithConverter creates a new variable resolver adapter with custom converter
func NewVariableResolverAdapterWithConverter(cfg *config.Config, logger *logger.LoggerAdapter, conv *converter.StructConverter) *VariableResolverAdapter {
	return &VariableResolverAdapter{
		config:    cfg,
		logger:    logger,
		converter: conv,
	}
}

// Resolve resolves complex variables to simple key-value pairs
// This is now a thin facade that delegates to TemplateVariables.Flatten()
func (vra *VariableResolverAdapter) Resolve(variables *generation.TemplateVariables) (map[string]string, error) {
	if variables == nil {
		vra.logger.Warn("Received nil TemplateVariables, returning empty map")
		return make(map[string]string), nil
	}

	vra.logger.DebugWithFields("Starting variable resolution",
		"variable_count", variables.Len(),
		"variable_keys", variables.Keys(),
	)

	// Delegate to TemplateVariables.Flatten() which uses the converter internally
	result := variables.Flatten()

	vra.logger.InfoWithFields("Variable resolution complete",
		"input_count", variables.Len(),
		"output_count", len(result),
	)

	return result, nil
}

// Flatten flattens nested variables into dot-notation paths
// This is now a thin facade that delegates to TemplateVariables.Flatten()
func (vra *VariableResolverAdapter) Flatten(variables *generation.TemplateVariables) map[string]string {
	if variables == nil {
		vra.logger.Warn("Received nil TemplateVariables, returning empty map")
		return make(map[string]string)
	}

	vra.logger.DebugWithFields("Flattening variables",
		"variable_count", variables.Len(),
	)

	// Delegate to TemplateVariables.Flatten()
	result := variables.Flatten()

	vra.logger.DebugWithFields("Flattening complete",
		"output_count", len(result),
	)

	return result
}

// Validate validates variables for correctness
// This is now a thin facade that delegates to TemplateVariables.Validate()
func (vra *VariableResolverAdapter) Validate(variables *generation.TemplateVariables) error {
	if variables == nil {
		return domain.VariableResolutionError{
			Code:    domain.ErrCodeVariableInvalid,
			Message: "variables cannot be nil",
			Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh),
		}
	}

	vra.logger.DebugWithFields("Validating variables",
		"variable_count", variables.Len(),
	)

	// Delegate to TemplateVariables.Validate()
	if err := variables.Validate(); err != nil {
		vra.logger.ErrorWithFields("Variable validation failed",
			"error", err,
		)
		// Format the error message properly to avoid literal %s
		errorMessage := fmt.Sprintf(api.ErrVariableValidationFailed, err.Error())
		return domain.VariableResolutionError{
			Code:    domain.ErrCodeVariableInvalid,
			Message: errorMessage,
			Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
				WithError(err),
		}
	}

	vra.logger.InfoWithFields("Variable validation complete",
		"variable_count", variables.Len(),
	)

	return nil
}

// Note: All old helper methods (resolveValue, mapToString, sliceToString, validateValue, etc.)
// have been removed as we now delegate to TemplateVariables which uses the StructConverter internally.
// This makes VariableResolverAdapter a thin facade as per the CLARITY principle.

// ConvertStruct converts a Go struct to TemplateVariables using the struct converter
func (vra *VariableResolverAdapter) ConvertStruct(v interface{}) (*generation.TemplateVariables, error) {
	vra.logger.DebugWithFields("Starting struct conversion",
		"struct_type", fmt.Sprintf("%T", v),
	)

	// Use TemplateVariables.NewTemplateVariablesFromStruct which delegates to converter
	templateVars, err := generation.NewTemplateVariablesFromStruct(v, vra.converter)
	if err != nil {
		vra.logger.ErrorWithFields("Failed to convert struct",
			"struct_type", fmt.Sprintf("%T", v),
			"error", err,
		)
		// Format the error message properly to avoid literal %s
		errorMessage := fmt.Sprintf(api.ErrVariableResolutionFailed, err.Error())
		return nil, domain.VariableResolutionError{
			Code:    domain.ErrCodeVariableInvalid,
			Message: errorMessage,
			Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
				AddContext("struct_type", fmt.Sprintf("%T", v)).
				WithError(err),
		}
	}

	vra.logger.InfoWithFields("Struct conversion complete",
		"struct_type", fmt.Sprintf("%T", v),
		"variable_count", templateVars.Len(),
	)

	return templateVars, nil
}

// ConvertStructToMap converts a Go struct to map[string]interface{} for backward compatibility
// Deprecated: Use ConvertStruct which returns TemplateVariables instead
func (vra *VariableResolverAdapter) ConvertStructToMap(v interface{}) (map[string]interface{}, error) {
	templateVars, err := vra.ConvertStruct(v)
	if err != nil {
		return nil, err
	}
	return templateVars.ToMap(), nil
}

// Note: variablesToMap, variableToInterface, and isSupportedType helper methods
// have been removed as they are no longer needed with the new TemplateVariables approach.
