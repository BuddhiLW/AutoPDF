// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter_test

import (
	"fmt"
	"sort"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/converter"
)

type Address struct {
	Street string `autopdf:"street"`
	City   string `autopdf:"city"`
}

type User struct {
	Name      string    `autopdf:"name"`
	CreatedAt time.Time `autopdf:"created_at"`
	Address   Address   `autopdf:"address"`
	Tags      []string  `autopdf:"tags"`
	IsActive  bool      `autopdf:"is_active"`
}

func printFlat(vars *config.Variables) {
	flat := vars.Flatten()
	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s = %s\n", k, flat[k])
	}
}

func ExampleStructConverter_ConvertStruct() {
	user := User{
		Name:      "John Doe",
		CreatedAt: time.Date(2025, 1, 7, 12, 0, 0, 0, time.UTC),
		Address:   Address{Street: "123 Main St", City: "Anytown"},
		Tags:      []string{"developer", "golang"},
		IsActive:  true,
	}

	vars, err := converter.BuildWithDefaults().ConvertStruct(user)
	if err != nil {
		fmt.Println(err)
		return
	}
	printFlat(vars)
	// Output:
	// address.city = Anytown
	// address.street = 123 Main St
	// created_at = 2025-01-07 12:00:00
	// is_active = true
	// name = John Doe
	// tags[0] = developer
	// tags[1] = golang
}

func ExampleFieldTag() {
	type Flattened struct {
		Name    string   `autopdf:"name"`
		Address Address  `autopdf:"address,flatten"`
		Tags    []string `autopdf:"tags,flatten"`
	}
	type Inlined struct {
		Name    string  `autopdf:"name"`
		Address Address `autopdf:"address,inline"`
	}

	conv := converter.BuildWithDefaults()
	for _, v := range []any{
		Flattened{Name: "Bob", Address: Address{Street: "321 Elm St", City: "Test City"}, Tags: []string{"user", "tester"}},
		Inlined{Name: "Charlie", Address: Address{Street: "654 Maple St", City: "View City"}},
	} {
		vars, err := conv.ConvertStruct(v)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%T:\n", v)
		printFlat(vars)
	}
	// Output:
	// converter_test.Flattened:
	// address.city = Test City
	// address.street = 321 Elm St
	// name = Bob
	// tags = user, tester
	// converter_test.Inlined:
	// city = View City
	// name = Charlie
	// street = 654 Maple St
}

type Secret struct {
	Kind  string
	Value string
}

func (s Secret) ToAutoPDFVariable() (config.Variable, error) {
	return &config.StringVariable{Value: s.Kind + ":" + s.Value}, nil
}

func ExampleAutoPDFFormattable() {
	type Document struct {
		Title  string `autopdf:"title"`
		Secret Secret `autopdf:"secret"`
	}

	vars, err := converter.BuildWithDefaults().ConvertStruct(Document{
		Title:  "Custom Type Example",
		Secret: Secret{Kind: "token", Value: "important-data"},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	printFlat(vars)
	// Output:
	// secret = token:important-data
	// title = Custom Type Example
}

func ExampleNewConverterBuilder() {
	type Report struct {
		Title       string    `autopdf:"title"`
		GeneratedAt time.Time `autopdf:"generated_at"`
		Notes       string    `autopdf:"notes"`
		Tags        []string  `autopdf:"tags,flatten"`
	}

	conv := converter.NewConverterBuilder().
		WithBuiltinConverters().
		WithTimeFormat("02/01/2006").
		WithOmitEmpty(true).
		Build()

	vars, err := conv.ConvertStruct(Report{
		Title:       "Monthly Report",
		GeneratedAt: time.Date(2025, 10, 16, 0, 0, 0, 0, time.UTC),
		Tags:        []string{"monthly", "performance"},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	printFlat(vars)
	// Output:
	// generated_at = 16/10/2025
	// tags = monthly, performance
	// title = Monthly Report
}
