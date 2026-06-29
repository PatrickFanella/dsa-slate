package admin

import "context"

type DraftStatus string

const (
	StatusDraft       DraftStatus = "draft"
	StatusNeedsReview DraftStatus = "needs_review"
	StatusPublished   DraftStatus = "published"
	StatusArchived    DraftStatus = "archived"
)

type ChapterDraft struct {
	ID               string      `json:"id"`
	ChapterID        string      `json:"chapterId"`
	Status           DraftStatus `json:"status"`
	Name             string      `json:"name"`
	State            string      `json:"state"`
	WebsiteURL       string      `json:"websiteUrl,omitempty"`
	OfficialDSAURL   string      `json:"officialDsaUrl,omitempty"`
	CoverageNotes    string      `json:"coverageNotes,omitempty"`
	GroupEmail       string      `json:"groupEmail,omitempty"`
	LinktreeURL      string      `json:"linktreeUrl,omitempty"`
	FacebookURL      string      `json:"facebookUrl,omitempty"`
	TwitterURL       string      `json:"twitterUrl,omitempty"`
	BlueskyURL       string      `json:"blueskyUrl,omitempty"`
	InstagramURL     string      `json:"instagramUrl,omitempty"`
	BlogURL          string      `json:"blogUrl,omitempty"`
	EventCalendarURL string      `json:"eventCalendarUrl,omitempty"`
	EmailListURL     string      `json:"emailListUrl,omitempty"`
	ReviewNotes      string      `json:"reviewNotes,omitempty"`
	ReviewedAt       string      `json:"reviewedAt,omitempty"`
}

type EndorsementDraft struct {
	ID              string      `json:"id"`
	EndorsementID   string      `json:"endorsementId"`
	Status          DraftStatus `json:"status"`
	CandidateID     string      `json:"candidateId"`
	RaceID          string      `json:"raceId"`
	ChapterID       string      `json:"chapterId"`
	VoterGuideID    string      `json:"voterGuideId,omitempty"`
	EndorsementType string      `json:"endorsementType"`
	SourceURL       string      `json:"sourceUrl"`
	SourceDate      string      `json:"sourceDate,omitempty"`
	Confidence      string      `json:"confidence"`
	Notes           string      `json:"notes,omitempty"`
	ReviewNotes     string      `json:"reviewNotes,omitempty"`
	ReviewedAt      string      `json:"reviewedAt,omitempty"`
}

type Store interface {
	ListChapterDrafts(ctx context.Context) ([]ChapterDraft, error)
	GetChapterDraft(ctx context.Context, id string) (ChapterDraft, error)
	UpsertChapterDraft(ctx context.Context, draft ChapterDraft, actor string) (ChapterDraft, error)
	PublishChapterDraft(ctx context.Context, id string, actor string) error
	ArchiveChapterDraft(ctx context.Context, id string, actor string) error
	ListEndorsementDrafts(ctx context.Context) ([]EndorsementDraft, error)
	GetEndorsementDraft(ctx context.Context, id string) (EndorsementDraft, error)
	UpsertEndorsementDraft(ctx context.Context, draft EndorsementDraft, actor string) (EndorsementDraft, error)
	PublishEndorsementDraft(ctx context.Context, id string, actor string) error
	ArchiveEndorsementDraft(ctx context.Context, id string, actor string) error
}
