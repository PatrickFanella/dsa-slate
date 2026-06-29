package admin

import (
	"net/url"
	"strings"
	"time"
)

var allowedDraftStatuses = map[DraftStatus]struct{}{
	StatusDraft:       {},
	StatusNeedsReview: {},
	StatusPublished:   {},
	StatusArchived:    {},
}

var allowedEndorsementTypes = map[string]struct{}{
	"endorsed":          {},
	"recommended":       {},
	"opposed":           {},
	"no_recommendation": {},
	"write_in":          {},
	"coalition":         {},
	"unknown":           {},
}

var allowedConfidenceValues = map[string]struct{}{
	"official_source":        {},
	"official_social_source": {},
	"campaign_source":        {},
	"press_source":           {},
	"needs_review":           {},
}

func ValidateChapterDraft(draft ChapterDraft) []string {
	var messages []string

	require := func(field, value string) {
		if strings.TrimSpace(value) == "" {
			messages = append(messages, field+" is required")
		}
	}

	require("id", draft.ID)
	require("chapterId", draft.ChapterID)
	require("name", draft.Name)
	require("state", draft.State)

	if err := validateDraftStatus(draft.Status); err != "" {
		messages = append(messages, err)
	}

	for _, candidate := range []struct {
		field string
		value string
	}{
		{field: "websiteUrl", value: draft.WebsiteURL},
		{field: "officialDsaUrl", value: draft.OfficialDSAURL},
		{field: "linktreeUrl", value: draft.LinktreeURL},
		{field: "facebookUrl", value: draft.FacebookURL},
		{field: "twitterUrl", value: draft.TwitterURL},
		{field: "blueskyUrl", value: draft.BlueskyURL},
		{field: "instagramUrl", value: draft.InstagramURL},
		{field: "blogUrl", value: draft.BlogURL},
		{field: "eventCalendarUrl", value: draft.EventCalendarURL},
		{field: "emailListUrl", value: draft.EmailListURL},
	} {
		if err := validateOptionalURL(candidate.field, candidate.value); err != "" {
			messages = append(messages, err)
		}
	}

	if err := validateISODate("reviewedAt", draft.ReviewedAt); err != "" {
		messages = append(messages, err)
	}

	return messages
}

func ValidateEndorsementDraft(draft EndorsementDraft) []string {
	var messages []string

	require := func(field, value string) {
		if strings.TrimSpace(value) == "" {
			messages = append(messages, field+" is required")
		}
	}

	require("id", draft.ID)
	require("endorsementId", draft.EndorsementID)
	require("candidateId", draft.CandidateID)
	require("raceId", draft.RaceID)
	require("chapterId", draft.ChapterID)
	require("endorsementType", draft.EndorsementType)
	require("confidence", draft.Confidence)

	if err := validateDraftStatus(draft.Status); err != "" {
		messages = append(messages, err)
	}
	if err := validateEndorsementType(draft.EndorsementType); err != "" {
		messages = append(messages, err)
	}
	if err := validateConfidence(draft.Confidence); err != "" {
		messages = append(messages, err)
	}
	if err := validateRequiredURL("sourceUrl", draft.SourceURL); err != "" {
		messages = append(messages, err)
	}
	if err := validateISODate("sourceDate", draft.SourceDate); err != "" {
		messages = append(messages, err)
	}
	if err := validateISODate("reviewedAt", draft.ReviewedAt); err != "" {
		messages = append(messages, err)
	}

	return messages
}

func validateDraftStatus(status DraftStatus) string {
	value := strings.TrimSpace(string(status))
	if value == "" {
		return ""
	}
	if _, ok := allowedDraftStatuses[DraftStatus(value)]; !ok {
		return "status must be one of draft, needs_review, published, archived"
	}
	return ""
}

func validateEndorsementType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if _, ok := allowedEndorsementTypes[value]; !ok {
		return "endorsementType must be one of endorsed, recommended, opposed, no_recommendation, write_in, coalition, unknown"
	}
	return ""
}

func validateConfidence(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if _, ok := allowedConfidenceValues[value]; !ok {
		return "confidence must be one of official_source, official_social_source, campaign_source, press_source, needs_review"
	}
	return ""
}

func validateRequiredURL(field, value string) string {
	if strings.TrimSpace(value) == "" {
		return field + " is required"
	}
	return validateOptionalURL(field, value)
}

func validateOptionalURL(field, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return field + " must be an absolute http(s) URL"
	}
	return ""
}

func validateISODate(field, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return field + " must be YYYY-MM-DD"
	}
	return ""
}
