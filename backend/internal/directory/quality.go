package directory

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type QualitySeverity string

const (
	QualityFail QualitySeverity = "FAIL"
	QualityWarn QualitySeverity = "WARN"
)

type QualityOptions struct {
	ReferenceDate  string
	StaleAfterDays int
}

type QualityFinding struct {
	Severity QualitySeverity
	Record   string
	Message  string
}

type QualityReport struct {
	Findings []QualityFinding
}

func (r QualityReport) HasFailures() bool {
	for _, finding := range r.Findings {
		if finding.Severity == QualityFail {
			return true
		}
	}
	return false
}

func (r QualityReport) String() string {
	if len(r.Findings) == 0 {
		return "PASS data quality checks\n"
	}

	findings := append([]QualityFinding(nil), r.Findings...)
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return findings[i].Severity < findings[j].Severity
		}
		if findings[i].Record != findings[j].Record {
			return findings[i].Record < findings[j].Record
		}
		return findings[i].Message < findings[j].Message
	})

	var builder strings.Builder
	for _, finding := range findings {
		builder.WriteString(fmt.Sprintf("%s %s %s\n", finding.Severity, finding.Record, finding.Message))
	}
	return builder.String()
}

func EvaluateQuality(data DirectoryData, options QualityOptions) QualityReport {
	report := QualityReport{Findings: make([]QualityFinding, 0)}
	data = NormalizeDirectoryData(data)

	staleAfterDays := options.StaleAfterDays
	if staleAfterDays <= 0 {
		staleAfterDays = 180
	}

	referenceDate := time.Now().UTC()
	if options.ReferenceDate != "" {
		parsed, err := time.Parse("2006-01-02", options.ReferenceDate)
		if err != nil {
			report.add(QualityWarn, "qualityOptions", "referenceDate must be YYYY-MM-DD")
		} else {
			referenceDate = parsed.UTC()
		}
	}

	index := newQualityIndex(data)

	for _, guide := range data.VoterGuides {
		record := fmt.Sprintf("voterGuides[%s]", guide.ID)
		report.requireReference(record, "chapterId", guide.ChapterID, index.chapters)
		report.requireOptionalReference(record, "electionId", guide.ElectionID, index.elections)
		report.requireSourceURL(record, guide.SourceURL)
		report.warnWeakConfidence(record, guide.Confidence)
		hasReviewMetadata := report.validateReviewMetadata(record, guide.ReviewedAt, guide.ReviewNotes)
		if guide.PublishedAt == "" {
			if !hasReviewMetadata {
				report.add(QualityWarn, record, "missing publishedAt")
			}
		} else {
			report.warnIfStale(record, "publishedAt", guide.PublishedAt, referenceDate, staleAfterDays)
		}
	}

	for _, endorsement := range data.Endorsements {
		record := fmt.Sprintf("endorsements[%s]", endorsement.ID)
		report.requireReference(record, "candidateId", endorsement.CandidateID, index.candidates)
		report.requireReference(record, "raceId", endorsement.RaceID, index.races)
		report.requireReference(record, "chapterId", endorsement.ChapterID, index.chapters)
		report.requireOptionalReference(record, "voterGuideId", endorsement.VoterGuideID, index.voterGuides)
		report.requireSourceURL(record, endorsement.SourceURL)
		report.warnWeakConfidence(record, endorsement.Confidence)
		hasReviewMetadata := report.validateReviewMetadata(record, endorsement.ReviewedAt, endorsement.ReviewNotes)
		if endorsement.SourceDate == "" {
			if !hasReviewMetadata {
				report.add(QualityWarn, record, "missing sourceDate")
			}
		} else {
			report.warnIfStale(record, "sourceDate", endorsement.SourceDate, referenceDate, staleAfterDays)
		}
		if endorsement.EndorsementType == "unknown" {
			report.add(QualityWarn, record, "endorsementType unknown requires manual review")
		}
	}

	for _, measure := range data.BallotMeasures {
		record := fmt.Sprintf("ballotMeasures[%s]", measure.ID)
		report.requireReference(record, "electionId", measure.ElectionID, index.elections)
	}

	for _, position := range data.MeasurePositions {
		record := fmt.Sprintf("measurePositions[%s]", position.ID)
		report.requireReference(record, "measureId", position.MeasureID, index.ballotMeasures)
		report.requireReference(record, "chapterId", position.ChapterID, index.chapters)
		report.requireSourceURL(record, position.SourceURL)
		report.warnWeakConfidence(record, position.Confidence)
		report.validateReviewMetadata(record, position.ReviewedAt, position.ReviewNotes)
		if position.Position == "unknown" {
			report.add(QualityWarn, record, "position unknown requires manual review")
		}
	}

	return report
}

func (r *QualityReport) add(severity QualitySeverity, record string, message string) {
	r.Findings = append(r.Findings, QualityFinding{Severity: severity, Record: record, Message: message})
}

func (r *QualityReport) requireSourceURL(record string, sourceURL string) {
	if strings.TrimSpace(sourceURL) == "" {
		r.add(QualityFail, record, "missing sourceUrl")
	}
}

func (r *QualityReport) validateReviewMetadata(record string, reviewedAt string, reviewNotes string) bool {
	reviewedAt = strings.TrimSpace(reviewedAt)
	reviewNotes = strings.TrimSpace(reviewNotes)
	if reviewedAt == "" && reviewNotes == "" {
		return false
	}
	if reviewedAt == "" {
		r.add(QualityWarn, record, "missing reviewedAt for reviewNotes")
		return false
	}
	if _, err := time.Parse("2006-01-02", reviewedAt); err != nil {
		r.add(QualityWarn, record, "reviewedAt must be YYYY-MM-DD")
		return false
	}
	if reviewNotes == "" {
		r.add(QualityWarn, record, "missing reviewNotes for reviewedAt")
		return false
	}
	return true
}

func (r *QualityReport) warnWeakConfidence(record string, confidence string) {
	switch confidence {
	case "needs_review", "campaign_source", "press_source":
		r.add(QualityWarn, record, fmt.Sprintf("confidence %s requires manual review", confidence))
	}
}

func (r *QualityReport) warnIfStale(record string, fieldName string, value string, referenceDate time.Time, staleAfterDays int) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		r.add(QualityWarn, record, fmt.Sprintf("%s must be YYYY-MM-DD", fieldName))
		return
	}

	if referenceDate.Sub(date.UTC()) > time.Duration(staleAfterDays)*24*time.Hour {
		r.add(QualityWarn, record, fmt.Sprintf("%s is older than %d days", fieldName, staleAfterDays))
	}
}

func (r *QualityReport) requireReference(record string, fieldName string, value string, ids map[string]struct{}) {
	if strings.TrimSpace(value) == "" {
		r.add(QualityFail, record, fmt.Sprintf("missing %s", fieldName))
		return
	}
	if _, ok := ids[value]; !ok {
		r.add(QualityFail, record, fmt.Sprintf("%s references missing %s[%s]", fieldName, fieldNameTarget(fieldName), value))
	}
}

func (r *QualityReport) requireOptionalReference(record string, fieldName string, value string, ids map[string]struct{}) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if _, ok := ids[value]; !ok {
		r.add(QualityFail, record, fmt.Sprintf("%s references missing %s[%s]", fieldName, fieldNameTarget(fieldName), value))
	}
}

func fieldNameTarget(fieldName string) string {
	switch fieldName {
	case "chapterId":
		return "chapters"
	case "electionId":
		return "elections"
	case "raceId":
		return "races"
	case "candidateId":
		return "candidates"
	case "voterGuideId":
		return "voterGuides"
	case "measureId":
		return "ballotMeasures"
	default:
		return fieldName
	}
}

type qualityIndex struct {
	chapters       map[string]struct{}
	elections      map[string]struct{}
	races          map[string]struct{}
	candidates     map[string]struct{}
	voterGuides    map[string]struct{}
	ballotMeasures map[string]struct{}
}

func newQualityIndex(data DirectoryData) qualityIndex {
	return qualityIndex{
		chapters:       chapterIDSet(data.Chapters),
		elections:      electionIDSet(data.Elections),
		races:          raceIDSet(data.Races),
		candidates:     candidateIDSet(data.Candidates),
		voterGuides:    voterGuideIDSet(data.VoterGuides),
		ballotMeasures: ballotMeasureIDSet(data.BallotMeasures),
	}
}

func chapterIDSet(records []Chapter) map[string]struct{} {
	ids := make(map[string]struct{}, len(records))
	for _, record := range records {
		ids[record.ID] = struct{}{}
	}
	return ids
}

func electionIDSet(records []Election) map[string]struct{} {
	ids := make(map[string]struct{}, len(records))
	for _, record := range records {
		ids[record.ID] = struct{}{}
	}
	return ids
}

func raceIDSet(records []Race) map[string]struct{} {
	ids := make(map[string]struct{}, len(records))
	for _, record := range records {
		ids[record.ID] = struct{}{}
	}
	return ids
}

func candidateIDSet(records []Candidate) map[string]struct{} {
	ids := make(map[string]struct{}, len(records))
	for _, record := range records {
		ids[record.ID] = struct{}{}
	}
	return ids
}

func voterGuideIDSet(records []VoterGuide) map[string]struct{} {
	ids := make(map[string]struct{}, len(records))
	for _, record := range records {
		ids[record.ID] = struct{}{}
	}
	return ids
}

func ballotMeasureIDSet(records []BallotMeasure) map[string]struct{} {
	ids := make(map[string]struct{}, len(records))
	for _, record := range records {
		ids[record.ID] = struct{}{}
	}
	return ids
}
