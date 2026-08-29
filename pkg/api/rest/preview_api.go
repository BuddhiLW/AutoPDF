package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	rootapi "github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/document"
	"github.com/BuddhiLW/AutoPDF/pkg/preview"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var (
	ErrPreviewDocumentEngineRequired = errors.New("rest: preview document engine is required")
	ErrPreviewCompilerFactoryNeeded  = errors.New("rest: preview compiler factory is required")
)

// PreviewCompilerFactory gives every browser session isolated compiler state.
type PreviewCompilerFactory interface {
	NewCompiler() (preview.Compiler, error)
}

// PreviewCompilerFactoryFunc adapts a constructor function.
type PreviewCompilerFactoryFunc func() (preview.Compiler, error)

func (function PreviewCompilerFactoryFunc) NewCompiler() (preview.Compiler, error) { return function() }

// PreviewAPIOptions configures the browser-facing SSE transport.
type PreviewAPIOptions struct {
	Engine          *rootapi.DocumentEngine
	CompilerFactory PreviewCompilerFactory
	SessionOptions  preview.Options
	HistoryLimit    int
}

// PreviewAPI owns revisioned preview sessions. Mount Routes at a stable prefix,
// for example /api/v1/preview.
type PreviewAPI struct {
	engine          *rootapi.DocumentEngine
	compilerFactory PreviewCompilerFactory
	sessionOptions  preview.Options
	historyLimit    int

	mu       sync.RWMutex
	sessions map[string]*previewTransportSession
}

// NewPreviewAPI builds the HTTP/SSE adapter without changing legacy PDF routes.
func NewPreviewAPI(options PreviewAPIOptions) (*PreviewAPI, error) {
	if options.Engine == nil {
		return nil, ErrPreviewDocumentEngineRequired
	}
	if options.CompilerFactory == nil {
		return nil, ErrPreviewCompilerFactoryNeeded
	}
	if options.HistoryLimit <= 0 {
		options.HistoryLimit = 32
	}
	return &PreviewAPI{
		engine: options.Engine, compilerFactory: options.CompilerFactory,
		sessionOptions: options.SessionOptions, historyLimit: options.HistoryLimit,
		sessions: make(map[string]*previewTransportSession),
	}, nil
}

// Routes exposes creation, monotonic revision submission, replayable SSE
// events, cancellation, and closure.
func (api *PreviewAPI) Routes() chi.Router {
	router := chi.NewRouter()
	router.Post("/sessions", api.createSession)
	router.Put("/sessions/{sessionID}/revisions/{revision}", api.submitRevision)
	router.Get("/sessions/{sessionID}/events", api.streamEvents)
	router.Delete("/sessions/{sessionID}", api.closeSession)
	return router
}

// RegisterPreviewRoutes mounts the browser transport at /api/v1/preview.
func RegisterPreviewRoutes(router chi.Router, api *PreviewAPI) {
	router.Mount("/api/v1/preview", api.Routes())
}

// Close terminates all sessions owned by this transport.
func (api *PreviewAPI) Close() error {
	api.mu.Lock()
	sessions := make([]*previewTransportSession, 0, len(api.sessions))
	for _, session := range api.sessions {
		sessions = append(sessions, session)
	}
	api.sessions = make(map[string]*previewTransportSession)
	api.mu.Unlock()
	var closeErr error
	for _, session := range sessions {
		if err := session.session.Close(); err != nil {
			closeErr = errors.Join(closeErr, err)
		}
	}
	return closeErr
}

type previewSessionResponse struct {
	SessionID string `json:"sessionId"`
	EventsURL string `json:"eventsUrl"`
}

type previewRevisionRequest struct {
	Document      document.DocumentSpec `json:"document"`
	FocusSections []string              `json:"focusSections,omitempty"`
}

type previewRevisionResponse struct {
	SessionID string `json:"sessionId"`
	Revision  uint64 `json:"revision"`
	Accepted  bool   `json:"accepted"`
}

// PreviewEvent is directly consumable by a reducer: changed pages replace by
// number, removed pages delete by number, and revisions never decrease.
type PreviewEvent struct {
	ID       uint64          `json:"id"`
	Type     string          `json:"type"`
	Revision uint64          `json:"revision"`
	Result   *preview.Result `json:"result,omitempty"`
	Error    string          `json:"error,omitempty"`
}

type previewTransportSession struct {
	session      *preview.Session
	historyLimit int

	mu            sync.Mutex
	lastSubmitted uint64
	nextEventID   uint64
	history       []PreviewEvent
	closed        bool
	notify        chan struct{}
}

func (api *PreviewAPI) createSession(writer http.ResponseWriter, request *http.Request) {
	compiler, err := api.compilerFactory.NewCompiler()
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	session, err := api.engine.NewPreviewSession(compiler, api.sessionOptions)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	id := uuid.NewString()
	transport := &previewTransportSession{session: session, historyLimit: api.historyLimit, notify: make(chan struct{})}
	api.mu.Lock()
	api.sessions[id] = transport
	api.mu.Unlock()
	go api.observeClosure(id, transport)
	writeJSON(writer, http.StatusCreated, previewSessionResponse{SessionID: id, EventsURL: "/sessions/" + id + "/events"})
}

func (api *PreviewAPI) submitRevision(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "sessionID")
	revision, err := strconv.ParseUint(chi.URLParam(request, "revision"), 10, 64)
	if err != nil || revision == 0 {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "revision must be a positive integer"})
		return
	}
	transport, exists := api.lookup(id)
	if !exists {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "preview session not found"})
		return
	}
	var body previewRevisionRequest
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if status := transport.acceptRevision(revision); status != http.StatusAccepted {
		writeJSON(writer, status, map[string]string{"error": "revision must be newer than the last submitted revision"})
		return
	}
	resultChannel := api.engine.Preview(context.WithoutCancel(request.Context()), transport.session, body.Document, rootapi.ProjectionOptions{FocusSections: body.FocusSections})
	go func() {
		result := <-resultChannel
		result.Revision = revision
		if errors.Is(result.Err, preview.ErrSuperseded) || errors.Is(result.Err, preview.ErrClosed) {
			return // canceled older work must never repaint the browser
		}
		result.PDF = nil // browser transport receives changed pages, not the full PDF
		event := PreviewEvent{Type: "result", Revision: revision, Result: &result}
		if result.Err != nil {
			event.Type = "error"
			event.Error = result.Err.Error()
			event.Result.Err = nil
		}
		transport.publish(event)
	}()
	writeJSON(writer, http.StatusAccepted, previewRevisionResponse{SessionID: id, Revision: revision, Accepted: true})
}

func (api *PreviewAPI) streamEvents(writer http.ResponseWriter, request *http.Request) {
	transport, exists := api.lookup(chi.URLParam(request, "sessionID"))
	if !exists {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "preview session not found"})
		return
	}
	flusher, ok := writer.(http.Flusher)
	if !ok {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "streaming unsupported"})
		return
	}
	after := eventCursor(request)
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.WriteHeader(http.StatusOK)
	flusher.Flush()
	for {
		events, wait, closed := transport.eventsAfter(after)
		for _, event := range events {
			data, err := json.Marshal(event)
			if err != nil {
				return
			}
			if _, err := fmt.Fprintf(writer, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.Type, data); err != nil {
				return
			}
			after = event.ID
		}
		if len(events) > 0 {
			flusher.Flush()
		}
		if closed {
			return
		}
		select {
		case <-request.Context().Done():
			return
		case <-wait:
		}
	}
}

func (api *PreviewAPI) closeSession(writer http.ResponseWriter, request *http.Request) {
	transport, exists := api.lookup(chi.URLParam(request, "sessionID"))
	if !exists {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "preview session not found"})
		return
	}
	if err := transport.session.Close(); err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (api *PreviewAPI) lookup(id string) (*previewTransportSession, bool) {
	api.mu.RLock()
	session, exists := api.sessions[id]
	api.mu.RUnlock()
	return session, exists
}

func (api *PreviewAPI) observeClosure(id string, transport *previewTransportSession) {
	<-transport.session.Done()
	transport.publish(PreviewEvent{Type: "closed", Revision: transport.currentRevision()})
	transport.markClosed()
	api.mu.Lock()
	if api.sessions[id] == transport {
		delete(api.sessions, id)
	}
	api.mu.Unlock()
}

func (session *previewTransportSession) acceptRevision(revision uint64) int {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return http.StatusGone
	}
	if revision <= session.lastSubmitted {
		return http.StatusConflict
	}
	session.lastSubmitted = revision
	return http.StatusAccepted
}

func (session *previewTransportSession) publish(event PreviewEvent) {
	session.mu.Lock()
	session.nextEventID++
	event.ID = session.nextEventID
	session.history = append(session.history, event)
	if len(session.history) > session.historyLimit {
		session.history = append([]PreviewEvent(nil), session.history[len(session.history)-session.historyLimit:]...)
	}
	close(session.notify)
	session.notify = make(chan struct{})
	session.mu.Unlock()
}

func (session *previewTransportSession) markClosed() {
	session.mu.Lock()
	session.closed = true
	close(session.notify)
	session.notify = make(chan struct{})
	session.mu.Unlock()
}

func (session *previewTransportSession) currentRevision() uint64 {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.lastSubmitted
}

func (session *previewTransportSession) eventsAfter(after uint64) ([]PreviewEvent, <-chan struct{}, bool) {
	session.mu.Lock()
	defer session.mu.Unlock()
	events := make([]PreviewEvent, 0)
	for _, event := range session.history {
		if event.ID > after {
			events = append(events, event)
		}
	}
	return events, session.notify, session.closed
}

func eventCursor(request *http.Request) uint64 {
	value := request.Header.Get("Last-Event-ID")
	if value == "" {
		value = request.URL.Query().Get("after")
	}
	cursor, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return cursor
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
