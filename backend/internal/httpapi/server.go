package httpapi

import (
	"encoding/json"
	"net/http"

	"git.subcult.tv/subculture-collective/dsa-slate/internal/app"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/directory"
)

type Server struct {
	store directory.Store
}

func NewServer(store directory.Store) http.Handler {
	return &Server{store: store}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/healthz":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.healthz(w, r)
	case "/api/directory":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.directory(w, r)
	default:
		s.notFound(w)
	}
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": app.ServiceName(),
	})
}

func (s *Server) directory(w http.ResponseWriter, r *http.Request) {
	data, err := s.store.Directory(r.Context())
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, directory.NormalizeDirectoryData(data))
}

func (s *Server) notFound(w http.ResponseWriter) {
	s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
}

func (s *Server) methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", http.MethodGet)
	s.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
}

func (s *Server) internalServerError(w http.ResponseWriter) {
	s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
