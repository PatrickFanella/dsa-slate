package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/dsa-slate/internal/admin"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/auth"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/civic"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/directory"
	"git.subcult.tv/subculture-collective/dsa-slate/internal/savedslate"
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

func addCSRF(req *http.Request) {
	req.Header.Set("X-DSA-Slate-CSRF", "1")
	req.Header.Set("Origin", "http://localhost:5173")
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

func TestCivicElectionsMethodNotAllowed(t *testing.T) {
	handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", &testCivicService{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/civic/elections", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("Allow = %q, want GET", got)
	}
}

func TestCivicElectionsNotConfigured(t *testing.T) {
	handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/civic/elections", nil)

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

func TestCivicElectionsSuccessAndUpstreamError(t *testing.T) {
	tests := []struct {
		name       string
		service    civic.LookupService
		wantStatus int
		wantError  string
	}{
		{
			name:       "success",
			service:    &testCivicService{electionsResponse: civic.ElectionsResponse{Source: civic.SourceGoogleCivicInformationAPI, Elections: []civic.Election{{ID: "2000", Name: "VIP Test Election", ElectionDay: "2031-12-06"}}}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "upstream error",
			service:    &testCivicService{electionsErr: civic.UpstreamError{StatusCode: http.StatusBadRequest, Message: "API key not authorized"}},
			wantStatus: http.StatusBadRequest,
			wantError:  "civic_lookup_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", tt.service)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/civic/elections", nil)

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus != http.StatusOK {
				var got map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if got["error"] != tt.wantError {
					t.Fatalf("error = %q, want %q", got["error"], tt.wantError)
				}
				return
			}

			var got civic.ElectionsResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(got.Elections) != 1 || got.Elections[0].ID != "2000" {
				t.Fatalf("response = %#v, want one election", got)
			}
		})
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

func TestCivicLookupSurfacesUpstreamMessage(t *testing.T) {
	handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", &testCivicService{lookupErr: civic.UpstreamError{StatusCode: http.StatusBadRequest, Message: "Election unknown"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/civic/lookup", strings.NewReader(`{"address":"123 Main St, San Diego, CA"}`))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["error"] != "civic_lookup_failed" || got["message"] != "Election unknown" {
		t.Fatalf("response = %#v, want upstream civic error", got)
	}
}

func TestAuthSessionAnonymous(t *testing.T) {
	handler := NewServerWithAuth(directory.NewMemoryStore(directory.DirectoryData{}), &testAuthStore{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got auth.State
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Authenticated {
		t.Fatalf("authenticated = true, want false")
	}
}

func TestAuthSessionResolved(t *testing.T) {
	now := time.Now().UTC()
	store := &testAuthStore{
		sessions: map[string]auth.Session{
			"session-1": {ID: "session-1", UserID: "user-1", ExpiresAt: now.Add(time.Hour)},
		},
		users: map[string]auth.User{
			"user-1": {ID: "user-1", Email: "editor@example.org", Role: auth.RoleEditor},
		},
	}
	handler := NewServerWithAuth(directory.NewMemoryStore(directory.DirectoryData{}), store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got auth.State
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.Authenticated || got.User == nil {
		t.Fatalf("state = %#v, want authenticated user", got)
	}
	if got.User.ID != "user-1" || got.User.Email != "editor@example.org" || got.User.Role != auth.RoleEditor {
		t.Fatalf("user = %#v, want user-1/editor@example.org/editor", got.User)
	}
	if store.sessions["session-1"].ExpiresAt.Before(now) {
		t.Fatal("test setup invalid")
	}
}

func TestAuthGoogleStartNotConfigured(t *testing.T) {
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), &testAuthStore{}, nil, nil, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/start", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["error"] != "auth_not_configured" {
		t.Fatalf("error = %q, want auth_not_configured", got["error"])
	}
}

func TestAuthGoogleStartDelegatesToService(t *testing.T) {
	service := &testGoogleSignInService{startStatus: http.StatusFound, startLocation: "/api/auth/google/redirect"}
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), &testAuthStore{}, service, nil, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/start", nil)

	handler.ServeHTTP(rec, req)

	if !service.startCalled {
		t.Fatal("service start was not called")
	}
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "/api/auth/google/redirect" {
		t.Fatalf("Location = %q, want redirect", got)
	}
}

func TestAuthLogoutClearsCookieAndDeletesSession(t *testing.T) {
	store := &testAuthStore{sessions: map[string]auth.Session{"session-1": {ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().UTC().Add(time.Hour)}}}
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), store, nil, nil, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	addCSRF(req)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if _, ok := store.sessions["session-1"]; ok {
		t.Fatal("session was not deleted")
	}
	if got := rec.Header().Values("Set-Cookie"); len(got) == 0 || !strings.Contains(strings.Join(got, "\n"), auth.SessionCookieName+"=") {
		t.Fatalf("expected cleared session cookie, got %v", got)
	}
	var body auth.State
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Authenticated {
		t.Fatalf("authenticated = true, want false")
	}
}

func TestAuthLogoutRejectsMissingCSRF(t *testing.T) {
	store := &testAuthStore{sessions: map[string]auth.Session{"session-1": {ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().UTC().Add(time.Hour)}}}
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), store, nil, nil, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if _, ok := store.sessions["session-1"]; !ok {
		t.Fatal("session was deleted despite CSRF failure")
	}
}

func TestSavedSlateRejectsAnonymousSession(t *testing.T) {
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), &testAuthStore{}, nil, nil, &testSavedSlateStore{}, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/me/slate", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestSavedSlateListAndUpsertForAuthenticatedUser(t *testing.T) {
	store := &testSavedSlateStore{}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "user-1", Email: "reader@example.org", Role: auth.RoleUser})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, nil, store, "", nil)

	putRec := httptest.NewRecorder()
	putReq := httptest.NewRequest(http.MethodPut, "/api/me/slate/items", strings.NewReader(`{"recordType":"endorsement","recordId":"endorsement-1"}`))
	putReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	addCSRF(putReq)
	handler.ServeHTTP(putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Fatalf("upsert status = %d, want %d", putRec.Code, http.StatusOK)
	}
	if store.lastUserID != "user-1" || store.lastRecordType != savedslate.RecordTypeEndorsement || store.lastRecordID != "endorsement-1" {
		t.Fatalf("unexpected store call user=%q type=%q id=%q", store.lastUserID, store.lastRecordType, store.lastRecordID)
	}
	var item savedslate.Item
	if err := json.NewDecoder(putRec.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}
	if item.ID == "" || item.UserID != "user-1" || item.RecordType != savedslate.RecordTypeEndorsement || item.RecordID != "endorsement-1" {
		t.Fatalf("unexpected item: %#v", item)
	}

	getRec := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/api/me/slate", nil)
	getReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", getRec.Code, http.StatusOK)
	}
	var slate savedslate.Slate
	if err := json.NewDecoder(getRec.Body).Decode(&slate); err != nil {
		t.Fatalf("decode slate: %v", err)
	}
	if len(slate.Items) != 1 || slate.Items[0].RecordID != "endorsement-1" {
		t.Fatalf("unexpected slate: %#v", slate)
	}
}

func TestSavedSlateCanSaveMeasurePositionsAndVoterGuides(t *testing.T) {
	store := &testSavedSlateStore{}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "user-1", Email: "reader@example.org", Role: auth.RoleUser})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, nil, store, "", nil)

	for _, body := range []string{
		`{"recordType":"measure_position","recordId":"measure-position-1"}`,
		`{"recordType":"voter_guide","recordId":"voter-guide-1"}`,
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/me/slate/items", strings.NewReader(body))
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
		addCSRF(req)

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d for %s", rec.Code, http.StatusOK, body)
		}
	}
}

func TestSavedSlateWriteRejectsMissingCSRF(t *testing.T) {
	store := &testSavedSlateStore{}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "user-1", Email: "reader@example.org", Role: auth.RoleUser})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, nil, store, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/me/slate/items", strings.NewReader(`{"recordType":"endorsement","recordId":"endorsement-1"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if store.lastUserID != "" {
		t.Fatalf("store was called despite CSRF failure: %#v", store)
	}
}

func TestSavedSlateDeleteRequiresOwner(t *testing.T) {
	store := &testSavedSlateStore{
		items: map[string][]savedslate.Item{
			"owner-1": {{ID: "item-1", UserID: "owner-1", RecordType: savedslate.RecordTypeEndorsement, RecordID: "endorsement-1"}},
		},
	}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "user-2", Email: "other@example.org", Role: auth.RoleUser})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, nil, store, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/me/slate/items/item-1", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	addCSRF(req)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestSavedSlateRejectsInvalidReference(t *testing.T) {
	store := &testSavedSlateStore{invalidRecords: map[string]bool{"endorsement\x00missing": true}}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "user-1", Email: "reader@example.org", Role: auth.RoleUser})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, nil, store, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/me/slate/items", strings.NewReader(`{"recordType":"endorsement","recordId":"missing"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	addCSRF(req)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSavedSlateRejectsInvalidRecordType(t *testing.T) {
	store := &testSavedSlateStore{}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "user-1", Email: "reader@example.org", Role: auth.RoleUser})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, nil, store, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/me/slate/items", strings.NewReader(`{"recordType":"candidate","recordId":"candidate-1"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	addCSRF(req)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAdminRoutesRejectAnonymousSession(t *testing.T) {
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), &testAuthStore{}, nil, &testAdminStore{}, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1"}`))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAdminRoutesRejectUserRole(t *testing.T) {
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "user-1", Email: "reader@example.org", Role: auth.RoleUser})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, &testAdminStore{}, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminCreateChapterDraftAllowsEditorAndRecordsActor(t *testing.T) {
	store := &testAdminStore{}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "editor-1", Email: "editor@example.org", Role: auth.RoleEditor})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, store, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1","chapterId":"chapter-1","name":"Chapter 1","state":"CA"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	addCSRF(req)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if store.lastChapterDraft.ID != "draft-1" || store.lastActor != "editor@example.org" {
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

func TestAdminWriteRejectsBadCSRFOrigin(t *testing.T) {
	store := &testAdminStore{}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "editor-1", Email: "editor@example.org", Role: auth.RoleEditor})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, store, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1","chapterId":"chapter-1","name":"Chapter 1","state":"CA"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	req.Header.Set("X-DSA-Slate-CSRF", "1")
	req.Header.Set("Origin", "https://evil.example")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if store.lastActor != "" {
		t.Fatalf("store was called despite CSRF failure: %#v", store)
	}
}

func TestAdminWriteRejectsCrossSiteFetchMetadata(t *testing.T) {
	store := &testAdminStore{}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "editor-1", Email: "editor@example.org", Role: auth.RoleEditor})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, store, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1","chapterId":"chapter-1","name":"Chapter 1","state":"CA"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	addCSRF(req)
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if store.lastActor != "" {
		t.Fatalf("store was called despite CSRF failure: %#v", store)
	}
}

func TestAdminCreateChapterDraftAllowsAdminAndRecordsIDActor(t *testing.T) {
	store := &testAdminStore{}
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "admin-1", Role: auth.RoleAdmin})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, store, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1","chapterId":"chapter-1","name":"Chapter 1","state":"CA"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	addCSRF(req)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if store.lastActor != "admin-1" {
		t.Fatalf("actor = %q, want admin-1", store.lastActor)
	}
}

func TestAdminCreateChapterDraftValidationFailure(t *testing.T) {
	authStore := newTestAuthStoreWithSession("session-1", auth.User{ID: "editor-1", Email: "editor@example.org", Role: auth.RoleEditor})
	handler := NewServerWithAuthAndServices(directory.NewMemoryStore(directory.DirectoryData{}), authStore, nil, &testAdminStore{}, nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/chapter-drafts", strings.NewReader(`{"id":"draft-1","chapterId":"chapter-1","name":"Chapter 1","state":"CA","websiteUrl":"ftp://example.com"}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "session-1"})
	addCSRF(req)

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
	called            bool
	request           civic.LookupRequest
	response          civic.LookupResponse
	lookupErr         error
	electionsResponse civic.ElectionsResponse
	electionsErr      error
}

func (s *testCivicService) Lookup(ctx context.Context, request civic.LookupRequest) (civic.LookupResponse, error) {
	s.called = true
	s.request = request
	if s.lookupErr != nil {
		return civic.LookupResponse{}, s.lookupErr
	}
	return s.response, nil
}

func (s *testCivicService) Elections(ctx context.Context) (civic.ElectionsResponse, error) {
	if s.electionsErr != nil {
		return civic.ElectionsResponse{}, s.electionsErr
	}
	return s.electionsResponse, nil
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

type testSavedSlateStore struct {
	items          map[string][]savedslate.Item
	invalidRecords map[string]bool
	lastUserID     string
	lastRecordType string
	lastRecordID   string
}

type testGoogleSignInService struct {
	startCalled      bool
	callbackCalled   bool
	startStatus      int
	startLocation    string
	callbackStatus   int
	callbackLocation string
}

func (s *testGoogleSignInService) Start(w http.ResponseWriter, r *http.Request) {
	s.startCalled = true
	if s.startStatus == 0 {
		s.startStatus = http.StatusFound
	}
	if s.startLocation != "" {
		w.Header().Set("Location", s.startLocation)
	}
	w.WriteHeader(s.startStatus)
}

func (s *testGoogleSignInService) Callback(w http.ResponseWriter, r *http.Request) {
	s.callbackCalled = true
	if s.callbackStatus == 0 {
		s.callbackStatus = http.StatusFound
	}
	if s.callbackLocation != "" {
		w.Header().Set("Location", s.callbackLocation)
	}
	w.WriteHeader(s.callbackStatus)
}

type testAuthStore struct {
	users      map[string]auth.User
	sessions   map[string]auth.Session
	identities map[string]auth.OAuthIdentity
}

func newTestAuthStoreWithSession(token string, user auth.User) *testAuthStore {
	now := time.Now().UTC()
	if user.ID == "" {
		user.ID = "user-1"
	}
	return &testAuthStore{
		users: map[string]auth.User{
			user.ID: user,
		},
		sessions: map[string]auth.Session{
			token: {ID: token, UserID: user.ID, ExpiresAt: now.Add(time.Hour), CreatedAt: now},
		},
	}
}

func (s *testAuthStore) UpsertUser(ctx context.Context, user auth.User) (auth.User, error) {
	if s.users == nil {
		s.users = make(map[string]auth.User)
	}
	s.users[user.ID] = user
	return user, nil
}

func (s *testAuthStore) GetUser(ctx context.Context, id string) (auth.User, error) {
	if user, ok := s.users[id]; ok {
		return user, nil
	}
	return auth.User{}, sql.ErrNoRows
}

func (s *testAuthStore) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return auth.User{}, sql.ErrNoRows
}

func (s *testAuthStore) UpsertOAuthIdentity(ctx context.Context, identity auth.OAuthIdentity) (auth.OAuthIdentity, error) {
	if s.identities == nil {
		s.identities = make(map[string]auth.OAuthIdentity)
	}
	key := identity.Provider + "\x00" + identity.ProviderSubject
	s.identities[key] = identity
	return identity, nil
}

func (s *testAuthStore) GetOAuthIdentity(ctx context.Context, provider string, subject string) (auth.OAuthIdentity, error) {
	if s.identities == nil {
		return auth.OAuthIdentity{}, sql.ErrNoRows
	}
	if identity, ok := s.identities[provider+"\x00"+subject]; ok {
		return identity, nil
	}
	return auth.OAuthIdentity{}, sql.ErrNoRows
}

func (s *testAuthStore) CreateSession(ctx context.Context, session auth.Session) (auth.Session, error) {
	if s.sessions == nil {
		s.sessions = make(map[string]auth.Session)
	}
	s.sessions[session.ID] = session
	return session, nil
}

func (s *testAuthStore) GetSession(ctx context.Context, token string) (auth.Session, error) {
	if session, ok := s.sessions[token]; ok {
		return session, nil
	}
	return auth.Session{}, sql.ErrNoRows
}

func (s *testAuthStore) DeleteSession(ctx context.Context, token string) error {
	delete(s.sessions, token)
	return nil
}

func (s *testSavedSlateStore) ListForUser(ctx context.Context, userID string) ([]savedslate.Item, error) {
	if s.items == nil {
		return []savedslate.Item{}, nil
	}
	items := s.items[userID]
	if items == nil {
		return []savedslate.Item{}, nil
	}
	return items, nil
}

func (s *testSavedSlateStore) UpsertForUser(ctx context.Context, userID string, recordType string, recordID string) (savedslate.Item, error) {
	s.lastUserID = userID
	s.lastRecordType = recordType
	s.lastRecordID = recordID
	if s.invalidRecords[recordType+"\x00"+recordID] {
		return savedslate.Item{}, savedslate.ErrInvalidReference
	}
	if s.items == nil {
		s.items = make(map[string][]savedslate.Item)
	}
	item := savedslate.Item{ID: "item-" + recordType + "-" + recordID, UserID: userID, RecordType: recordType, RecordID: recordID, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	for i, existing := range s.items[userID] {
		if existing.RecordType == recordType && existing.RecordID == recordID {
			s.items[userID][i] = item
			return item, nil
		}
	}
	s.items[userID] = append(s.items[userID], item)
	return item, nil
}

func (s *testSavedSlateStore) DeleteForUser(ctx context.Context, userID string, itemID string) error {
	items := s.items[userID]
	for i, item := range items {
		if item.ID == itemID {
			s.items[userID] = append(items[:i], items[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
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
