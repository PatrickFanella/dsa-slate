package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"git.subcult.tv/subculture-collective/dsa-slate/internal/admin"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/app"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/directory"
)

type Server struct {
	store      directory.Store
	adminStore admin.Store
	adminToken string
}

func NewServer(store directory.Store) http.Handler {
	return &Server{store: store}
}

func NewServerWithAdmin(store directory.Store, adminStore admin.Store, adminToken string) http.Handler {
	return &Server{
		store:      store,
		adminStore: adminStore,
		adminToken: strings.TrimSpace(adminToken),
	}
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
		if strings.HasPrefix(r.URL.Path, "/api/admin/") {
			s.handleAdmin(w, r)
			return
		}
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

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	if s.adminStore == nil || s.adminToken == "" {
		s.notFound(w)
		return
	}
	if !s.requireAdminAuth(w, r) {
		return
	}

	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/admin"))
	if len(parts) == 0 {
		s.notFound(w)
		return
	}

	switch parts[0] {
	case "chapter-drafts":
		s.handleChapterDraftRoute(w, r, parts)
	case "endorsement-drafts":
		s.handleEndorsementDraftRoute(w, r, parts)
	default:
		s.notFound(w)
	}
}

func (s *Server) handleChapterDraftRoute(w http.ResponseWriter, r *http.Request, parts []string) {
	switch len(parts) {
	case 1:
		s.handleChapterDraftCollection(w, r)
	case 2:
		s.handleChapterDraftItem(w, r, parts[1])
	case 3:
		s.handleChapterDraftAction(w, r, parts[1], parts[2])
	default:
		s.notFound(w)
	}
}

func (s *Server) handleEndorsementDraftRoute(w http.ResponseWriter, r *http.Request, parts []string) {
	switch len(parts) {
	case 1:
		s.handleEndorsementDraftCollection(w, r)
	case 2:
		s.handleEndorsementDraftItem(w, r, parts[1])
	case 3:
		s.handleEndorsementDraftAction(w, r, parts[1], parts[2])
	default:
		s.notFound(w)
	}
}

func (s *Server) handleChapterDraftCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		drafts, err := s.adminStore.ListChapterDrafts(r.Context())
		if err != nil {
			s.internalServerError(w)
			return
		}
		s.writeJSON(w, http.StatusOK, drafts)
	case http.MethodPost:
		draft, ok := s.decodeChapterDraft(w, r)
		if !ok {
			return
		}
		if errs := admin.ValidateChapterDraft(draft); len(errs) > 0 {
			s.writeValidationFailed(w, errs)
			return
		}
		saved, err := s.adminStore.UpsertChapterDraft(r.Context(), draft, s.adminActor())
		if err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, saved)
	default:
		s.methodNotAllowedAllow(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) handleChapterDraftItem(w http.ResponseWriter, r *http.Request, id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		s.notFound(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		draft, err := s.adminStore.GetChapterDraft(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.notFound(w)
				return
			}
			s.internalServerError(w)
			return
		}
		s.writeJSON(w, http.StatusOK, draft)
	case http.MethodPut:
		draft, ok := s.decodeChapterDraft(w, r)
		if !ok {
			return
		}
		if draft.ID != "" && strings.TrimSpace(draft.ID) != id {
			s.writeValidationFailed(w, []string{"id must match the draft id in the URL"})
			return
		}
		draft.ID = id
		if errs := admin.ValidateChapterDraft(draft); len(errs) > 0 {
			s.writeValidationFailed(w, errs)
			return
		}
		saved, err := s.adminStore.UpsertChapterDraft(r.Context(), draft, s.adminActor())
		if err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, saved)
	default:
		s.methodNotAllowedAllow(w, http.MethodGet, http.MethodPut)
	}
}

func (s *Server) handleChapterDraftAction(w http.ResponseWriter, r *http.Request, id string, action string) {
	id = strings.TrimSpace(id)
	if id == "" || r.Method != http.MethodPost {
		s.methodNotAllowedAllow(w, http.MethodPost)
		return
	}

	switch action {
	case "publish":
		if err := s.adminStore.PublishChapterDraft(r.Context(), id, s.adminActor()); err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
	case "archive":
		if err := s.adminStore.ArchiveChapterDraft(r.Context(), id, s.adminActor()); err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
	default:
		s.notFound(w)
	}
}

func (s *Server) handleEndorsementDraftCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		drafts, err := s.adminStore.ListEndorsementDrafts(r.Context())
		if err != nil {
			s.internalServerError(w)
			return
		}
		s.writeJSON(w, http.StatusOK, drafts)
	case http.MethodPost:
		draft, ok := s.decodeEndorsementDraft(w, r)
		if !ok {
			return
		}
		if errs := admin.ValidateEndorsementDraft(draft); len(errs) > 0 {
			s.writeValidationFailed(w, errs)
			return
		}
		saved, err := s.adminStore.UpsertEndorsementDraft(r.Context(), draft, s.adminActor())
		if err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, saved)
	default:
		s.methodNotAllowedAllow(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) handleEndorsementDraftItem(w http.ResponseWriter, r *http.Request, id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		s.notFound(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		draft, err := s.adminStore.GetEndorsementDraft(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.notFound(w)
				return
			}
			s.internalServerError(w)
			return
		}
		s.writeJSON(w, http.StatusOK, draft)
	case http.MethodPut:
		draft, ok := s.decodeEndorsementDraft(w, r)
		if !ok {
			return
		}
		if draft.ID != "" && strings.TrimSpace(draft.ID) != id {
			s.writeValidationFailed(w, []string{"id must match the draft id in the URL"})
			return
		}
		draft.ID = id
		if errs := admin.ValidateEndorsementDraft(draft); len(errs) > 0 {
			s.writeValidationFailed(w, errs)
			return
		}
		saved, err := s.adminStore.UpsertEndorsementDraft(r.Context(), draft, s.adminActor())
		if err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, saved)
	default:
		s.methodNotAllowedAllow(w, http.MethodGet, http.MethodPut)
	}
}

func (s *Server) handleEndorsementDraftAction(w http.ResponseWriter, r *http.Request, id string, action string) {
	id = strings.TrimSpace(id)
	if id == "" || r.Method != http.MethodPost {
		s.methodNotAllowedAllow(w, http.MethodPost)
		return
	}

	switch action {
	case "publish":
		if err := s.adminStore.PublishEndorsementDraft(r.Context(), id, s.adminActor()); err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
	case "archive":
		if err := s.adminStore.ArchiveEndorsementDraft(r.Context(), id, s.adminActor()); err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
	default:
		s.notFound(w)
	}
}

func (s *Server) decodeChapterDraft(w http.ResponseWriter, r *http.Request) (admin.ChapterDraft, bool) {
	var draft admin.ChapterDraft
	if err := decodeStrictJSON(r, &draft); err != nil {
		s.writeValidationFailed(w, []string{"invalid JSON body"})
		return admin.ChapterDraft{}, false
	}
	return draft, true
}

func (s *Server) decodeEndorsementDraft(w http.ResponseWriter, r *http.Request) (admin.EndorsementDraft, bool) {
	var draft admin.EndorsementDraft
	if err := decodeStrictJSON(r, &draft); err != nil {
		s.writeValidationFailed(w, []string{"invalid JSON body"})
		return admin.EndorsementDraft{}, false
	}
	return draft, true
}

func decodeStrictJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	return nil
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func (s *Server) requireAdminAuth(w http.ResponseWriter, r *http.Request) bool {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		s.unauthorized(w)
		return false
	}
	if strings.TrimSpace(strings.TrimPrefix(header, prefix)) != s.adminToken {
		s.unauthorized(w)
		return false
	}
	return true
}

func (s *Server) adminActor() string {
	return "admin"
}

func (s *Server) writeValidationFailed(w http.ResponseWriter, messages []string) {
	s.writeJSON(w, http.StatusBadRequest, map[string]any{
		"error":    "validation_failed",
		"messages": messages,
	})
}

func (s *Server) unauthorized(w http.ResponseWriter) {
	s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func (s *Server) writeAdminStoreError(w http.ResponseWriter, err error) {
	var validationErr admin.ValidationError
	if errors.As(err, &validationErr) {
		s.writeValidationFailed(w, validationErr.Messages)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		s.notFound(w)
		return
	}
	if strings.EqualFold(err.Error(), "not found") {
		s.notFound(w)
		return
	}
	s.internalServerError(w)
}

func (s *Server) methodNotAllowedAllow(w http.ResponseWriter, methods ...string) {
	if len(methods) == 0 {
		s.methodNotAllowed(w)
		return
	}
	w.Header().Set("Allow", strings.Join(methods, ", "))
	s.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
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
