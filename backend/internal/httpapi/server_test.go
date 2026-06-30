package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/dsa-slate/internal/admin"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/civic"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/directory"
)

func TestHealthz(t *testing.T) {
	handler := NewServer(directory.NewMemoryStore(directory.DirectoryData{}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got["ok"] != true || got["service"] != "dsa-slate" {
		t.Fatalf("unexpected body: %#v", got)
	}
}

func TestDirectoryEmptyCollectionsAreArrays(t *testing.T) {
	handler := NewServer(directory.NewMemoryStore(directory.DirectoryData{}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/directory", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	for _, key := range []string{"chapters", "elections", "races", "candidates", "voterGuides", "endorsements", "ballotMeasures", "measurePositions"} {
		items, ok := got[key].([]any)
		if !ok {
			t.Fatalf("%s type = %T, want []", key, got[key])
		}
		if items == nil || len(items) != 0 {
			t.Fatalf("%s = %#v, want empty array", key, items)
		}
	}
}

func TestDirectoryUsesCamelCaseFieldNames(t *testing.T) {
	handler := NewServer(directory.NewMemoryStore(directory.DirectoryData{
		Chapters: []directory.Chapter{{
			ID:               "chap-1",
			Name:             "Chapter 1",
			State:            "CA",
			Status:           "Chapter",
			GroupEmail:       "chapter@example.org",
			InstagramURL:     "https://www.instagram.com/chapter",
			EventCalendarURL: "https://example.org/calendar",
			Active:           false,
		}},
		VoterGuides: []directory.VoterGuide{{
			ID:          "guide-1",
			Title:       "Guide 1",
			ChapterID:   "chap-1",
			SourceURL:   "https://example.com/guide",
			Confidence:  "needs_review",
			ReviewedAt:  "2026-06-28",
			ReviewNotes: "Manual review documented no visible source date.",
		}},
		Candidates: []directory.Candidate{{ID: "cand-1", Name: "Candidate 1", Incumbent: false}},
		MeasurePositions: []directory.MeasurePosition{{
			ID:         "pos-1",
			MeasureID:  "meas-1",
			ChapterID:  "chap-1",
			Position:   "yes",
			SourceURL:  "https://example.com/position",
			Confidence: "needs_review",
		}},
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/directory", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if _, ok := got["voterGuides"]; !ok {
		t.Fatalf("missing voterGuides key: %#v", got)
	}
	if _, ok := got["measurePositions"]; !ok {
		t.Fatalf("missing measurePositions key: %#v", got)
	}
	if _, ok := got["voter_guides"]; ok {
		t.Fatalf("unexpected snake_case key: %#v", got)
	}

	voterGuides := got["voterGuides"].([]any)
	guide := voterGuides[0].(map[string]any)
	if _, ok := guide["sourceUrl"]; !ok {
		t.Fatalf("missing sourceUrl in voter guide: %#v", guide)
	}
	if _, ok := guide["source_url"]; ok {
		t.Fatalf("unexpected snake_case source_url in voter guide: %#v", guide)
	}
	if got := guide["reviewedAt"]; got != "2026-06-28" {
		t.Fatalf("reviewedAt = %#v, want 2026-06-28", got)
	}
	if _, ok := guide["reviewed_at"]; ok {
		t.Fatalf("unexpected snake_case reviewed_at in voter guide: %#v", guide)
	}
	if got := guide["reviewNotes"]; got != "Manual review documented no visible source date." {
		t.Fatalf("reviewNotes = %#v, want documented review note", got)
	}
	if _, ok := guide["review_notes"]; ok {
		t.Fatalf("unexpected snake_case review_notes in voter guide: %#v", guide)
	}

	chapters := got["chapters"].([]any)
	chapter := chapters[0].(map[string]any)
	if active, ok := chapter["active"].(bool); !ok || active {
		t.Fatalf("active = %#v, want false", chapter["active"])
	}
	if got := chapter["groupEmail"]; got != "chapter@example.org" {
		t.Fatalf("groupEmail = %#v, want chapter@example.org", got)
	}
	if _, ok := chapter["group_email"]; ok {
		t.Fatalf("unexpected snake_case group_email in chapter: %#v", chapter)
	}
	if got := chapter["eventCalendarUrl"]; got != "https://example.org/calendar" {
		t.Fatalf("eventCalendarUrl = %#v, want https://example.org/calendar", got)
	}

	candidates := got["candidates"].([]any)
	candidate := candidates[0].(map[string]any)
	if incumbent, ok := candidate["incumbent"].(bool); !ok || incumbent {
		t.Fatalf("incumbent = %#v, want false", candidate["incumbent"])
	}
}

func TestCollectionEndpointReturnsArrays(t *testing.T) {
	handler := NewServer(directory.NewMemoryStore(directory.DirectoryData{}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/chapters", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got []any
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got == nil || len(got) != 0 {
		t.Fatalf("body = %#v, want empty array", got)
	}
}

func TestRacesFilteredByQueryParams(t *testing.T) {
	handler := NewServer(directory.NewMemoryStore(directory.DirectoryData{
		Races: []directory.Race{
			{ID: "race-1", ElectionID: "election-1", Office: "Mayor", Level: "statewide", Jurisdiction: "California"},
			{ID: "race-2", ElectionID: "election-2", Office: "Governor", Level: "statewide", Jurisdiction: "California"},
			{ID: "race-3", ElectionID: "election-1", Office: "City Council", Level: "local", Jurisdiction: "California"},
		},
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/races?electionId=election-1&level=statewide&jurisdiction=California", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got []directory.Race
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != "race-1" {
		t.Fatalf("race id = %q, want %q", got[0].ID, "race-1")
	}
}

func TestInvalidRaceLevel(t *testing.T) {
	handler := NewServer(directory.NewMemoryStore(directory.DirectoryData{}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/races?level=municipal", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got["error"] != "invalid_query" {
		t.Fatalf("error = %q, want %q", got["error"], "invalid_query")
	}
	if got["message"] != "level must be one of federal, statewide, state_legislature, local, unknown" {
		t.Fatalf("message = %q, want documented validation message", got["message"])
	}
}

func TestNotFound(t *testing.T) {
	handler := NewServer(directory.NewMemoryStore(directory.DirectoryData{}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	handler := NewServer(directory.NewMemoryStore(directory.DirectoryData{}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/chapters", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("Allow = %q, want %q", got, http.MethodGet)
	}
}

func TestCivicLookupMethodNotAllowed(t *testing.T) {
	handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/civic/lookup", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != http.MethodPost {
		t.Fatalf("Allow = %q, want %q", got, http.MethodPost)
	}
}

func TestCivicLookupNotConfigured(t *testing.T) {
	handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/civic/lookup", strings.NewReader(`{"address":"123 Main St, San Diego, CA"}`))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["error"] != "civic_not_configured" {
		t.Fatalf("error = %q, want civic_not_configured", got["error"])
	}
}

func TestCivicLookupValidationAndSuccess(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		service        civic.LookupService
		wantStatus     int
		wantError      string
		wantMessages   []string
		wantPollingLen int
	}{
		{
			name:         "blank address",
			body:         `{"address":"   "}`,
			service:      &testCivicService{},
			wantStatus:   http.StatusBadRequest,
			wantError:    "validation_failed",
			wantMessages: []string{"address is required"},
		},
		{
			name:           "success",
			body:           `{"address":"123 Main St, San Diego, CA","electionId":"2000"}`,
			service:        &testCivicService{response: civic.LookupResponse{Status: "success", Source: civic.SourceGoogleCivicInformationAPI, Disclaimer: civic.RegistrationDisclaimer, PollingLocations: []civic.VotingLocation{{Address: civic.Address{Line1: "100 POLLING PL"}}}}},
			wantStatus:     http.StatusOK,
			wantPollingLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testService := tt.service.(*testCivicService)
			testService.called = false
			testService.request = civic.LookupRequest{}

			handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", tt.service)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/civic/lookup", strings.NewReader(tt.body))

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusBadRequest {
				var got map[string]any
				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if got["error"] != tt.wantError {
					t.Fatalf("error = %q, want %q", got["error"], tt.wantError)
				}
				messages := got["messages"].([]any)
				if len(messages) != len(tt.wantMessages) || messages[0] != tt.wantMessages[0] {
					t.Fatalf("messages = %#v, want %#v", messages, tt.wantMessages)
				}
				if testService.called {
					t.Fatal("service should not be called for invalid input")
				}
				return
			}

			if !testService.called {
				t.Fatal("service was not called")
			}
			if testService.request.Address != "123 Main St, San Diego, CA" || testService.request.ElectionID != "2000" {
				t.Fatalf("request = %#v, want trimmed input passed to service", testService.request)
			}

			var got civic.LookupResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Disclaimer != civic.RegistrationDisclaimer {
				t.Fatalf("disclaimer = %q, want civic disclaimer", got.Disclaimer)
			}
			if len(got.PollingLocations) != tt.wantPollingLen {
				t.Fatalf("polling locations len = %d, want %d", len(got.PollingLocations), tt.wantPollingLen)
			}
		})
	}
}

func TestAdminRoutesRequireToken(t *testing.T) {
	handler := NewServerWithAdmin(directory.NewMemoryStore(directory.DirectoryData{}), &testAdminStore{}, "secret")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1"}`))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1"}`))
	req.Header.Set("Authorization", "Bearer wrong")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAdminCreateChapterDraft(t *testing.T) {
	store := &testAdminStore{}
	handler := NewServerWithAdmin(directory.NewMemoryStore(directory.DirectoryData{}), store, "secret")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1","chapterId":"chapter-1","name":"Chapter 1","state":"CA"}`))
	req.Header.Set("Authorization", "Bearer secret")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if store.lastChapterDraft.ID != "draft-1" || store.lastActor != "admin" {
		t.Fatalf("unexpected store call: %#v actor=%q", store.lastChapterDraft, store.lastActor)
	}

	var got admin.ChapterDraft
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "draft-1" || got.ChapterID != "chapter-1" {
		t.Fatalf("unexpected response: %#v", got)
	}
}

func TestAdminCreateChapterDraftValidationFailure(t *testing.T) {
	handler := NewServerWithAdmin(directory.NewMemoryStore(directory.DirectoryData{}), &testAdminStore{}, "secret")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1","chapterId":"chapter-1","name":"Chapter 1","state":"CA","websiteUrl":"ftp://example.com"}`))
	req.Header.Set("Authorization", "Bearer secret")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var got map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["error"] != "validation_failed" {
		t.Fatalf("error = %#v, want validation_failed", got["error"])
	}
}

type testCivicService struct {
	called    bool
	request   civic.LookupRequest
	response  civic.LookupResponse
	lookupErr error
}

func (s *testCivicService) Lookup(ctx context.Context, request civic.LookupRequest) (civic.LookupResponse, error) {
	s.called = true
	s.request = request
	if s.lookupErr != nil {
		return civic.LookupResponse{}, s.lookupErr
	}
	return s.response, nil
}

type testAdminStore struct {
	chapterDrafts     []admin.ChapterDraft
	endorsementDrafts []admin.EndorsementDraft
	chapterDraft      admin.ChapterDraft
	endorsementDraft  admin.EndorsementDraft
	lastActor         string
	lastChapterDraft  admin.ChapterDraft
	lastEndorsement   admin.EndorsementDraft
}

func (s *testAdminStore) ListChapterDrafts(ctx context.Context) ([]admin.ChapterDraft, error) {
	return s.chapterDrafts, nil
}

func (s *testAdminStore) GetChapterDraft(ctx context.Context, id string) (admin.ChapterDraft, error) {
	if s.chapterDraft.ID == id {
		return s.chapterDraft, nil
	}
	return admin.ChapterDraft{}, sql.ErrNoRows
}

func (s *testAdminStore) UpsertChapterDraft(ctx context.Context, draft admin.ChapterDraft, actor string) (admin.ChapterDraft, error) {
	s.lastActor = actor
	s.lastChapterDraft = draft
	if draft.ID == "" {
		draft.ID = "draft-1"
	}
	if draft.Status == "" {
		draft.Status = admin.StatusDraft
	}
	return draft, nil
}

func (s *testAdminStore) PublishChapterDraft(ctx context.Context, id string, actor string) error {
	s.lastActor = actor
	return nil
}

func (s *testAdminStore) ArchiveChapterDraft(ctx context.Context, id string, actor string) error {
	s.lastActor = actor
	return nil
}

func (s *testAdminStore) ListEndorsementDrafts(ctx context.Context) ([]admin.EndorsementDraft, error) {
	return s.endorsementDrafts, nil
}

func (s *testAdminStore) GetEndorsementDraft(ctx context.Context, id string) (admin.EndorsementDraft, error) {
	if s.endorsementDraft.ID == id {
		return s.endorsementDraft, nil
	}
	return admin.EndorsementDraft{}, sql.ErrNoRows
}

func (s *testAdminStore) UpsertEndorsementDraft(ctx context.Context, draft admin.EndorsementDraft, actor string) (admin.EndorsementDraft, error) {
	s.lastActor = actor
	s.lastEndorsement = draft
	if draft.ID == "" {
		draft.ID = "draft-1"
	}
	if draft.Status == "" {
		draft.Status = admin.StatusDraft
	}
	return draft, nil
}

func (s *testAdminStore) PublishEndorsementDraft(ctx context.Context, id string, actor string) error {
	s.lastActor = actor
	return nil
}

func (s *testAdminStore) ArchiveEndorsementDraft(ctx context.Context, id string, actor string) error {
	s.lastActor = actor
	return nil
}
