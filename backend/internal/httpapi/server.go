package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/dsa-slate/internal/admin"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/app"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/auth"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/civic"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/directory"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/savedslate"
)

type Server struct {
	store        directory.Store
	authStore    auth.Store
	googleAuth   auth.GoogleSignInService
	adminStore   admin.Store
	slateStore   savedslate.Store
	adminToken   string
	civicService civic.LookupService
}

// NewServer creates a public-read server without auth or admin dependencies.
func NewServer(store directory.Store) http.Handler {
	return &Server{store: store}
}

// NewServerWithAuth creates a server with auth session resolution enabled.
func NewServerWithAuth(store directory.Store, authStore auth.Store) http.Handler {
	return NewServerWithAuthAndServices(store, authStore, nil, nil, nil, "", nil)
}

// NewServerWithAdmin creates a server with admin routes enabled.
func NewServerWithAdmin(store directory.Store, adminStore admin.Store, adminToken string) http.Handler {
	return NewServerWithAuthAndServices(store, nil, nil, adminStore, nil, adminToken, nil)
}

// NewServerWithServices creates a server with admin and civic dependencies.
func NewServerWithServices(store directory.Store, adminStore admin.Store, adminToken string, civicService civic.LookupService) http.Handler {
	return NewServerWithAuthAndServices(store, nil, nil, adminStore, nil, adminToken, civicService)
}

// NewServerWithAuthAndServices creates a server with auth, admin, saved slate, and civic dependencies.
func NewServerWithAuthAndServices(store directory.Store, authStore auth.Store, googleAuth auth.GoogleSignInService, adminStore admin.Store, slateStore savedslate.Store, adminToken string, civicService civic.LookupService) http.Handler {
	return &Server{
		store:        store,
		authStore:    authStore,
		googleAuth:   googleAuth,
		adminStore:   adminStore,
		slateStore:   slateStore,
		adminToken:   strings.TrimSpace(adminToken),
		civicService: civicService,
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
	case "/api/civic/lookup":
		if r.Method != http.MethodPost {
			s.methodNotAllowedAllow(w, http.MethodPost)
			return
		}
		s.civicLookup(w, r)
	case "/api/auth/session":
		if r.Method != http.MethodGet {
			s.methodNotAllowedAllow(w, http.MethodGet)
			return
		}
		s.authSession(w, r)
	case "/api/auth/google/start":
		if r.Method != http.MethodGet {
			s.methodNotAllowedAllow(w, http.MethodGet)
			return
		}
		s.authGoogleStart(w, r)
	case "/api/auth/google/callback":
		if r.Method != http.MethodGet {
			s.methodNotAllowedAllow(w, http.MethodGet)
			return
		}
		s.authGoogleCallback(w, r)
	case "/api/auth/logout":
		if r.Method != http.MethodPost {
			s.methodNotAllowedAllow(w, http.MethodPost)
			return
		}
		s.authLogout(w, r)
	case "/api/me/slate":
		if r.Method != http.MethodGet {
			s.methodNotAllowedAllow(w, http.MethodGet)
			return
		}
		s.savedSlate(w, r)
	case "/api/me/slate/items":
		if r.Method != http.MethodPut {
			s.methodNotAllowedAllow(w, http.MethodPut)
			return
		}
		s.savedSlateItemCollection(w, r)
	case "/api/civic/elections":
		if r.Method != http.MethodGet {
			s.methodNotAllowedAllow(w, http.MethodGet)
			return
		}
		s.civicElections(w, r)
	default:
		if strings.HasPrefix(r.URL.Path, "/api/me/slate/items/") {
			s.savedSlateItem(w, r)
			return
		}
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

func (s *Server) civicLookup(w http.ResponseWriter, r *http.Request) {
	if s.civicService == nil {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "civic_not_configured"})
		return
	}

	request, ok := s.decodeCivicLookup(w, r)
	if !ok {
		return
	}
	request.Address = strings.TrimSpace(request.Address)
	if request.Address == "" {
		s.writeValidationFailed(w, []string{"address is required"})
		return
	}

	response, err := s.civicService.Lookup(r.Context(), request)
	if err != nil {
		statusCode := http.StatusBadGateway
		message := "Google Civic lookup failed"
		var upstreamErr civic.UpstreamError
		if errors.As(err, &upstreamErr) {
			message = upstreamErr.Message
			if upstreamErr.StatusCode >= http.StatusBadRequest && upstreamErr.StatusCode < http.StatusInternalServerError {
				statusCode = http.StatusBadRequest
			}
		}
		s.writeJSON(w, statusCode, map[string]string{
			"error":   "civic_lookup_failed",
			"message": message,
		})
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) civicElections(w http.ResponseWriter, r *http.Request) {
	if s.civicService == nil {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "civic_not_configured"})
		return
	}

	response, err := s.civicService.Elections(r.Context())
	if err != nil {
		statusCode := http.StatusBadGateway
		message := "Google Civic elections lookup failed"
		var upstreamErr civic.UpstreamError
		if errors.As(err, &upstreamErr) {
			message = upstreamErr.Message
			if upstreamErr.StatusCode >= http.StatusBadRequest && upstreamErr.StatusCode < http.StatusInternalServerError {
				statusCode = http.StatusBadRequest
			}
		}
		s.writeJSON(w, statusCode, map[string]string{"error": "civic_lookup_failed", "message": message})
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) authSession(w http.ResponseWriter, r *http.Request) {
	if s.authStore == nil {
		s.writeJSON(w, http.StatusOK, auth.State{})
		return
	}

	cookie, err := r.Cookie(auth.SessionCookieName)
	if err != nil {
		s.writeJSON(w, http.StatusOK, auth.State{})
		return
	}

	state, err := auth.ResolveState(r.Context(), s.authStore, cookie.Value, timeNow())
	if err != nil {
		s.internalServerError(w)
		return
	}

	s.writeJSON(w, http.StatusOK, state)
}

func (s *Server) authGoogleStart(w http.ResponseWriter, r *http.Request) {
	if s.googleAuth == nil {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth_not_configured"})
		return
	}
	s.googleAuth.Start(w, r)
}

func (s *Server) authGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if s.googleAuth == nil {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth_not_configured"})
		return
	}
	s.googleAuth.Callback(w, r)
}

func (s *Server) authLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
		Secure:   strings.HasPrefix(app.AppBaseURL(), "https://"),
	})

	if s.authStore != nil {
		if cookie, err := r.Cookie(auth.SessionCookieName); err == nil && strings.TrimSpace(cookie.Value) != "" {
			if err := s.authStore.DeleteSession(r.Context(), cookie.Value); err != nil {
				s.internalServerError(w)
				return
			}
		}
	}

	s.writeJSON(w, http.StatusOK, auth.State{})
}

func (s *Server) savedSlate(w http.ResponseWriter, r *http.Request) {
	if s.slateStore == nil {
		s.notFound(w)
		return
	}
	user, ok := s.requireUserAuth(w, r)
	if !ok {
		return
	}
	items, err := s.slateStore.ListForUser(r.Context(), user.ID)
	if err != nil {
		s.internalServerError(w)
		return
	}
	s.writeJSON(w, http.StatusOK, savedslate.Slate{Items: items})
}

func (s *Server) savedSlateItemCollection(w http.ResponseWriter, r *http.Request) {
	if s.slateStore == nil {
		s.notFound(w)
		return
	}
	user, ok := s.requireUserAuth(w, r)
	if !ok {
		return
	}
	request, ok := s.decodeSavedSlateItem(w, r)
	if !ok {
		return
	}
	item, err := s.slateStore.UpsertForUser(r.Context(), user.ID, request.EndorsementID)
	if err != nil {
		s.writeSavedSlateStoreError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, item)
}

func (s *Server) savedSlateItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		s.methodNotAllowedAllow(w, http.MethodDelete)
		return
	}
	if s.slateStore == nil {
		s.notFound(w)
		return
	}
	user, ok := s.requireUserAuth(w, r)
	if !ok {
		return
	}
	itemID := strings.TrimPrefix(r.URL.Path, "/api/me/slate/items/")
	if strings.TrimSpace(itemID) == "" || strings.Contains(itemID, "/") {
		s.notFound(w)
		return
	}
	if err := s.slateStore.DeleteForUser(r.Context(), user.ID, itemID); err != nil {
		s.writeSavedSlateStoreError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type savedSlateItemRequest struct {
	EndorsementID string `json:"endorsementId"`
}

func (s *Server) decodeSavedSlateItem(w http.ResponseWriter, r *http.Request) (savedSlateItemRequest, bool) {
	var request savedSlateItemRequest
	if err := decodeStrictJSON(r, &request); err != nil {
		s.writeValidationFailed(w, []string{"invalid JSON body"})
		return savedSlateItemRequest{}, false
	}
	request.EndorsementID = strings.TrimSpace(request.EndorsementID)
	if request.EndorsementID == "" {
		s.writeValidationFailed(w, []string{"endorsementId is required"})
		return savedSlateItemRequest{}, false
	}
	return request, true
}

func (s *Server) decodeCivicLookup(w http.ResponseWriter, r *http.Request) (civic.LookupRequest, bool) {
	var request civic.LookupRequest
	if err := decodeStrictJSON(r, &request); err != nil {
		s.writeValidationFailed(w, []string{"invalid JSON body"})
		return civic.LookupRequest{}, false
	}
	return request, true
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	if s.adminStore == nil {
		s.notFound(w)
		return
	}
	actor, ok := s.requireAdminAuth(w, r)
	if !ok {
		return
	}

	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/admin"))
	if len(parts) == 0 {
		s.notFound(w)
		return
	}

	switch parts[0] {
	case "chapter-drafts":
		s.handleChapterDraftRoute(w, r, parts, actor)
	case "endorsement-drafts":
		s.handleEndorsementDraftRoute(w, r, parts, actor)
	default:
		s.notFound(w)
	}
}

func (s *Server) handleChapterDraftRoute(w http.ResponseWriter, r *http.Request, parts []string, actor string) {
	switch len(parts) {
	case 1:
		s.handleChapterDraftCollection(w, r, actor)
	case 2:
		s.handleChapterDraftItem(w, r, parts[1], actor)
	case 3:
		s.handleChapterDraftAction(w, r, parts[1], parts[2], actor)
	default:
		s.notFound(w)
	}
}

func (s *Server) handleEndorsementDraftRoute(w http.ResponseWriter, r *http.Request, parts []string, actor string) {
	switch len(parts) {
	case 1:
		s.handleEndorsementDraftCollection(w, r, actor)
	case 2:
		s.handleEndorsementDraftItem(w, r, parts[1], actor)
	case 3:
		s.handleEndorsementDraftAction(w, r, parts[1], parts[2], actor)
	default:
		s.notFound(w)
	}
}

func (s *Server) handleChapterDraftCollection(w http.ResponseWriter, r *http.Request, actor string) {
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
		saved, err := s.adminStore.UpsertChapterDraft(r.Context(), draft, actor)
		if err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, saved)
	default:
		s.methodNotAllowedAllow(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) handleChapterDraftItem(w http.ResponseWriter, r *http.Request, id string, actor string) {
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
		saved, err := s.adminStore.UpsertChapterDraft(r.Context(), draft, actor)
		if err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, saved)
	default:
		s.methodNotAllowedAllow(w, http.MethodGet, http.MethodPut)
	}
}

func (s *Server) handleChapterDraftAction(w http.ResponseWriter, r *http.Request, id string, action string, actor string) {
	id = strings.TrimSpace(id)
	if id == "" || r.Method != http.MethodPost {
		s.methodNotAllowedAllow(w, http.MethodPost)
		return
	}

	switch action {
	case "publish":
		if err := s.adminStore.PublishChapterDraft(r.Context(), id, actor); err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
	case "archive":
		if err := s.adminStore.ArchiveChapterDraft(r.Context(), id, actor); err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
	default:
		s.notFound(w)
	}
}

func (s *Server) handleEndorsementDraftCollection(w http.ResponseWriter, r *http.Request, actor string) {
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
		saved, err := s.adminStore.UpsertEndorsementDraft(r.Context(), draft, actor)
		if err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, saved)
	default:
		s.methodNotAllowedAllow(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) handleEndorsementDraftItem(w http.ResponseWriter, r *http.Request, id string, actor string) {
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
		saved, err := s.adminStore.UpsertEndorsementDraft(r.Context(), draft, actor)
		if err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, saved)
	default:
		s.methodNotAllowedAllow(w, http.MethodGet, http.MethodPut)
	}
}

func (s *Server) handleEndorsementDraftAction(w http.ResponseWriter, r *http.Request, id string, action string, actor string) {
	id = strings.TrimSpace(id)
	if id == "" || r.Method != http.MethodPost {
		s.methodNotAllowedAllow(w, http.MethodPost)
		return
	}

	switch action {
	case "publish":
		if err := s.adminStore.PublishEndorsementDraft(r.Context(), id, actor); err != nil {
			s.writeAdminStoreError(w, err)
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
	case "archive":
		if err := s.adminStore.ArchiveEndorsementDraft(r.Context(), id, actor); err != nil {
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

func (s *Server) requireUserAuth(w http.ResponseWriter, r *http.Request) (*auth.Summary, bool) {
	if s.authStore == nil {
		s.unauthorized(w)
		return nil, false
	}
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		s.unauthorized(w)
		return nil, false
	}
	state, err := auth.ResolveState(r.Context(), s.authStore, cookie.Value, timeNow())
	if err != nil {
		s.internalServerError(w)
		return nil, false
	}
	if !state.Authenticated || state.User == nil || strings.TrimSpace(state.User.ID) == "" {
		s.unauthorized(w)
		return nil, false
	}
	return state.User, true
}

func (s *Server) requireAdminAuth(w http.ResponseWriter, r *http.Request) (string, bool) {
	if s.authStore != nil {
		cookie, err := r.Cookie(auth.SessionCookieName)
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			s.unauthorized(w)
			return "", false
		}

		state, err := auth.ResolveState(r.Context(), s.authStore, cookie.Value, timeNow())
		if err != nil {
			s.internalServerError(w)
			return "", false
		}
		if !state.Authenticated || state.User == nil {
			s.unauthorized(w)
			return "", false
		}
		if state.User.Role != auth.RoleEditor && state.User.Role != auth.RoleAdmin {
			s.forbidden(w)
			return "", false
		}
		return authActor(state.User), true
	}

	header := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if s.adminToken == "" || !strings.HasPrefix(header, prefix) {
		s.unauthorized(w)
		return "", false
	}
	if strings.TrimSpace(strings.TrimPrefix(header, prefix)) != s.adminToken {
		s.unauthorized(w)
		return "", false
	}
	return "admin-token", true
}

func authActor(user *auth.Summary) string {
	if user == nil {
		return "unknown"
	}
	if email := strings.TrimSpace(user.Email); email != "" {
		return email
	}
	if id := strings.TrimSpace(user.ID); id != "" {
		return id
	}
	return "unknown"
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

func (s *Server) forbidden(w http.ResponseWriter) {
	s.writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden", "message": "editor or admin role required"})
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

func (s *Server) writeSavedSlateStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, savedslate.ErrInvalidReference) {
		s.writeValidationFailed(w, []string{"endorsementId must reference an existing endorsement"})
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
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

func timeNow() time.Time {
	return time.Now().UTC()
}
