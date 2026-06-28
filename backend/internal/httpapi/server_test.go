package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
		Chapters: []directory.Chapter{{ID: "chap-1", Name: "Chapter 1", State: "CA", Active: false}},
		VoterGuides: []directory.VoterGuide{{
			ID:         "guide-1",
			Title:      "Guide 1",
			ChapterID:  "chap-1",
			SourceURL:  "https://example.com/guide",
			Confidence: "needs_review",
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

	chapters := got["chapters"].([]any)
	chapter := chapters[0].(map[string]any)
	if active, ok := chapter["active"].(bool); !ok || active {
		t.Fatalf("active = %#v, want false", chapter["active"])
	}

	candidates := got["candidates"].([]any)
	candidate := candidates[0].(map[string]any)
	if incumbent, ok := candidate["incumbent"].(bool); !ok || incumbent {
		t.Fatalf("incumbent = %#v, want false", candidate["incumbent"])
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
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("Allow = %q, want %q", got, http.MethodGet)
	}
}
