package admin

import (
	"strings"
	"testing"
)

func TestValidateChapterDraftMissingRequiredFields(t *testing.T) {
	errs := ValidateChapterDraft(ChapterDraft{})
	for _, want := range []string{
		"id is required",
		"chapterId is required",
		"name is required",
		"state is required",
	} {
		if !contains(errs, want) {
			t.Fatalf("expected %q in %v", want, errs)
		}
	}
}

func TestValidateChapterDraftRejectsInvalidURLAndStatusAndDate(t *testing.T) {
	errs := ValidateChapterDraft(ChapterDraft{
		ID:         "chapter-1",
		ChapterID:  "chapter-1",
		Name:       "Chapter 1",
		State:      "CA",
		Status:     DraftStatus("oops"),
		WebsiteURL: "ftp://example.com",
		ReviewedAt: "06/29/2026",
	})

	for _, want := range []string{
		"status must be one of draft, needs_review, published, archived",
		"websiteUrl must be an absolute http(s) URL",
		"reviewedAt must be YYYY-MM-DD",
	} {
		if !contains(errs, want) {
			t.Fatalf("expected %q in %v", want, errs)
		}
	}
}

func TestValidateEndorsementDraftRejectsInvalidTypeConfidenceUrlAndDate(t *testing.T) {
	errs := ValidateEndorsementDraft(EndorsementDraft{
		ID:              "endorsement-1",
		EndorsementID:   "endorsement-1",
		CandidateID:     "candidate-1",
		RaceID:          "race-1",
		ChapterID:       "chapter-1",
		EndorsementType: "maybe",
		SourceURL:       "not-a-url",
		SourceDate:      "2026-06-31",
		Confidence:      "unknown-confidence",
	})

	for _, want := range []string{
		"endorsementType must be one of endorsed, recommended, opposed, no_recommendation, write_in, coalition, unknown",
		"sourceUrl must be an absolute http(s) URL",
		"confidence must be one of official_source, official_social_source, campaign_source, press_source, needs_review",
		"sourceDate must be YYYY-MM-DD",
	} {
		if !contains(errs, want) {
			t.Fatalf("expected %q in %v", want, errs)
		}
	}
}

func TestValidateEndorsementDraftMissingSourceURL(t *testing.T) {
	errs := ValidateEndorsementDraft(EndorsementDraft{
		ID:              "endorsement-1",
		EndorsementID:   "endorsement-1",
		CandidateID:     "candidate-1",
		RaceID:          "race-1",
		ChapterID:       "chapter-1",
		EndorsementType: "endorsed",
		Confidence:      "official_source",
	})

	if !contains(errs, "sourceUrl is required") {
		t.Fatalf("expected missing sourceUrl error in %v", errs)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if strings.Contains(value, want) {
			return true
		}
	}
	return false
}
