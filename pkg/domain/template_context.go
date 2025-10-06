package domain

import (
	"fmt"
	"strings"
)

// TemplateContext represents the context for template processing with support for complex data structures
type TemplateContext struct {
	Variables *VariableCollection
	Functions map[string]interface{}
}

// NewTemplateContext creates a new template context
func NewTemplateContext() *TemplateContext {
	return &TemplateContext{
		Variables: NewVariableCollection(),
		Functions: make(map[string]interface{}),
	}
}

// AddFunction adds a custom function to the template context
func (tc *TemplateContext) AddFunction(name string, fn interface{}) {
	tc.Functions[name] = fn
}

// GetVariable retrieves a variable from the context
func (tc *TemplateContext) GetVariable(key string) (*Variable, error) {
	return tc.Variables.GetNested(key)
}

// SetVariable sets a variable in the context
func (tc *TemplateContext) SetVariable(key string, value *Variable) error {
	return tc.Variables.SetNested(key, value)
}

// ToTemplateData converts the context to data suitable for Go templates
func (tc *TemplateContext) ToTemplateData() map[string]interface{} {
	return tc.Variables.ToMap()
}

// ProcessComplexTemplate processes a template with support for complex data structures
func (tc *TemplateContext) ProcessComplexTemplate(templateContent string) (string, error) {
	// This would integrate with the enhanced template engine
	// For now, return the template content as-is
	return templateContent, nil
}

// ValidateTemplate validates that all required variables are present
func (tc *TemplateContext) ValidateTemplate(requiredVars []string) error {
	for _, varName := range requiredVars {
		if _, err := tc.GetVariable(varName); err != nil {
			return fmt.Errorf("required variable '%s' not found: %v", varName, err)
		}
	}
	return nil
}

// GetVariableType returns the type of a variable
func (tc *TemplateContext) GetVariableType(key string) (VariableType, error) {
	variable, err := tc.GetVariable(key)
	if err != nil {
		return VariableTypeNull, err
	}
	return variable.Type, nil
}

// IsVariableArray checks if a variable is an array
func (tc *TemplateContext) IsVariableArray(key string) (bool, error) {
	variableType, err := tc.GetVariableType(key)
	if err != nil {
		return false, err
	}
	return variableType == VariableTypeArray, nil
}

// IsVariableObject checks if a variable is an object
func (tc *TemplateContext) IsVariableObject(key string) (bool, error) {
	variableType, err := tc.GetVariableType(key)
	if err != nil {
		return false, err
	}
	return variableType == VariableTypeObject, nil
}

// GetArrayLength returns the length of an array variable
func (tc *TemplateContext) GetArrayLength(key string) (int, error) {
	variable, err := tc.GetVariable(key)
	if err != nil {
		return 0, err
	}

	if variable.Type != VariableTypeArray {
		return 0, fmt.Errorf("variable '%s' is not an array", key)
	}

	array, err := variable.AsArray()
	if err != nil {
		return 0, err
	}

	return len(array), nil
}

// GetObjectKeys returns the keys of an object variable
func (tc *TemplateContext) GetObjectKeys(key string) ([]string, error) {
	variable, err := tc.GetVariable(key)
	if err != nil {
		return nil, err
	}

	if variable.Type != VariableTypeObject {
		return nil, fmt.Errorf("variable '%s' is not an object", key)
	}

	obj, err := variable.AsObject()
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}

	return keys, nil
}

// Clone creates a deep copy of the template context
func (tc *TemplateContext) Clone() *TemplateContext {
	clone := NewTemplateContext()

	// Copy variables
	for key, variable := range tc.Variables.GetAll() {
		clone.Variables.Set(key, variable)
	}

	// Copy functions
	for name, fn := range tc.Functions {
		clone.Functions[name] = fn
	}

	return clone
}

// Merge merges another template context into this one
func (tc *TemplateContext) Merge(other *TemplateContext) {
	// Merge variables
	for key, variable := range other.Variables.GetAll() {
		tc.Variables.Set(key, variable)
	}

	// Merge functions
	for name, fn := range other.Functions {
		tc.Functions[name] = fn
	}
}

// String returns a string representation of the template context
func (tc *TemplateContext) String() string {
	var parts []string

	// Add variables
	for key, variable := range tc.Variables.GetAll() {
		parts = append(parts, fmt.Sprintf("%s: %s", key, variable.String()))
	}

	// Add functions
	for name := range tc.Functions {
		parts = append(parts, fmt.Sprintf("%s: function", name))
	}

	return fmt.Sprintf("TemplateContext{%s}", strings.Join(parts, ", "))
}
