package document

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Problem is a field-addressable validation failure suitable for API and web
// editor responses.
type Problem struct {
	ComponentID string `json:"componentId,omitempty"`
	Path        string `json:"path"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}

func (problem Problem) Error() string {
	if problem.ComponentID == "" {
		return fmt.Sprintf("%s: %s", problem.Path, problem.Message)
	}
	return fmt.Sprintf("component %s %s: %s", problem.ComponentID, problem.Path, problem.Message)
}

// Problems aggregates validation failures without losing their field paths.
type Problems []Problem

func (problems Problems) Error() string {
	messages := make([]string, len(problems))
	for index, problem := range problems {
		messages[index] = problem.Error()
	}
	return strings.Join(messages, "; ")
}

// Problems returns every structural validation failure in stable tree order.
func (spec DocumentSpec) Problems() Problems {
	var problems Problems
	if spec.SchemaVersion != CurrentSchemaVersion {
		problems = append(problems, Problem{
			Path:    "schemaVersion",
			Code:    "document.schema_version.unsupported",
			Message: fmt.Sprintf("expected %d, got %d", CurrentSchemaVersion, spec.SchemaVersion),
		})
	}

	seenIDs := make(map[string]string)
	for index := range spec.Blocks {
		problems = append(problems, validateComponent(spec.Blocks[index], fmt.Sprintf("blocks[%d]", index), seenIDs)...)
	}
	return problems
}

func validateComponent(component Component, path string, seenIDs map[string]string) Problems {
	var problems Problems
	id := strings.TrimSpace(component.ID)
	if id == "" {
		problems = append(problems, Problem{Path: path + ".id", Code: "component.id.required", Message: "component ID is required"})
	} else if firstPath, exists := seenIDs[id]; exists {
		problems = append(problems, Problem{
			ComponentID: id,
			Path:        path + ".id",
			Code:        "component.id.duplicate",
			Message:     "component ID already used at " + firstPath,
		})
	} else {
		seenIDs[id] = path
	}
	if strings.TrimSpace(component.Kind) == "" {
		problems = append(problems, Problem{ComponentID: id, Path: path + ".kind", Code: "component.kind.required", Message: "component kind is required"})
	}
	if strings.TrimSpace(component.Variant) == "" {
		problems = append(problems, Problem{ComponentID: id, Path: path + ".variant", Code: "component.variant.required", Message: "component variant is required"})
	}
	if _, err := json.Marshal(component.Props); err != nil {
		problems = append(problems, Problem{ComponentID: id, Path: path + ".props", Code: "component.props.invalid", Message: err.Error()})
	}
	if _, err := json.Marshal(component.Style); err != nil {
		problems = append(problems, Problem{ComponentID: id, Path: path + ".style", Code: "component.style.invalid", Message: err.Error()})
	}

	assetIDs := make(map[string]struct{}, len(component.Assets))
	for index, asset := range component.Assets {
		assetPath := fmt.Sprintf("%s.assets[%d].id", path, index)
		if strings.TrimSpace(asset.ID) == "" {
			problems = append(problems, Problem{ComponentID: id, Path: assetPath, Code: "asset.id.required", Message: "asset ID is required"})
			continue
		}
		if _, exists := assetIDs[asset.ID]; exists {
			problems = append(problems, Problem{ComponentID: id, Path: assetPath, Code: "asset.id.duplicate", Message: "asset ID is duplicated within component"})
		}
		assetIDs[asset.ID] = struct{}{}
	}

	for index := range component.Children {
		childPath := fmt.Sprintf("%s.children[%d]", path, index)
		problems = append(problems, validateComponent(component.Children[index], childPath, seenIDs)...)
	}
	return problems
}
