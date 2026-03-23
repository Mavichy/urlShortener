package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Mavichy/urlShortener/internal/model"
	"github.com/Mavichy/urlShortener/internal/service"
	"github.com/Mavichy/urlShortener/internal/storage"
	"log"
	"net/http"
	"strings"
	"time"
)

type Shortener interface {
	Create(ctx context.Context, originalURL string) (model.Link, error)
	Resolve(ctx context.Context, code string) (model.Link, error)
}

type Server struct {
	httpServer     *http.Server
	requestTimeout time.Duration
	baseURL        string
	service        Shortener
}

type createLinkRequest struct {
	URL string `json:"url"`
}

type linkResponse struct {
	Code        string `json:"code"`
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

type ServerConfig struct {
	Addr           string
	BaseURL        string
	RequestTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
}

func New(cfg ServerConfig, svc Shortener) *Server {
	s := &Server{
		requestTimeout: cfg.RequestTimeout,
		baseURL:        strings.TrimRight(cfg.BaseURL, "/"),
		service:        svc,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/links", s.handleCreate)
	mux.HandleFunc("GET /{code}", s.handleResolve)

	s.httpServer = &http.Server{
		Addr:         cfg.Addr,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return s
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.requestTimeout)
	defer cancel()

	var req createLinkRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	link, err := s.service.Create(ctx, req.URL)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, linkResponse{
		Code:        link.Code,
		ShortURL:    s.buildShortURL(link.Code),
		OriginalURL: link.OriginalURL,
	})
}

func (s *Server) handleResolve(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.requestTimeout)
	defer cancel()

	code := strings.TrimSpace(r.PathValue("code"))
	if code == "" {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}

	link, err := s.service.Resolve(ctx, code)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, linkResponse{
		Code:        link.Code,
		ShortURL:    s.buildShortURL(link.Code),
		OriginalURL: link.OriginalURL,
	})
}

func (s *Server) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidURL), errors.Is(err, service.ErrInvalidCode):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, storage.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func (s *Server) buildShortURL(code string) string {
	return s.baseURL + "/" + code
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(sw, r)
		log.Printf("method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}
