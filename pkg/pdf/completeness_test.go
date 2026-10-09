// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package pdf_test

import (
	"bytes"
	"errors"
	"math/rand"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/pdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspect(t *testing.T) {
	cases := []struct {
		name    string
		content []byte
		want    pdf.Defect
	}{
		{"empty", nil, pdf.DefectEmpty},
		{"not a pdf", []byte("hello %%EOF"), pdf.DefectMissingHeader},
		{"header only", []byte("%PDF-1.7\n"), pdf.DefectMissingTrailer},
		{"finished", []byte("%PDF-1.7\nbody\n%%EOF\n"), pdf.DefectNone},
		{"trailer just inside the window", append([]byte("%PDF-1.7\n%%EOF"), bytes.Repeat([]byte(" "), pdf.TrailerWindow-5)...), pdf.DefectNone},
		{"trailer outside the window", append([]byte("%PDF-1.7\n%%EOF"), bytes.Repeat([]byte(" "), pdf.TrailerWindow)...), pdf.DefectMissingTrailer},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, pdf.Inspect(c.content))
		})
	}
}

// A finished PDF cut anywhere before its trailer is never finished. This is
// the shape of the production failure: a 1.69 MB prefix of a 3.0 MB letter.
func TestInspectRefusesEveryTruncation(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	for run := 0; run < 200; run++ {
		body := make([]byte, random.Intn(8*pdf.TrailerWindow))
		for i := range body {
			body[i] = byte('a' + random.Intn(26))
		}
		finished := append(append([]byte("%PDF-1.7\n"), body...), []byte("\n%%EOF\n")...)
		require.Equal(t, pdf.DefectNone, pdf.Inspect(finished))

		cut := len("%PDF-") + random.Intn(len(finished)-len("%PDF-")-len("%%EOF\n"))
		assert.NotEqual(t, pdf.DefectNone, pdf.Inspect(finished[:cut]), "cut at %d of %d", cut, len(finished))
	}
}

func TestCheckNamesTheDefect(t *testing.T) {
	err := pdf.Check("/out/letter.pdf", []byte("%PDF-1.7\npartial"))

	var incomplete *pdf.IncompleteError
	require.True(t, errors.As(err, &incomplete))
	assert.Equal(t, pdf.DefectMissingTrailer, incomplete.Defect)
	assert.Equal(t, "/out/letter.pdf", incomplete.Path)
	assert.NoError(t, pdf.Check("/out/letter.pdf", []byte("%PDF-1.7\n%%EOF")))
}
