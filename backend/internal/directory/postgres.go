package directory

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Directory(ctx context.Context) (DirectoryData, error) {
	chapters := make([]Chapter, 0)
	var err error
	chapters, err = s.chapters(ctx)
	if err != nil {
		return DirectoryData{}, err
	}

	elections := make([]Election, 0)
	elections, err = s.elections(ctx)
	if err != nil {
		return DirectoryData{}, err
	}

	races := make([]Race, 0)
	races, err = s.races(ctx, RaceFilter{})
	if err != nil {
		return DirectoryData{}, err
	}

	candidates := make([]Candidate, 0)
	candidates, err = s.candidates(ctx)
	if err != nil {
		return DirectoryData{}, err
	}

	voterGuides := make([]VoterGuide, 0)
	voterGuides, err = s.voterGuides(ctx, VoterGuideFilter{})
	if err != nil {
		return DirectoryData{}, err
	}

	endorsements := make([]Endorsement, 0)
	endorsements, err = s.endorsements(ctx, EndorsementFilter{})
	if err != nil {
		return DirectoryData{}, err
	}

	ballotMeasures := make([]BallotMeasure, 0)
	ballotMeasures, err = s.ballotMeasures(ctx, BallotMeasureFilter{})
	if err != nil {
		return DirectoryData{}, err
	}

	measurePositions := make([]MeasurePosition, 0)
	measurePositions, err = s.measurePositions(ctx, MeasurePositionFilter{})
	if err != nil {
		return DirectoryData{}, err
	}

	return DirectoryData{
		Chapters:         chapters,
		Elections:        elections,
		Races:            races,
		Candidates:       candidates,
		VoterGuides:      voterGuides,
		Endorsements:     endorsements,
		BallotMeasures:   ballotMeasures,
		MeasurePositions: measurePositions,
	}, nil
}

func (s *PostgresStore) Chapters(ctx context.Context) ([]Chapter, error) {
	return s.chapters(ctx)
}

func (s *PostgresStore) Elections(ctx context.Context) ([]Election, error) {
	return s.elections(ctx)
}

func (s *PostgresStore) Races(ctx context.Context, filter RaceFilter) ([]Race, error) {
	return s.races(ctx, filter)
}

func (s *PostgresStore) Candidates(ctx context.Context) ([]Candidate, error) {
	return s.candidates(ctx)
}

func (s *PostgresStore) VoterGuides(ctx context.Context, filter VoterGuideFilter) ([]VoterGuide, error) {
	return s.voterGuides(ctx, filter)
}

func (s *PostgresStore) Endorsements(ctx context.Context, filter EndorsementFilter) ([]Endorsement, error) {
	return s.endorsements(ctx, filter)
}

func (s *PostgresStore) BallotMeasures(ctx context.Context, filter BallotMeasureFilter) ([]BallotMeasure, error) {
	return s.ballotMeasures(ctx, filter)
}

func (s *PostgresStore) MeasurePositions(ctx context.Context, filter MeasurePositionFilter) ([]MeasurePosition, error) {
	return s.measurePositions(ctx, filter)
}

func (s *PostgresStore) chapters(ctx context.Context) ([]Chapter, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, state, website_url, official_dsa_url, coverage_notes, active
		FROM chapters
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query chapters: %w", err)
	}
	defer rows.Close()

	chapters := make([]Chapter, 0)
	for rows.Next() {
		var (
			id             sql.NullString
			name           sql.NullString
			state          sql.NullString
			websiteURL     sql.NullString
			officialDSAURL sql.NullString
			coverageNotes  sql.NullString
			active         sql.NullBool
		)

		if err := rows.Scan(&id, &name, &state, &websiteURL, &officialDSAURL, &coverageNotes, &active); err != nil {
			return nil, fmt.Errorf("scan chapters: %w", err)
		}

		chapterID, err := requiredString(id, "chapters", "<row>", "id")
		if err != nil {
			return nil, err
		}
		chapterName, err := requiredString(name, "chapters", chapterID, "name")
		if err != nil {
			return nil, err
		}

		chapters = append(chapters, Chapter{
			ID:             chapterID,
			Name:           chapterName,
			State:          optionalString(state),
			WebsiteURL:     optionalString(websiteURL),
			OfficialDSAURL: optionalString(officialDSAURL),
			CoverageNotes:  optionalString(coverageNotes),
			Active:         active.Valid && active.Bool,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chapters: %w", err)
	}

	return chapters, nil
}

func (s *PostgresStore) elections(ctx context.Context) ([]Election, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, election_date, election_type, jurisdiction, official_url
		FROM elections
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query elections: %w", err)
	}
	defer rows.Close()

	elections := make([]Election, 0)
	for rows.Next() {
		var (
			id           sql.NullString
			name         sql.NullString
			electionDate sql.NullTime
			electionType sql.NullString
			jurisdiction sql.NullString
			officialURL  sql.NullString
		)

		if err := rows.Scan(&id, &name, &electionDate, &electionType, &jurisdiction, &officialURL); err != nil {
			return nil, fmt.Errorf("scan elections: %w", err)
		}

		electionID, err := requiredString(id, "elections", "<row>", "id")
		if err != nil {
			return nil, err
		}
		electionName, err := requiredString(name, "elections", electionID, "name")
		if err != nil {
			return nil, err
		}

		elections = append(elections, Election{
			ID:           electionID,
			Name:         electionName,
			ElectionDate: optionalDate(electionDate),
			ElectionType: optionalString(electionType),
			Jurisdiction: optionalString(jurisdiction),
			OfficialURL:  optionalString(officialURL),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate elections: %w", err)
	}

	return elections, nil
}

func (s *PostgresStore) races(ctx context.Context, filter RaceFilter) ([]Race, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT id, election_id, office, district, level, jurisdiction
		FROM races
	`)

	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if filter.ElectionID != "" {
		args = append(args, filter.ElectionID)
		clauses = append(clauses, fmt.Sprintf("election_id = $%d", len(args)))
	}
	if filter.Jurisdiction != "" {
		args = append(args, filter.Jurisdiction)
		clauses = append(clauses, fmt.Sprintf("jurisdiction = $%d", len(args)))
	}
	if filter.Level != "" {
		args = append(args, filter.Level)
		clauses = append(clauses, fmt.Sprintf("level = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query.WriteString("\tWHERE ")
		query.WriteString(strings.Join(clauses, " AND "))
		query.WriteString("\n")
	}
	query.WriteString("\tORDER BY id\n")

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query races: %w", err)
	}
	defer rows.Close()

	races := make([]Race, 0)
	for rows.Next() {
		var (
			id           sql.NullString
			electionID   sql.NullString
			office       sql.NullString
			district     sql.NullString
			level        sql.NullString
			jurisdiction sql.NullString
		)

		if err := rows.Scan(&id, &electionID, &office, &district, &level, &jurisdiction); err != nil {
			return nil, fmt.Errorf("scan races: %w", err)
		}

		raceID, err := requiredString(id, "races", "<row>", "id")
		if err != nil {
			return nil, err
		}
		requiredElectionID, err := requiredString(electionID, "races", raceID, "election_id")
		if err != nil {
			return nil, err
		}
		requiredOffice, err := requiredString(office, "races", raceID, "office")
		if err != nil {
			return nil, err
		}
		requiredLevel, err := requiredString(level, "races", raceID, "level")
		if err != nil {
			return nil, err
		}
		requiredJurisdiction, err := requiredString(jurisdiction, "races", raceID, "jurisdiction")
		if err != nil {
			return nil, err
		}

		races = append(races, Race{
			ID:           raceID,
			ElectionID:   requiredElectionID,
			Office:       requiredOffice,
			District:     optionalString(district),
			Level:        requiredLevel,
			Jurisdiction: requiredJurisdiction,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate races: %w", err)
	}

	return races, nil
}

func (s *PostgresStore) candidates(ctx context.Context) ([]Candidate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, campaign_url, party, incumbent, photo_url
		FROM candidates
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query candidates: %w", err)
	}
	defer rows.Close()

	candidates := make([]Candidate, 0)
	for rows.Next() {
		var (
			id          sql.NullString
			name        sql.NullString
			campaignURL sql.NullString
			party       sql.NullString
			incumbent   sql.NullBool
			photoURL    sql.NullString
		)

		if err := rows.Scan(&id, &name, &campaignURL, &party, &incumbent, &photoURL); err != nil {
			return nil, fmt.Errorf("scan candidates: %w", err)
		}

		candidateID, err := requiredString(id, "candidates", "<row>", "id")
		if err != nil {
			return nil, err
		}
		candidateName, err := requiredString(name, "candidates", candidateID, "name")
		if err != nil {
			return nil, err
		}

		candidates = append(candidates, Candidate{
			ID:          candidateID,
			Name:        candidateName,
			CampaignURL: optionalString(campaignURL),
			Party:       optionalString(party),
			Incumbent:   incumbent.Valid && incumbent.Bool,
			PhotoURL:    optionalString(photoURL),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate candidates: %w", err)
	}

	return candidates, nil
}

func (s *PostgresStore) voterGuides(ctx context.Context, filter VoterGuideFilter) ([]VoterGuide, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT id, title, chapter_id, election_id, source_url, confidence, published_at, summary
		FROM voter_guides
	`)

	clauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if filter.ChapterID != "" {
		args = append(args, filter.ChapterID)
		clauses = append(clauses, fmt.Sprintf("chapter_id = $%d", len(args)))
	}
	if filter.ElectionID != "" {
		args = append(args, filter.ElectionID)
		clauses = append(clauses, fmt.Sprintf("election_id = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query.WriteString("\tWHERE ")
		query.WriteString(strings.Join(clauses, " AND "))
		query.WriteString("\n")
	}
	query.WriteString("\tORDER BY id\n")

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query voter_guides: %w", err)
	}
	defer rows.Close()

	voterGuides := make([]VoterGuide, 0)
	for rows.Next() {
		var (
			id          sql.NullString
			title       sql.NullString
			chapterID   sql.NullString
			electionID  sql.NullString
			sourceURL   sql.NullString
			confidence  sql.NullString
			publishedAt sql.NullTime
			summary     sql.NullString
		)

		if err := rows.Scan(&id, &title, &chapterID, &electionID, &sourceURL, &confidence, &publishedAt, &summary); err != nil {
			return nil, fmt.Errorf("scan voter_guides: %w", err)
		}

		guideID, err := requiredString(id, "voter_guides", "<row>", "id")
		if err != nil {
			return nil, err
		}
		requiredTitle, err := requiredString(title, "voter_guides", guideID, "title")
		if err != nil {
			return nil, err
		}
		requiredChapterID, err := requiredString(chapterID, "voter_guides", guideID, "chapter_id")
		if err != nil {
			return nil, err
		}
		requiredSourceURL, err := requiredString(sourceURL, "voter_guides", guideID, "source_url")
		if err != nil {
			return nil, err
		}
		requiredConfidence, err := requiredString(confidence, "voter_guides", guideID, "confidence")
		if err != nil {
			return nil, err
		}

		voterGuides = append(voterGuides, VoterGuide{
			ID:          guideID,
			Title:       requiredTitle,
			ChapterID:   requiredChapterID,
			ElectionID:  optionalString(electionID),
			SourceURL:   requiredSourceURL,
			Confidence:  requiredConfidence,
			PublishedAt: optionalDate(publishedAt),
			Summary:     optionalString(summary),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate voter_guides: %w", err)
	}

	return voterGuides, nil
}

func (s *PostgresStore) endorsements(ctx context.Context, filter EndorsementFilter) ([]Endorsement, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT e.id, e.candidate_id, e.race_id, e.chapter_id, e.voter_guide_id, e.endorsement_type, e.source_url, e.source_date, e.confidence, e.notes
		FROM endorsements e
	`)

	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if filter.ElectionID != "" {
		query.WriteString("\tJOIN races r ON r.id = e.race_id\n")
		args = append(args, filter.ElectionID)
		clauses = append(clauses, fmt.Sprintf("r.election_id = $%d", len(args)))
	}
	if filter.ChapterID != "" {
		args = append(args, filter.ChapterID)
		clauses = append(clauses, fmt.Sprintf("e.chapter_id = $%d", len(args)))
	}
	if filter.RaceID != "" {
		args = append(args, filter.RaceID)
		clauses = append(clauses, fmt.Sprintf("e.race_id = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query.WriteString("\tWHERE ")
		query.WriteString(strings.Join(clauses, " AND "))
		query.WriteString("\n")
	}
	query.WriteString("\tORDER BY e.id\n")

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query endorsements: %w", err)
	}
	defer rows.Close()

	endorsements := make([]Endorsement, 0)
	for rows.Next() {
		var (
			id              sql.NullString
			candidateID     sql.NullString
			raceID          sql.NullString
			chapterID       sql.NullString
			voterGuideID    sql.NullString
			endorsementType sql.NullString
			sourceURL       sql.NullString
			sourceDate      sql.NullTime
			confidence      sql.NullString
			notes           sql.NullString
		)

		if err := rows.Scan(&id, &candidateID, &raceID, &chapterID, &voterGuideID, &endorsementType, &sourceURL, &sourceDate, &confidence, &notes); err != nil {
			return nil, fmt.Errorf("scan endorsements: %w", err)
		}

		endorsementID, err := requiredString(id, "endorsements", "<row>", "id")
		if err != nil {
			return nil, err
		}
		requiredCandidateID, err := requiredString(candidateID, "endorsements", endorsementID, "candidate_id")
		if err != nil {
			return nil, err
		}
		requiredRaceID, err := requiredString(raceID, "endorsements", endorsementID, "race_id")
		if err != nil {
			return nil, err
		}
		requiredChapterID, err := requiredString(chapterID, "endorsements", endorsementID, "chapter_id")
		if err != nil {
			return nil, err
		}
		requiredType, err := requiredString(endorsementType, "endorsements", endorsementID, "endorsement_type")
		if err != nil {
			return nil, err
		}
		requiredSourceURL, err := requiredString(sourceURL, "endorsements", endorsementID, "source_url")
		if err != nil {
			return nil, err
		}
		requiredConfidence, err := requiredString(confidence, "endorsements", endorsementID, "confidence")
		if err != nil {
			return nil, err
		}

		endorsements = append(endorsements, Endorsement{
			ID:              endorsementID,
			CandidateID:     requiredCandidateID,
			RaceID:          requiredRaceID,
			ChapterID:       requiredChapterID,
			VoterGuideID:    optionalString(voterGuideID),
			EndorsementType: requiredType,
			SourceURL:       requiredSourceURL,
			SourceDate:      optionalDate(sourceDate),
			Confidence:      requiredConfidence,
			Notes:           optionalString(notes),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate endorsements: %w", err)
	}

	return endorsements, nil
}

func (s *PostgresStore) ballotMeasures(ctx context.Context, filter BallotMeasureFilter) ([]BallotMeasure, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT id, election_id, title, jurisdiction, measure_number, summary
		FROM ballot_measures
	`)

	clauses := make([]string, 0, 1)
	args := make([]any, 0, 1)
	if filter.ElectionID != "" {
		args = append(args, filter.ElectionID)
		clauses = append(clauses, fmt.Sprintf("election_id = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query.WriteString("\tWHERE ")
		query.WriteString(strings.Join(clauses, " AND "))
		query.WriteString("\n")
	}
	query.WriteString("\tORDER BY id\n")

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query ballot_measures: %w", err)
	}
	defer rows.Close()

	ballotMeasures := make([]BallotMeasure, 0)
	for rows.Next() {
		var (
			id            sql.NullString
			electionID    sql.NullString
			title         sql.NullString
			jurisdiction  sql.NullString
			measureNumber sql.NullString
			summary       sql.NullString
		)

		if err := rows.Scan(&id, &electionID, &title, &jurisdiction, &measureNumber, &summary); err != nil {
			return nil, fmt.Errorf("scan ballot_measures: %w", err)
		}

		measureID, err := requiredString(id, "ballot_measures", "<row>", "id")
		if err != nil {
			return nil, err
		}
		requiredElectionID, err := requiredString(electionID, "ballot_measures", measureID, "election_id")
		if err != nil {
			return nil, err
		}
		requiredTitle, err := requiredString(title, "ballot_measures", measureID, "title")
		if err != nil {
			return nil, err
		}
		requiredJurisdiction, err := requiredString(jurisdiction, "ballot_measures", measureID, "jurisdiction")
		if err != nil {
			return nil, err
		}

		ballotMeasures = append(ballotMeasures, BallotMeasure{
			ID:            measureID,
			ElectionID:    requiredElectionID,
			Title:         requiredTitle,
			Jurisdiction:  requiredJurisdiction,
			MeasureNumber: optionalString(measureNumber),
			Summary:       optionalString(summary),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ballot_measures: %w", err)
	}

	return ballotMeasures, nil
}

func (s *PostgresStore) measurePositions(ctx context.Context, filter MeasurePositionFilter) ([]MeasurePosition, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT id, measure_id, chapter_id, position, source_url, confidence
		FROM measure_positions
	`)

	clauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if filter.ChapterID != "" {
		args = append(args, filter.ChapterID)
		clauses = append(clauses, fmt.Sprintf("chapter_id = $%d", len(args)))
	}
	if filter.MeasureID != "" {
		args = append(args, filter.MeasureID)
		clauses = append(clauses, fmt.Sprintf("measure_id = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query.WriteString("\tWHERE ")
		query.WriteString(strings.Join(clauses, " AND "))
		query.WriteString("\n")
	}
	query.WriteString("\tORDER BY id\n")

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query measure_positions: %w", err)
	}
	defer rows.Close()

	measurePositions := make([]MeasurePosition, 0)
	for rows.Next() {
		var (
			id         sql.NullString
			measureID  sql.NullString
			chapterID  sql.NullString
			position   sql.NullString
			sourceURL  sql.NullString
			confidence sql.NullString
		)

		if err := rows.Scan(&id, &measureID, &chapterID, &position, &sourceURL, &confidence); err != nil {
			return nil, fmt.Errorf("scan measure_positions: %w", err)
		}

		positionID, err := requiredString(id, "measure_positions", "<row>", "id")
		if err != nil {
			return nil, err
		}
		requiredMeasureID, err := requiredString(measureID, "measure_positions", positionID, "measure_id")
		if err != nil {
			return nil, err
		}
		requiredChapterID, err := requiredString(chapterID, "measure_positions", positionID, "chapter_id")
		if err != nil {
			return nil, err
		}
		requiredPosition, err := requiredString(position, "measure_positions", positionID, "position")
		if err != nil {
			return nil, err
		}
		requiredSourceURL, err := requiredString(sourceURL, "measure_positions", positionID, "source_url")
		if err != nil {
			return nil, err
		}
		requiredConfidence, err := requiredString(confidence, "measure_positions", positionID, "confidence")
		if err != nil {
			return nil, err
		}

		measurePositions = append(measurePositions, MeasurePosition{
			ID:         positionID,
			MeasureID:  requiredMeasureID,
			ChapterID:  requiredChapterID,
			Position:   requiredPosition,
			SourceURL:  requiredSourceURL,
			Confidence: requiredConfidence,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate measure_positions: %w", err)
	}

	return measurePositions, nil
}

func requiredString(value sql.NullString, table string, id string, field string) (string, error) {
	if !value.Valid {
		return "", fmt.Errorf("%s %s missing required field %s", table, id, field)
	}

	trimmed := strings.TrimSpace(value.String)
	if trimmed == "" {
		return "", fmt.Errorf("%s %s has blank required field %s", table, id, field)
	}

	return trimmed, nil
}

func optionalString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}

	return strings.TrimSpace(value.String)
}

func optionalDate(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}

	return value.Time.Format("2006-01-02")
}

var _ Store = (*PostgresStore)(nil)
