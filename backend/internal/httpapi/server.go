package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	case "/api/chapters":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.chapters(w, r)
	case "/api/elections":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.elections(w, r)
	case "/api/races":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.races(w, r)
	case "/api/candidates":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.candidates(w, r)
	case "/api/voter-guides":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.voterGuides(w, r)
	case "/api/endorsements":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.endorsements(w, r)
	case "/api/ballot-measures":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.ballotMeasures(w, r)
	case "/api/measure-positions":
		if r.Method != http.MethodGet {
			s.methodNotAllowed(w)
			return
		}
		s.measurePositions(w, r)
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

func (s *Server) chapters(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Chapters(r.Context())
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, items)
}

func (s *Server) elections(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Elections(r.Context())
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, items)
}

func (s *Server) races(w http.ResponseWriter, r *http.Request) {
	filter, err := parseRaceFilter(r)
	if err != nil {
		s.invalidQuery(w, err.Error())
		return
	}

	items, err := s.store.Races(r.Context(), filter)
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, items)
}

func (s *Server) candidates(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Candidates(r.Context())
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, items)
}

func (s *Server) voterGuides(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.VoterGuides(r.Context(), directory.VoterGuideFilter{
		ChapterID:  r.URL.Query().Get("chapterId"),
		ElectionID: r.URL.Query().Get("electionId"),
	})
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, items)
}

func (s *Server) endorsements(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Endorsements(r.Context(), directory.EndorsementFilter{
		ChapterID:  r.URL.Query().Get("chapterId"),
		ElectionID: r.URL.Query().Get("electionId"),
		RaceID:     r.URL.Query().Get("raceId"),
	})
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, items)
}

func (s *Server) ballotMeasures(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.BallotMeasures(r.Context(), directory.BallotMeasureFilter{
		ElectionID: r.URL.Query().Get("electionId"),
	})
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, items)
}

func (s *Server) measurePositions(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.MeasurePositions(r.Context(), directory.MeasurePositionFilter{
		ChapterID: r.URL.Query().Get("chapterId"),
		MeasureID: r.URL.Query().Get("measureId"),
	})
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, items)
}

func parseRaceFilter(r *http.Request) (directory.RaceFilter, error) {
	level := strings.TrimSpace(r.URL.Query().Get("level"))
	if level != "" {
		switch level {
		case "federal", "statewide", "state_legislature", "local", "unknown":
		default:
			return directory.RaceFilter{}, fmt.Errorf("level must be one of federal, statewide, state_legislature, local, unknown")
		}
	}

	return directory.RaceFilter{
		ElectionID:   r.URL.Query().Get("electionId"),
		Jurisdiction: r.URL.Query().Get("jurisdiction"),
		Level:        level,
	}, nil
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

func (s *Server) invalidQuery(w http.ResponseWriter, message string) {
	s.writeJSON(w, http.StatusBadRequest, map[string]string{
		"error":   "invalid_query",
		"message": message,
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
