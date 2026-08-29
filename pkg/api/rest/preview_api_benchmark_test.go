package rest

import (
	"encoding/json"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/preview"
)

func BenchmarkPreviewEventSerialization(b *testing.B) {
	page := make([]byte, 128*1024)
	event := PreviewEvent{
		ID: 42, Type: "result", Revision: 42,
		Result: &preview.Result{
			Revision: 42,
			ChangedPages: []preview.Page{
				{Number: 2, MediaType: "image/png", Hash: "two", Data: page},
				{Number: 5, MediaType: "image/png", Hash: "five", Data: page},
			},
			RemovedPages: []int{7},
		},
	}
	b.SetBytes(int64(len(page) * 2))
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		if _, err := json.Marshal(event); err != nil {
			b.Fatal(err)
		}
	}
}
