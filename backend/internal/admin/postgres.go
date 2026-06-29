package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ValidationError struct {
	Messages []string
}

func (e ValidationError) Error() string {
	return strings.Join(e.Messages, "; ")
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

type rowScanner interface {
	Scan(...any) error
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) ListChapterDrafts(ctx context.Context) ([]ChapterDraft, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, status, name, state, website_url, official_dsa_url, coverage_notes, group_email, linktree_url, facebook_url, twitter_url, bluesky_url, instagram_url, blog_url, event_calendar_url, email_list_url, review_notes, reviewed_at
		FROM admin_chapter_drafts
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query chapter drafts: %w", err)
	}
	defer rows.Close()

	drafts := make([]ChapterDraft, 0)
	for rows.Next() {
		draft, err := scanChapterDraftRow(rows)
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, draft)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chapter drafts: %w", err)
	}

	return drafts, nil
}

func (s *PostgresStore) GetChapterDraft(ctx context.Context, id string) (ChapterDraft, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, status, name, state, website_url, official_dsa_url, coverage_notes, group_email, linktree_url, facebook_url, twitter_url, bluesky_url, instagram_url, blog_url, event_calendar_url, email_list_url, review_notes, reviewed_at
		FROM admin_chapter_drafts
		WHERE id = $1
	`, strings.TrimSpace(id))
	return scanChapterDraftRow(row)
}

func (s *PostgresStore) UpsertChapterDraft(ctx context.Context, draft ChapterDraft, actor string) (ChapterDraft, error) {
	if messages := ValidateChapterDraft(draft); len(messages) > 0 {
		return ChapterDraft{}, ValidationError{Messages: messages}
	}

	var saved ChapterDraft
	if err := s.withTx(ctx, func(db queryer) error {
		inserted, err := exists(ctx, db, "admin_chapter_drafts", draft.ID)
		if err != nil {
			return err
		}

		row := db.QueryRow(ctx, `
			INSERT INTO admin_chapter_drafts (
				id, chapter_id, status, name, state, website_url, official_dsa_url, coverage_notes, group_email, linktree_url, facebook_url, twitter_url, bluesky_url, instagram_url, blog_url, event_calendar_url, email_list_url, review_notes, reviewed_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
			)
			ON CONFLICT (id) DO UPDATE SET
				chapter_id = EXCLUDED.chapter_id,
				status = EXCLUDED.status,
				name = EXCLUDED.name,
				state = EXCLUDED.state,
				website_url = EXCLUDED.website_url,
				official_dsa_url = EXCLUDED.official_dsa_url,
				coverage_notes = EXCLUDED.coverage_notes,
				group_email = EXCLUDED.group_email,
				linktree_url = EXCLUDED.linktree_url,
				facebook_url = EXCLUDED.facebook_url,
				twitter_url = EXCLUDED.twitter_url,
				bluesky_url = EXCLUDED.bluesky_url,
				instagram_url = EXCLUDED.instagram_url,
				blog_url = EXCLUDED.blog_url,
				event_calendar_url = EXCLUDED.event_calendar_url,
				email_list_url = EXCLUDED.email_list_url,
				review_notes = EXCLUDED.review_notes,
				reviewed_at = EXCLUDED.reviewed_at,
				updated_at = now()
			RETURNING id, chapter_id, status, name, state, website_url, official_dsa_url, coverage_notes, group_email, linktree_url, facebook_url, twitter_url, bluesky_url, instagram_url, blog_url, event_calendar_url, email_list_url, review_notes, reviewed_at
		`, chapterDraftArgs(draft)...)

		savedDraft, err := scanChapterDraftRow(row)
		if err != nil {
			return err
		}
		saved = savedDraft

		action := "update"
		if !inserted {
			action = "create"
		}
		message := fmt.Sprintf("chapter draft %s", action)
		return insertAudit(ctx, db, actor, action, "chapter_draft", saved.ID, message)
	}); err != nil {
		return ChapterDraft{}, err
	}

	return saved, nil
}

func (s *PostgresStore) PublishChapterDraft(ctx context.Context, id string, actor string) error {
	return s.withTx(ctx, func(db queryer) error {
		draft, err := getChapterDraft(ctx, db, id)
		if err != nil {
			return err
		}

		if _, err := db.Exec(ctx, `
			INSERT INTO chapters (
				id, name, state, website_url, official_dsa_url, coverage_notes, group_email, linktree_url, facebook_url, twitter_url, bluesky_url, instagram_url, blog_url, event_calendar_url, email_list_url, active
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, true
			)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				state = EXCLUDED.state,
				website_url = EXCLUDED.website_url,
				official_dsa_url = EXCLUDED.official_dsa_url,
				coverage_notes = EXCLUDED.coverage_notes,
				group_email = EXCLUDED.group_email,
				linktree_url = EXCLUDED.linktree_url,
				facebook_url = EXCLUDED.facebook_url,
				twitter_url = EXCLUDED.twitter_url,
				bluesky_url = EXCLUDED.bluesky_url,
				instagram_url = EXCLUDED.instagram_url,
				blog_url = EXCLUDED.blog_url,
				event_calendar_url = EXCLUDED.event_calendar_url,
				email_list_url = EXCLUDED.email_list_url,
				active = true
		`, chapterPublishArgs(draft)...); err != nil {
			return fmt.Errorf("publish chapter draft: %w", err)
		}

		if _, err := db.Exec(ctx, `
			UPDATE admin_chapter_drafts
			SET status = 'published', updated_at = now()
			WHERE id = $1
		`, draft.ID); err != nil {
			return fmt.Errorf("mark chapter draft published: %w", err)
		}

		return insertAudit(ctx, db, actor, "publish", "chapter_draft", draft.ID, fmt.Sprintf("published chapter %s", draft.ChapterID))
	})
}

func (s *PostgresStore) ArchiveChapterDraft(ctx context.Context, id string, actor string) error {
	return s.withTx(ctx, func(db queryer) error {
		draft, err := getChapterDraft(ctx, db, id)
		if err != nil {
			return err
		}

		if _, err := db.Exec(ctx, `
			UPDATE admin_chapter_drafts
			SET status = 'archived', updated_at = now()
			WHERE id = $1
		`, draft.ID); err != nil {
			return fmt.Errorf("archive chapter draft: %w", err)
		}

		return insertAudit(ctx, db, actor, "archive", "chapter_draft", draft.ID, fmt.Sprintf("archived chapter %s", draft.ChapterID))
	})
}

func (s *PostgresStore) ListEndorsementDrafts(ctx context.Context) ([]EndorsementDraft, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, endorsement_id, status, candidate_id, race_id, chapter_id, voter_guide_id, endorsement_type, source_url, source_date, confidence, notes, review_notes, reviewed_at
		FROM admin_endorsement_drafts
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query endorsement drafts: %w", err)
	}
	defer rows.Close()

	drafts := make([]EndorsementDraft, 0)
	for rows.Next() {
		draft, err := scanEndorsementDraftRow(rows)
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, draft)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate endorsement drafts: %w", err)
	}

	return drafts, nil
}

func (s *PostgresStore) GetEndorsementDraft(ctx context.Context, id string) (EndorsementDraft, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, endorsement_id, status, candidate_id, race_id, chapter_id, voter_guide_id, endorsement_type, source_url, source_date, confidence, notes, review_notes, reviewed_at
		FROM admin_endorsement_drafts
		WHERE id = $1
	`, strings.TrimSpace(id))
	return scanEndorsementDraftRow(row)
}

func (s *PostgresStore) UpsertEndorsementDraft(ctx context.Context, draft EndorsementDraft, actor string) (EndorsementDraft, error) {
	if messages := ValidateEndorsementDraft(draft); len(messages) > 0 {
		return EndorsementDraft{}, ValidationError{Messages: messages}
	}

	var saved EndorsementDraft
	if err := s.withTx(ctx, func(db queryer) error {
		existing, err := exists(ctx, db, "admin_endorsement_drafts", draft.ID)
		if err != nil {
			return err
		}

		row := db.QueryRow(ctx, `
			INSERT INTO admin_endorsement_drafts (
				id, endorsement_id, status, candidate_id, race_id, chapter_id, voter_guide_id, endorsement_type, source_url, source_date, confidence, notes, review_notes, reviewed_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
			)
			ON CONFLICT (id) DO UPDATE SET
				endorsement_id = EXCLUDED.endorsement_id,
				status = EXCLUDED.status,
				candidate_id = EXCLUDED.candidate_id,
				race_id = EXCLUDED.race_id,
				chapter_id = EXCLUDED.chapter_id,
				voter_guide_id = EXCLUDED.voter_guide_id,
				endorsement_type = EXCLUDED.endorsement_type,
				source_url = EXCLUDED.source_url,
				source_date = EXCLUDED.source_date,
				confidence = EXCLUDED.confidence,
				notes = EXCLUDED.notes,
				review_notes = EXCLUDED.review_notes,
				reviewed_at = EXCLUDED.reviewed_at,
				updated_at = now()
			RETURNING id, endorsement_id, status, candidate_id, race_id, chapter_id, voter_guide_id, endorsement_type, source_url, source_date, confidence, notes, review_notes, reviewed_at
		`, endorsementDraftArgs(draft)...)

		savedDraft, err := scanEndorsementDraftRow(row)
		if err != nil {
			return err
		}
		saved = savedDraft

		action := "update"
		if !existing {
			action = "create"
		}
		return insertAudit(ctx, db, actor, action, "endorsement_draft", saved.ID, fmt.Sprintf("endorsement draft %s", action))
	}); err != nil {
		return EndorsementDraft{}, err
	}

	return saved, nil
}

func (s *PostgresStore) PublishEndorsementDraft(ctx context.Context, id string, actor string) error {
	return s.withTx(ctx, func(db queryer) error {
		draft, err := getEndorsementDraft(ctx, db, id)
		if err != nil {
			return err
		}

		if err := ensureEndorsementReferences(ctx, db, draft); err != nil {
			return err
		}

		if _, err := db.Exec(ctx, `
			INSERT INTO endorsements (
				id, candidate_id, race_id, chapter_id, voter_guide_id, endorsement_type, source_url, source_date, confidence, notes, reviewed_at, review_notes
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
			)
			ON CONFLICT (id) DO UPDATE SET
				candidate_id = EXCLUDED.candidate_id,
				race_id = EXCLUDED.race_id,
				chapter_id = EXCLUDED.chapter_id,
				voter_guide_id = EXCLUDED.voter_guide_id,
				endorsement_type = EXCLUDED.endorsement_type,
				source_url = EXCLUDED.source_url,
				source_date = EXCLUDED.source_date,
				confidence = EXCLUDED.confidence,
				notes = EXCLUDED.notes,
				reviewed_at = EXCLUDED.reviewed_at,
				review_notes = EXCLUDED.review_notes
		`, endorsementPublishArgs(draft)...); err != nil {
			return fmt.Errorf("publish endorsement draft: %w", err)
		}

		if _, err := db.Exec(ctx, `
			UPDATE admin_endorsement_drafts
			SET status = 'published', updated_at = now()
			WHERE id = $1
		`, draft.ID); err != nil {
			return fmt.Errorf("mark endorsement draft published: %w", err)
		}

		return insertAudit(ctx, db, actor, "publish", "endorsement_draft", draft.ID, fmt.Sprintf("published endorsement %s", draft.EndorsementID))
	})
}

func (s *PostgresStore) ArchiveEndorsementDraft(ctx context.Context, id string, actor string) error {
	return s.withTx(ctx, func(db queryer) error {
		draft, err := getEndorsementDraft(ctx, db, id)
		if err != nil {
			return err
		}

		if _, err := db.Exec(ctx, `
			UPDATE admin_endorsement_drafts
			SET status = 'archived', updated_at = now()
			WHERE id = $1
		`, draft.ID); err != nil {
			return fmt.Errorf("archive endorsement draft: %w", err)
		}

		return insertAudit(ctx, db, actor, "archive", "endorsement_draft", draft.ID, fmt.Sprintf("archived endorsement %s", draft.EndorsementID))
	})
}

func (s *PostgresStore) withTx(ctx context.Context, fn func(queryer) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func exists(ctx context.Context, db queryer, table string, id string) (bool, error) {
	var found int
	err := db.QueryRow(ctx, fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1", table), strings.TrimSpace(id)).Scan(&found)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func getChapterDraft(ctx context.Context, db queryer, id string) (ChapterDraft, error) {
	return scanChapterDraftRow(db.QueryRow(ctx, `
		SELECT id, chapter_id, status, name, state, website_url, official_dsa_url, coverage_notes, group_email, linktree_url, facebook_url, twitter_url, bluesky_url, instagram_url, blog_url, event_calendar_url, email_list_url, review_notes, reviewed_at
		FROM admin_chapter_drafts
		WHERE id = $1
	`, strings.TrimSpace(id)))
}

func getEndorsementDraft(ctx context.Context, db queryer, id string) (EndorsementDraft, error) {
	return scanEndorsementDraftRow(db.QueryRow(ctx, `
		SELECT id, endorsement_id, status, candidate_id, race_id, chapter_id, voter_guide_id, endorsement_type, source_url, source_date, confidence, notes, review_notes, reviewed_at
		FROM admin_endorsement_drafts
		WHERE id = $1
	`, strings.TrimSpace(id)))
}

func insertAudit(ctx context.Context, db queryer, actor string, action string, recordType string, recordID string, message string) error {
	if strings.TrimSpace(actor) == "" {
		actor = "admin"
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO admin_audit_events (actor, action, record_type, record_id, message)
		VALUES ($1, $2, $3, $4, $5)
	`, actor, action, recordType, recordID, message); err != nil {
		return fmt.Errorf("insert admin audit event: %w", err)
	}
	return nil
}

func ensureEndorsementReferences(ctx context.Context, db queryer, draft EndorsementDraft) error {
	messages := make([]string, 0, 3)
	check := func(table, field, id string) error {
		if strings.TrimSpace(id) == "" {
			return nil
		}
		var found int
		err := db.QueryRow(ctx, fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1", table), strings.TrimSpace(id)).Scan(&found)
		if err == nil {
			return nil
		}
		if errors.Is(err, sql.ErrNoRows) {
			messages = append(messages, fmt.Sprintf("%s references missing %s[%s]", field, table, strings.TrimSpace(id)))
			return nil
		}
		return err
	}

	if err := check("candidates", "candidateId", draft.CandidateID); err != nil {
		return err
	}
	if err := check("races", "raceId", draft.RaceID); err != nil {
		return err
	}
	if err := check("chapters", "chapterId", draft.ChapterID); err != nil {
		return err
	}
	if len(messages) > 0 {
		return ValidationError{Messages: messages}
	}
	return nil
}

func chapterDraftArgs(draft ChapterDraft) []any {
	reviewedAt, _ := parseOptionalDate(draft.ReviewedAt)
	return []any{
		strings.TrimSpace(draft.ID),
		strings.TrimSpace(draft.ChapterID),
		normalizeDraftStatus(draft.Status),
		strings.TrimSpace(draft.Name),
		strings.TrimSpace(draft.State),
		nullableString(draft.WebsiteURL),
		nullableString(draft.OfficialDSAURL),
		nullableString(draft.CoverageNotes),
		nullableString(draft.GroupEmail),
		nullableString(draft.LinktreeURL),
		nullableString(draft.FacebookURL),
		nullableString(draft.TwitterURL),
		nullableString(draft.BlueskyURL),
		nullableString(draft.InstagramURL),
		nullableString(draft.BlogURL),
		nullableString(draft.EventCalendarURL),
		nullableString(draft.EmailListURL),
		nullableString(draft.ReviewNotes),
		reviewedAt,
	}
}

func chapterPublishArgs(draft ChapterDraft) []any {
	return []any{
		strings.TrimSpace(draft.ChapterID),
		strings.TrimSpace(draft.Name),
		strings.TrimSpace(draft.State),
		nullableString(draft.WebsiteURL),
		nullableString(draft.OfficialDSAURL),
		nullableString(draft.CoverageNotes),
		nullableString(draft.GroupEmail),
		nullableString(draft.LinktreeURL),
		nullableString(draft.FacebookURL),
		nullableString(draft.TwitterURL),
		nullableString(draft.BlueskyURL),
		nullableString(draft.InstagramURL),
		nullableString(draft.BlogURL),
		nullableString(draft.EventCalendarURL),
		nullableString(draft.EmailListURL),
	}
}

func endorsementDraftArgs(draft EndorsementDraft) []any {
	sourceDate, _ := parseOptionalDate(draft.SourceDate)
	reviewedAt, _ := parseOptionalDate(draft.ReviewedAt)
	return []any{
		strings.TrimSpace(draft.ID),
		strings.TrimSpace(draft.EndorsementID),
		normalizeDraftStatus(draft.Status),
		strings.TrimSpace(draft.CandidateID),
		strings.TrimSpace(draft.RaceID),
		strings.TrimSpace(draft.ChapterID),
		nullableString(draft.VoterGuideID),
		strings.TrimSpace(draft.EndorsementType),
		strings.TrimSpace(draft.SourceURL),
		sourceDate,
		strings.TrimSpace(draft.Confidence),
		nullableString(draft.Notes),
		nullableString(draft.ReviewNotes),
		reviewedAt,
	}
}

func endorsementPublishArgs(draft EndorsementDraft) []any {
	sourceDate, _ := parseOptionalDate(draft.SourceDate)
	reviewedAt, _ := parseOptionalDate(draft.ReviewedAt)
	return []any{
		strings.TrimSpace(draft.EndorsementID),
		strings.TrimSpace(draft.CandidateID),
		strings.TrimSpace(draft.RaceID),
		strings.TrimSpace(draft.ChapterID),
		nullableString(draft.VoterGuideID),
		strings.TrimSpace(draft.EndorsementType),
		strings.TrimSpace(draft.SourceURL),
		sourceDate,
		strings.TrimSpace(draft.Confidence),
		nullableString(draft.Notes),
		reviewedAt,
		nullableString(draft.ReviewNotes),
	}
}

func scanChapterDraftRow(scanner rowScanner) (ChapterDraft, error) {
	var (
		id               sql.NullString
		chapterID        sql.NullString
		status           sql.NullString
		name             sql.NullString
		state            sql.NullString
		websiteURL       sql.NullString
		officialDSAURL   sql.NullString
		coverageNotes    sql.NullString
		groupEmail       sql.NullString
		linktreeURL      sql.NullString
		facebookURL      sql.NullString
		twitterURL       sql.NullString
		blueskyURL       sql.NullString
		instagramURL     sql.NullString
		blogURL          sql.NullString
		eventCalendarURL sql.NullString
		emailListURL     sql.NullString
		reviewNotes      sql.NullString
		reviewedAt       sql.NullTime
	)
	if err := scanner.Scan(&id, &chapterID, &status, &name, &state, &websiteURL, &officialDSAURL, &coverageNotes, &groupEmail, &linktreeURL, &facebookURL, &twitterURL, &blueskyURL, &instagramURL, &blogURL, &eventCalendarURL, &emailListURL, &reviewNotes, &reviewedAt); err != nil {
		return ChapterDraft{}, fmt.Errorf("scan chapter draft: %w", err)
	}

	chapterDraftID, err := requiredDraftString(id, "admin_chapter_drafts", "<row>", "id")
	if err != nil {
		return ChapterDraft{}, err
	}
	chapterPublicID, err := requiredDraftString(chapterID, "admin_chapter_drafts", chapterDraftID, "chapter_id")
	if err != nil {
		return ChapterDraft{}, err
	}
	draftName, err := requiredDraftString(name, "admin_chapter_drafts", chapterDraftID, "name")
	if err != nil {
		return ChapterDraft{}, err
	}
	draftState, err := requiredDraftString(state, "admin_chapter_drafts", chapterDraftID, "state")
	if err != nil {
		return ChapterDraft{}, err
	}

	return ChapterDraft{
		ID:               chapterDraftID,
		ChapterID:        chapterPublicID,
		Status:           normalizeDraftStatus(DraftStatus(optionalDraftString(status))),
		Name:             draftName,
		State:            draftState,
		WebsiteURL:       optionalDraftString(websiteURL),
		OfficialDSAURL:   optionalDraftString(officialDSAURL),
		CoverageNotes:    optionalDraftString(coverageNotes),
		GroupEmail:       optionalDraftString(groupEmail),
		LinktreeURL:      optionalDraftString(linktreeURL),
		FacebookURL:      optionalDraftString(facebookURL),
		TwitterURL:       optionalDraftString(twitterURL),
		BlueskyURL:       optionalDraftString(blueskyURL),
		InstagramURL:     optionalDraftString(instagramURL),
		BlogURL:          optionalDraftString(blogURL),
		EventCalendarURL: optionalDraftString(eventCalendarURL),
		EmailListURL:     optionalDraftString(emailListURL),
		ReviewNotes:      optionalDraftString(reviewNotes),
		ReviewedAt:       optionalDraftDate(reviewedAt),
	}, nil
}

func scanEndorsementDraftRow(scanner rowScanner) (EndorsementDraft, error) {
	var (
		id              sql.NullString
		endorsementID   sql.NullString
		status          sql.NullString
		candidateID     sql.NullString
		raceID          sql.NullString
		chapterID       sql.NullString
		voterGuideID    sql.NullString
		endorsementType sql.NullString
		sourceURL       sql.NullString
		sourceDate      sql.NullTime
		confidence      sql.NullString
		notes           sql.NullString
		reviewNotes     sql.NullString
		reviewedAt      sql.NullTime
	)
	if err := scanner.Scan(&id, &endorsementID, &status, &candidateID, &raceID, &chapterID, &voterGuideID, &endorsementType, &sourceURL, &sourceDate, &confidence, &notes, &reviewNotes, &reviewedAt); err != nil {
		return EndorsementDraft{}, fmt.Errorf("scan endorsement draft: %w", err)
	}

	endorsementDraftID, err := requiredDraftString(id, "admin_endorsement_drafts", "<row>", "id")
	if err != nil {
		return EndorsementDraft{}, err
	}
	publicID, err := requiredDraftString(endorsementID, "admin_endorsement_drafts", endorsementDraftID, "endorsement_id")
	if err != nil {
		return EndorsementDraft{}, err
	}
	candidatePublicID, err := requiredDraftString(candidateID, "admin_endorsement_drafts", endorsementDraftID, "candidate_id")
	if err != nil {
		return EndorsementDraft{}, err
	}
	racePublicID, err := requiredDraftString(raceID, "admin_endorsement_drafts", endorsementDraftID, "race_id")
	if err != nil {
		return EndorsementDraft{}, err
	}
	chapterPublicID, err := requiredDraftString(chapterID, "admin_endorsement_drafts", endorsementDraftID, "chapter_id")
	if err != nil {
		return EndorsementDraft{}, err
	}
	endorsementTypeValue, err := requiredDraftString(endorsementType, "admin_endorsement_drafts", endorsementDraftID, "endorsement_type")
	if err != nil {
		return EndorsementDraft{}, err
	}
	sourceURLValue, err := requiredDraftString(sourceURL, "admin_endorsement_drafts", endorsementDraftID, "source_url")
	if err != nil {
		return EndorsementDraft{}, err
	}
	confidenceValue, err := requiredDraftString(confidence, "admin_endorsement_drafts", endorsementDraftID, "confidence")
	if err != nil {
		return EndorsementDraft{}, err
	}

	return EndorsementDraft{
		ID:              endorsementDraftID,
		EndorsementID:   publicID,
		Status:          normalizeDraftStatus(DraftStatus(optionalDraftString(status))),
		CandidateID:     candidatePublicID,
		RaceID:          racePublicID,
		ChapterID:       chapterPublicID,
		VoterGuideID:    optionalDraftString(voterGuideID),
		EndorsementType: endorsementTypeValue,
		SourceURL:       sourceURLValue,
		SourceDate:      optionalDraftDate(sourceDate),
		Confidence:      confidenceValue,
		Notes:           optionalDraftString(notes),
		ReviewNotes:     optionalDraftString(reviewNotes),
		ReviewedAt:      optionalDraftDate(reviewedAt),
	}, nil
}

func requiredDraftString(value sql.NullString, table string, id string, field string) (string, error) {
	if !value.Valid {
		return "", fmt.Errorf("%s %s missing required field %s", table, id, field)
	}
	trimmed := strings.TrimSpace(value.String)
	if trimmed == "" {
		return "", fmt.Errorf("%s %s has blank required field %s", table, id, field)
	}
	return trimmed, nil
}

func optionalDraftString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
}

func optionalDraftDate(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format("2006-01-02")
}

func parseOptionalDate(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q", trimmed)
	}
	t := parsed.UTC()
	return &t, nil
}

func normalizeDraftStatus(status DraftStatus) DraftStatus {
	trimmed := strings.TrimSpace(string(status))
	if trimmed == "" {
		return StatusDraft
	}
	return DraftStatus(trimmed)
}

func nullableString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

var _ Store = (*PostgresStore)(nil)
