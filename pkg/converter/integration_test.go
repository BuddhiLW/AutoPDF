// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter_test

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test structs
type IntegrationUser struct {
	ID        int       `autopdf:"id"`
	Name      string    `autopdf:"name"`
	Email     string    `autopdf:"email"`
	CreatedAt time.Time `autopdf:"created_at"`
	Profile   struct {
		Bio     string `autopdf:"bio"`
		Website string `autopdf:"website"`
		Avatar  string `autopdf:"avatar"`
	} `autopdf:"profile"`
	Settings struct {
		Theme    string `autopdf:"theme"`
		Language string `autopdf:"language"`
		Privacy  string `autopdf:"privacy"`
	} `autopdf:"settings"`
	Tags []string `autopdf:"tags,flatten"`
}

type IntegrationDocument struct {
	Title       string                 `autopdf:"title"`
	Author      IntegrationUser        `autopdf:"author"`
	Content     string                 `autopdf:"content"`
	CreatedAt   time.Time              `autopdf:"created_at"`
	UpdatedAt   *time.Time             `autopdf:"updated_at"`
	URL         url.URL                `autopdf:"url"`
	Tags        []string               `autopdf:"tags,flatten"`
	Metadata    map[string]interface{} `autopdf:"metadata"`
	IsPublished bool                   `autopdf:"is_published"`
	Version     int                    `autopdf:"version"`
}

type IntegrationReport struct {
	ReportID    string                `autopdf:"report_id"`
	Title       string                `autopdf:"title"`
	GeneratedAt time.Time             `autopdf:"generated_at"`
	Author      IntegrationUser       `autopdf:"author"`
	Data        []IntegrationDocument `autopdf:"data"`
	Summary     struct {
		TotalDocuments int     `autopdf:"total_documents"`
		TotalUsers     int     `autopdf:"total_users"`
		AverageScore   float64 `autopdf:"average_score"`
	} `autopdf:"summary"`
	Config struct {
		Format   string `autopdf:"format"`
		Language string `autopdf:"language"`
		Timezone string `autopdf:"timezone"`
	} `autopdf:"config"`
}

// TestIntegrationWithComplexDocument tests integration with complex document structures
func TestIntegrationWithComplexDocument(t *testing.T) {
	converter := converter.BuildWithDefaults()

	now := time.Now()
	docURL, _ := url.Parse("https://example.com/document/123")

	document := IntegrationDocument{
		Title: "AutoPDF Integration Test",
		Author: IntegrationUser{
			ID:        456,
			Name:      "Jane Smith",
			Email:     "jane.smith@example.com",
			CreatedAt: time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC),
			Profile: struct {
				Bio     string `autopdf:"bio"`
				Website string `autopdf:"website"`
				Avatar  string `autopdf:"avatar"`
			}{
				Bio:     "Technical writer",
				Website: "https://janesmith.dev",
				Avatar:  "https://example.com/jane-avatar.jpg",
			},
			Settings: struct {
				Theme    string `autopdf:"theme"`
				Language string `autopdf:"language"`
				Privacy  string `autopdf:"privacy"`
			}{
				Theme:    "light",
				Language: "en",
				Privacy:  "private",
			},
			Tags: []string{"writer", "technical"},
		},
		Content:   "This is a test document for AutoPDF integration.",
		CreatedAt: now,
		UpdatedAt: &now,
		URL:       *docURL,
		Tags:      []string{"integration", "test", "autopdf"},
		Metadata: map[string]interface{}{
			"source":   "api",
			"version":  "1.0",
			"category": "technical",
		},
		IsPublished: true,
		Version:     1,
	}

	variables, err := converter.ConvertStruct(document)
	require.NoError(t, err)
	require.NotNil(t, variables)

	// Verify document fields
	flattened := variables.Flatten()
	assert.Equal(t, "AutoPDF Integration Test", flattened["title"])
	assert.Equal(t, "This is a test document for AutoPDF integration.", flattened["content"])
	assert.Equal(t, "true", flattened["is_published"])
	assert.Equal(t, "1", flattened["version"])
	assert.Equal(t, "https://example.com/document/123", flattened["url"])
	assert.Equal(t, "integration, test, autopdf", flattened["tags"])

	// Verify author fields
	assert.Equal(t, "456", flattened["author.id"])
	assert.Equal(t, "Jane Smith", flattened["author.name"])
	assert.Equal(t, "jane.smith@example.com", flattened["author.email"])
	assert.Equal(t, "Technical writer", flattened["author.profile.bio"])
	assert.Equal(t, "https://janesmith.dev", flattened["author.profile.website"])
	assert.Equal(t, "light", flattened["author.settings.theme"])
	assert.Equal(t, "writer, technical", flattened["author.tags"])

	// Verify metadata
	assert.Equal(t, "api", flattened["metadata.source"])
	assert.Equal(t, "1.0", flattened["metadata.version"])
	assert.Equal(t, "technical", flattened["metadata.category"])
}

// TestIntegrationWithReport tests integration with complex report structures
func TestIntegrationWithReport(t *testing.T) {
	converter := converter.BuildWithDefaults()

	// Create test documents
	doc1 := IntegrationDocument{
		Title:       "Document 1",
		Content:     "Content 1",
		CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Tags:        []string{"doc1", "test"},
		IsPublished: true,
		Version:     1,
	}

	doc2 := IntegrationDocument{
		Title:       "Document 2",
		Content:     "Content 2",
		CreatedAt:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		Tags:        []string{"doc2", "test"},
		IsPublished: false,
		Version:     2,
	}

	report := IntegrationReport{
		ReportID:    "RPT-2025-001",
		Title:       "Monthly Report",
		GeneratedAt: time.Date(2025, 1, 7, 15, 30, 0, 0, time.UTC),
		Author: IntegrationUser{
			ID:        789,
			Name:      "Report Generator",
			Email:     "reports@example.com",
			CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Tags:      []string{"system", "reports"},
		},
		Data: []IntegrationDocument{doc1, doc2},
		Summary: struct {
			TotalDocuments int     `autopdf:"total_documents"`
			TotalUsers     int     `autopdf:"total_users"`
			AverageScore   float64 `autopdf:"average_score"`
		}{
			TotalDocuments: 2,
			TotalUsers:     1,
			AverageScore:   95.5,
		},
		Config: struct {
			Format   string `autopdf:"format"`
			Language string `autopdf:"language"`
			Timezone string `autopdf:"timezone"`
		}{
			Format:   "PDF",
			Language: "en",
			Timezone: "UTC",
		},
	}

	variables, err := converter.ConvertStruct(report)
	require.NoError(t, err)
	require.NotNil(t, variables)

	flattened := variables.Flatten()

	// Verify report fields
	assert.Equal(t, "RPT-2025-001", flattened["report_id"])
	assert.Equal(t, "Monthly Report", flattened["title"])
	assert.Equal(t, "2025-01-07 15:30:00", flattened["generated_at"])

	// Verify author
	assert.Equal(t, "789", flattened["author.id"])
	assert.Equal(t, "Report Generator", flattened["author.name"])
	assert.Equal(t, "reports@example.com", flattened["author.email"])
	assert.Equal(t, "system, reports", flattened["author.tags"])

	// Verify summary
	assert.Equal(t, "2", flattened["summary.total_documents"])
	assert.Equal(t, "1", flattened["summary.total_users"])
	assert.Equal(t, "95.5", flattened["summary.average_score"])

	// Verify config
	assert.Equal(t, "PDF", flattened["config.format"])
	assert.Equal(t, "en", flattened["config.language"])
	assert.Equal(t, "UTC", flattened["config.timezone"])

	// Verify data array (should be converted to indexed fields)
	assert.Equal(t, "Document 1", flattened["data[0].title"])
	assert.Equal(t, "Content 1", flattened["data[0].content"])
	assert.Equal(t, "true", flattened["data[0].is_published"])
	assert.Equal(t, "1", flattened["data[0].version"])
	assert.Equal(t, "doc1, test", flattened["data[0].tags"])

	assert.Equal(t, "Document 2", flattened["data[1].title"])
	assert.Equal(t, "Content 2", flattened["data[1].content"])
	assert.Equal(t, "false", flattened["data[1].is_published"])
	assert.Equal(t, "2", flattened["data[1].version"])
	assert.Equal(t, "doc2, test", flattened["data[1].tags"])
}

// CustomID is a custom type for testing
type CustomID struct {
	Prefix string
	Number int
}

// ToAutoPDFVariable implements the custom conversion interface
func (cid CustomID) ToAutoPDFVariable() (config.Variable, error) {
	return &config.StringVariable{Value: fmt.Sprintf("%s-%d", cid.Prefix, cid.Number)}, nil
}

// TestIntegrationWithCustomConverters tests integration with custom converters
func TestIntegrationWithCustomConverters(t *testing.T) {

	type CustomDocument struct {
		ID      CustomID `autopdf:"id"`
		Title   string   `autopdf:"title"`
		Content string   `autopdf:"content"`
	}

	converter := converter.BuildWithDefaults()

	document := CustomDocument{
		ID:      CustomID{Prefix: "DOC", Number: 12345},
		Title:   "Custom Document",
		Content: "This document uses custom ID conversion.",
	}

	variables, err := converter.ConvertStruct(document)
	require.NoError(t, err)
	require.NotNil(t, variables)

	// Verify custom conversion
	flattened := variables.Flatten()
	assert.Equal(t, "DOC-12345", flattened["id"])
	assert.Equal(t, "Custom Document", flattened["title"])
	assert.Equal(t, "This document uses custom ID conversion.", flattened["content"])
}

// TestIntegrationWithNilValues tests integration with nil values and edge cases
func TestIntegrationWithNilValues(t *testing.T) {
	type NilTestStruct struct {
		StringPtr   *string                 `autopdf:"string_ptr"`
		IntPtr      *int                    `autopdf:"int_ptr"`
		TimePtr     *time.Time              `autopdf:"time_ptr"`
		URLPtr      *url.URL                `autopdf:"url_ptr"`
		StructPtr   *IntegrationUser        `autopdf:"struct_ptr"`
		SlicePtr    *[]string               `autopdf:"slice_ptr"`
		MapPtr      *map[string]interface{} `autopdf:"map_ptr"`
		NormalField string                  `autopdf:"normal_field"`
	}

	converter := converter.BuildWithDefaults()

	testStruct := NilTestStruct{
		StringPtr:   nil,
		IntPtr:      nil,
		TimePtr:     nil,
		URLPtr:      nil,
		StructPtr:   nil,
		SlicePtr:    nil,
		MapPtr:      nil,
		NormalField: "normal value",
	}

	variables, err := converter.ConvertStruct(testStruct)
	require.NoError(t, err)
	require.NotNil(t, variables)

	flattened := variables.Flatten()

	// Verify nil values are handled gracefully
	assert.Equal(t, "", flattened["string_ptr"])
	assert.Equal(t, "", flattened["int_ptr"])
	assert.Equal(t, "", flattened["time_ptr"])
	assert.Equal(t, "", flattened["url_ptr"])
	assert.Equal(t, "", flattened["struct_ptr"])
	assert.Equal(t, "", flattened["slice_ptr"])
	assert.Equal(t, "", flattened["map_ptr"])
	assert.Equal(t, "normal value", flattened["normal_field"])
}

// TestIntegrationWithOmitEmpty tests integration with omitempty functionality
func TestIntegrationWithOmitEmpty(t *testing.T) {
	type OmitEmptyTestStruct struct {
		Name        string                 `autopdf:"name"`
		Description string                 `autopdf:"description,omitempty"`
		Value       int                    `autopdf:"value,omitempty"`
		Active      bool                   `autopdf:"active,omitempty"`
		Tags        []string               `autopdf:"tags,omitempty,flatten"`
		Metadata    map[string]interface{} `autopdf:"metadata,omitempty"`
	}

	converter := converter.BuildForTemplates() // This has omitempty enabled

	// Test with empty values
	testStruct := OmitEmptyTestStruct{
		Name: "Test Item",
		// All other fields are empty/zero values
	}

	variables, err := converter.ConvertStruct(testStruct)
	require.NoError(t, err)
	require.NotNil(t, variables)

	flattened := variables.Flatten()

	// Verify that only non-empty fields are present
	assert.Equal(t, "Test Item", flattened["name"])
	_, exists := flattened["description"]
	assert.False(t, exists, "description should be omitted (empty)")
	_, exists = flattened["value"]
	assert.False(t, exists, "value should be omitted (zero)")
	_, exists = flattened["active"]
	assert.False(t, exists, "active should be omitted (false)")
	_, exists = flattened["tags"]
	assert.False(t, exists, "tags should be omitted (empty slice)")
	_, exists = flattened["metadata"]
	assert.False(t, exists, "metadata should be omitted (empty map)")

	// Test with non-empty values
	testStruct2 := OmitEmptyTestStruct{
		Name:        "Test Item 2",
		Description: "This is a description",
		Value:       42,
		Active:      true,
		Tags:        []string{"tag1", "tag2"},
		Metadata:    map[string]interface{}{"key": "value"},
	}

	variables2, err := converter.ConvertStruct(testStruct2)
	require.NoError(t, err)
	require.NotNil(t, variables2)

	flattened2 := variables2.Flatten()

	// Verify that all fields are present
	assert.Equal(t, "Test Item 2", flattened2["name"])
	assert.Equal(t, "This is a description", flattened2["description"])
	assert.Equal(t, "42", flattened2["value"])
	assert.Equal(t, "true", flattened2["active"])
	assert.Equal(t, "tag1, tag2", flattened2["tags"])
	assert.Equal(t, "value", flattened2["metadata.key"])
}

// TestIntegrationWithFlattening tests integration with flattening functionality
func TestIntegrationWithFlattening(t *testing.T) {
	type FlattenTestStruct struct {
		User struct {
			Profile struct {
				Bio     string `autopdf:"bio"`
				Website string `autopdf:"website"`
			} `autopdf:"profile"`
		} `autopdf:"user,flatten"`
		Tags []string `autopdf:"tags,flatten"`
	}

	converter := converter.BuildWithDefaults()

	testStruct := FlattenTestStruct{
		User: struct {
			Profile struct {
				Bio     string `autopdf:"bio"`
				Website string `autopdf:"website"`
			} `autopdf:"profile"`
		}{
			Profile: struct {
				Bio     string `autopdf:"bio"`
				Website string `autopdf:"website"`
			}{
				Bio:     "Software engineer",
				Website: "https://example.com",
			},
		},
		Tags: []string{"developer", "golang", "autopdf"},
	}

	variables, err := converter.ConvertStruct(testStruct)
	require.NoError(t, err)
	require.NotNil(t, variables)

	flattened := variables.Flatten()

	// Verify flattening works correctly
	assert.Equal(t, "Software engineer", flattened["user.profile.bio"])
	assert.Equal(t, "https://example.com", flattened["user.profile.website"])
	assert.Equal(t, "developer, golang, autopdf", flattened["tags"])
}
