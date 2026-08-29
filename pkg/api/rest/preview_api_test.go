package rest

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	rootapi "github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
	"github.com/go-chi/chi/v5"
)

type previewCompilerFunc func(context.Context, preview.CompileRequest) (preview.CompileOutput, error)

func (function previewCompilerFunc) Compile(ctx context.Context, request preview.CompileRequest) (preview.CompileOutput, error) {
	return function(ctx, request)
}

func previewDocument(content string) document.DocumentSpec {
	return document.DocumentSpec{
		SchemaVersion: document.CurrentSchemaVersion, ID: "preview",
		Blocks: []document.Component{{ID: "body", Kind: "text", Variant: "default", Props: document.Props{"content": content}}},
	}
}

func TestPreviewAPIMonotonicRevisionsPartialUpdatesAndReconnect(t *testing.T) {
	engine, err := rootapi.NewDefaultDocumentEngine("latex", rootapi.DocumentEngineConfig{})
	if err != nil {
		t.Fatal(err)
	}
	transport, err := NewPreviewAPI(PreviewAPIOptions{
		Engine: engine,
		CompilerFactory: PreviewCompilerFactoryFunc(func() (preview.Compiler, error) {
			return previewCompilerFunc(func(_ context.Context, request preview.CompileRequest) (preview.CompileOutput, error) {
				data, readErr := os.ReadFile(filepath.Join(request.Workspace, request.Main))
				if readErr != nil {
					return preview.CompileOutput{}, readErr
				}
				return preview.CompileOutput{Pages: []preview.Page{{Number: 1, MediaType: "image/png", Data: data}}}, nil
			}), nil
		}),
		SessionOptions: preview.Options{WorkspaceRoot: t.TempDir()},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer transport.Close()
	router := chi.NewRouter()
	router.Mount("/preview", transport.Routes())
	server := httptest.NewServer(router)
	defer server.Close()

	sessionID := createPreviewSession(t, server.URL)
	if status := submitPreviewRevision(t, server.URL, sessionID, 2, previewDocument("first")); status != http.StatusAccepted {
		t.Fatalf("revision 2 status %d", status)
	}
	if status := submitPreviewRevision(t, server.URL, sessionID, 1, previewDocument("stale")); status != http.StatusConflict {
		t.Fatalf("stale revision status %d", status)
	}
	first := readPreviewEvent(t, server.URL, sessionID, 0)
	if first.Type != "result" || first.Revision != 2 || first.Result == nil || len(first.Result.ChangedPages) != 1 {
		t.Fatalf("first event wrong: %+v", first)
	}
	if status := submitPreviewRevision(t, server.URL, sessionID, 3, previewDocument("first")); status != http.StatusAccepted {
		t.Fatalf("revision 3 status %d", status)
	}
	second := readPreviewEvent(t, server.URL, sessionID, first.ID)
	if second.Revision != 3 || second.Result == nil || len(second.Result.ChangedPages) != 0 || len(second.Result.RemovedPages) != 0 {
		t.Fatalf("reconnect/partial event wrong: %+v", second)
	}
}

func TestPreviewAPITTLPublishesClosureAndRemovesSession(t *testing.T) {
	engine, err := rootapi.NewDefaultDocumentEngine("latex", rootapi.DocumentEngineConfig{})
	if err != nil {
		t.Fatal(err)
	}
	transport, err := NewPreviewAPI(PreviewAPIOptions{
		Engine: engine,
		CompilerFactory: PreviewCompilerFactoryFunc(func() (preview.Compiler, error) {
			return previewCompilerFunc(func(context.Context, preview.CompileRequest) (preview.CompileOutput, error) {
				return preview.CompileOutput{}, nil
			}), nil
		}),
		SessionOptions: preview.Options{WorkspaceRoot: t.TempDir(), IdleTTL: 20 * time.Millisecond},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer transport.Close()
	router := chi.NewRouter()
	router.Mount("/preview", transport.Routes())
	server := httptest.NewServer(router)
	defer server.Close()
	sessionID := createPreviewSession(t, server.URL)
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		transport.mu.RLock()
		_, exists := transport.sessions[sessionID]
		transport.mu.RUnlock()
		if !exists {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("expired session remained addressable")
}

func createPreviewSession(t *testing.T, baseURL string) string {
	t.Helper()
	response, err := http.Post(baseURL+"/preview/sessions", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create status %d", response.StatusCode)
	}
	var body previewSessionResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body.SessionID
}

func submitPreviewRevision(t *testing.T, baseURL, sessionID string, revision uint64, spec document.DocumentSpec) int {
	t.Helper()
	body, err := json.Marshal(previewRevisionRequest{Document: spec})
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/preview/sessions/%s/revisions/%d", baseURL, sessionID, revision), bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	return response.StatusCode
}

func readPreviewEvent(t *testing.T, baseURL, sessionID string, after uint64) PreviewEvent {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/preview/sessions/%s/events?after=%d", baseURL, sessionID, after), nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var event PreviewEvent
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
			t.Fatal(err)
		}
		return event
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	t.Fatal("event stream ended without event")
	return PreviewEvent{}
}
