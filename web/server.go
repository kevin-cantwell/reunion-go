// Package web provides an HTTP server for exploring a Reunion family file.
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"

	reunion "github.com/kedoco/reunion-explore"
	"github.com/kedoco/reunion-explore/index"
	"github.com/kedoco/reunion-explore/model"
)

//go:embed static/*
var staticFiles embed.FS

// snapshot holds a parsed FamilyFile and its prebuilt index.
type snapshot struct {
	ff  *model.FamilyFile
	idx *index.Index
}

// Server serves the REST API and SPA for a parsed FamilyFile.
type Server struct {
	data   atomic.Pointer[snapshot]
	logger *slog.Logger
}

// New creates a Server, building the index from the FamilyFile.
func New(ff *model.FamilyFile, logger *slog.Logger) *Server {
	s := &Server{logger: logger}
	s.data.Store(&snapshot{ff: ff, idx: index.BuildIndex(ff)})
	return s
}

// load returns the current snapshot.
func (s *Server) load() *snapshot {
	return s.data.Load()
}

// Watch watches the bundle at bundlePath for changes and reloads automatically.
// It blocks until the context is done or an unrecoverable error occurs.
func (s *Server) Watch(bundlePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	// Watch the bundle directory itself.
	absPath, err := filepath.Abs(bundlePath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}
	if err := watcher.Add(absPath); err != nil {
		return fmt.Errorf("watching %s: %w", absPath, err)
	}
	s.logger.Info("watching bundle for changes", "path", absPath)

	// Debounce: Reunion writes multiple files when saving.
	var debounce *time.Timer
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}
			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(500*time.Millisecond, func() {
				s.logger.Info("bundle changed, reloading", "trigger", filepath.Base(event.Name))
				ff, err := reunion.Open(bundlePath, nil)
				if err != nil {
					s.logger.Error("reload failed", "err", err)
					return
				}
				s.data.Store(&snapshot{ff: ff, idx: index.BuildIndex(ff)})
				s.logger.Info("reloaded",
					"persons", len(ff.Persons),
					"families", len(ff.Families),
				)
			})
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			s.logger.Error("watcher error", "err", err)
		}
	}
}

// ListenAndServe starts the HTTP server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}

// Handler returns the http.Handler with all routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/stats", s.handleStats)
	mux.HandleFunc("GET /api/persons", s.handlePersons)
	mux.HandleFunc("GET /api/persons/{id}", s.handlePerson)
	mux.HandleFunc("GET /api/persons/{id}/families", s.handlePersonFamilies)
	mux.HandleFunc("GET /api/persons/{id}/ancestors", s.handlePersonAncestors)
	mux.HandleFunc("GET /api/persons/{id}/descendants", s.handlePersonDescendants)
	mux.HandleFunc("GET /api/persons/{id}/treetops", s.handlePersonTreetops)
	mux.HandleFunc("GET /api/persons/{id}/summary", s.handlePersonSummary)
	mux.HandleFunc("GET /api/families", s.handleFamilies)
	mux.HandleFunc("GET /api/families/{id}", s.handleFamily)
	mux.HandleFunc("GET /api/places", s.handlePlaces)
	mux.HandleFunc("GET /api/places/{id}", s.handlePlace)
	mux.HandleFunc("GET /api/places/{id}/persons", s.handlePlacePersons)
	mux.HandleFunc("GET /api/events", s.handleEvents)
	mux.HandleFunc("GET /api/events/{id}", s.handleEvent)
	mux.HandleFunc("GET /api/events/{id}/persons", s.handleEventPersons)
	mux.HandleFunc("GET /api/sources", s.handleSources)
	mux.HandleFunc("GET /api/sources/{id}", s.handleSource)
	mux.HandleFunc("GET /api/sources/{id}/persons", s.handleSourcePersons)
	mux.HandleFunc("GET /api/notes", s.handleNotes)
	mux.HandleFunc("GET /api/notes/{id}", s.handleNote)
	mux.HandleFunc("GET /api/search", s.handleSearch)
	mux.HandleFunc("GET /api/openapi.json", s.handleOpenAPI)

	// Static files with SPA fallback
	staticFS, _ := fs.Sub(staticFiles, "static")
	fileServer := http.FileServer(http.FS(staticFS))

	mux.HandleFunc("GET /static/", func(w http.ResponseWriter, r *http.Request) {
		// Strip /static/ prefix and serve from embedded FS
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")
		fileServer.ServeHTTP(w, r)
	})

	// SPA fallback: serve index.html for non-API, non-static paths
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/static/") {
			// SPA fallback
		}
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	return s.withMiddleware(mux)
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Logging
		s.logger.Info("request", "method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseID(r *http.Request) (uint32, error) {
	s := r.PathValue("id")
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid ID %q", s)
	}
	return uint32(n), nil
}

func parseIntQuery(r *http.Request, name string, defaultVal int) int {
	s := r.URL.Query().Get(name)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
}
