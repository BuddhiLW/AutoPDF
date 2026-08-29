package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

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

// PreviewAPIOptions configures the browser-facing streaming transports.
type PreviewAPIOptions struct {
	Engine          *rootapi.DocumentEngine
	CompilerFactory PreviewCompilerFactory
	SessionOptions  preview.Options
	HistoryLimit    int

	// Heartbeat is how often an idle stream signals liveness. Defaults to
	// DefaultStreamHeartbeat.
	Heartbeat time.Duration
	// RetryHint tells SSE clients how long to wait before reconnecting.
	// Defaults to DefaultStreamRetry.
	RetryHint time.Duration
	// OriginPatterns lists host patterns permitted to open a WebSocket. Empty
	// means same-origin only, which is what a browser preview needs.
	OriginPatterns []string
	// RevisionQueueDepth bounds revisions awaiting publication per session.
	// Defaults to DefaultRevisionQueueDepth.
	RevisionQueueDepth int
}

// PreviewAPI owns revisioned preview sessions. Mount Routes at a stable prefix,
// for example /api/v1/preview.
type PreviewAPI struct {
	engine          *rootapi.DocumentEngine
	compilerFactory PreviewCompilerFactory
	sessionOptions  preview.Options
	historyLimit    int
	heartbeat       time.Duration
	retryHint       time.Duration
	originPatterns  []string
	queueDepth      int

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
	if options.Heartbeat <= 0 {
		options.Heartbeat = DefaultStreamHeartbeat
	}
	if options.RetryHint <= 0 {
		options.RetryHint = DefaultStreamRetry
	}
	if options.RevisionQueueDepth <= 0 {
		options.RevisionQueueDepth = DefaultRevisionQueueDepth
	}
	return &PreviewAPI{
		engine: options.Engine, compilerFactory: options.CompilerFactory,
		sessionOptions: options.SessionOptions, historyLimit: options.HistoryLimit,
		heartbeat: options.Heartbeat, retryHint: options.RetryHint,
		originPatterns: options.OriginPatterns,
		queueDepth:     options.RevisionQueueDepth,
		sessions:       make(map[string]*previewTransportSession),
	}, nil
}

// Routes exposes creation, monotonic revision submission, replayable SSE
// events, cancellation, and closure.
func (api *PreviewAPI) Routes() chi.Router {
	router := chi.NewRouter()
	router.Post("/sessions", api.createSession)
	router.Put("/sessions/{sessionID}/revisions/{revision}", api.submitRevision)
	router.Get("/sessions/{sessionID}/events", api.streamEvents)
	router.Get("/sessions/{sessionID}/ws", api.streamWebSocket)
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

// DefaultRevisionQueueDepth bounds how many submitted revisions may await
// publication per session. It is the backpressure point: revision submission is
// client-driven, so without a bound a fast typist's keystrokes become unbounded
// server-side work.
const DefaultRevisionQueueDepth = 64

// pendingRevision is one submitted revision awaiting its compilation result.
type pendingRevision struct {
	revision uint64
	results  <-chan preview.Result
}

type previewTransportSession struct {
	session      *preview.Session
	historyLimit int

	// pending is drained in order by exactly one publisher goroutine per
	// session, so goroutine count tracks sessions rather than revisions.
	pending chan pendingRevision

	mu            sync.Mutex
	lastSubmitted uint64
	nextEventID   uint64
	history       []PreviewEvent
	closed        bool
	pendingClosed bool
	notify        chan struct{}
}

// enqueue offers a revision to the session's publisher, reporting false when
// the queue is full rather than blocking the HTTP handler or spawning work.
// Holding mu across the offer is what makes it safe against closePending: a
// send racing a close would panic.
func (session *previewTransportSession) enqueue(revision pendingRevision) bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.pendingClosed {
		return false
	}
	select {
	case session.pending <- revision:
		return true
	default:
		return false
	}
}

// closePending ends the publisher goroutine. Idempotent.
func (session *previewTransportSession) closePending() {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.pendingClosed {
		return
	}
	session.pendingClosed = true
	close(session.pending)
}

// publishResults drains submitted revisions in order and turns each into an
// event. One goroutine per session, started at creation and ended when the
// session closes.
func (session *previewTransportSession) publishResults() {
	for entry := range session.pending {
		result := <-entry.results
		result.Revision = entry.revision
		if errors.Is(result.Err, preview.ErrSuperseded) || errors.Is(result.Err, preview.ErrClosed) {
			continue // canceled older work must never repaint the browser
		}
		result.PDF = nil // browser transport receives changed pages, not the full PDF
		event := PreviewEvent{Type: "result", Revision: entry.revision, Result: &result}
		if result.Err != nil {
			event.Type = "error"
			event.Error = result.Err.Error()
			event.Result.Err = nil
		}
		session.publish(event)
	}
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
	transport := &previewTransportSession{
		session:      session,
		historyLimit: api.historyLimit,
		notify:       make(chan struct{}),
		pending:      make(chan pendingRevision, api.queueDepth),
	}
	api.mu.Lock()
	api.sessions[id] = transport
	api.mu.Unlock()
	go transport.publishResults()
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
	if !transport.enqueue(pendingRevision{revision: revision, results: resultChannel}) {
		writeJSON(writer, http.StatusTooManyRequests, map[string]string{
			"error": "preview queue is full; slow down revision submission",
		})
		return
	}
	writeJSON(writer, http.StatusAccepted, previewRevisionResponse{SessionID: id, Revision: revision, Accepted: true})
}

func (api *PreviewAPI) streamEvents(writer http.ResponseWriter, request *http.Request) {
	transport, exists := api.lookup(chi.URLParam(request, "sessionID"))
	if !exists {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "preview session not found"})
		return
	}
	sink, err := newSSESink(writer, api.retryHint)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "streaming unsupported"})
		return
	}
	_ = pumpEvents(request.Context(), transport, sink, streamOptions{
		after:     eventCursor(request),
		heartbeat: api.heartbeat,
	})
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
	transport.closePending() // ends the publisher goroutine
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

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
