package directory

import (
	"strings"
	"testing"
)

func TestQualityReportPassesCleanSourceLinkedData(t *testing.T) {
	data := DirectoryData{
		Chapters:         []Chapter{{ID: "dsa-san-diego", Name: "DSA San Diego", State: "CA", WebsiteURL: "https://dsasandiego.org/", Active: true}},
		Elections:        []Election{{ID: "2026-ca-primary", Name: "2026 California Primary", ElectionDate: "2026-06-02", Jurisdiction: "CA"}},
		Races:            []Race{{ID: "ca-48-us-house-2026-primary", ElectionID: "2026-ca-primary", Office: "U.S. House", District: "CA-48", Level: "federal", Jurisdiction: "CA-48"}},
		Candidates:       []Candidate{{ID: "corinna-contreras", Name: "Corinna Contreras"}},
		VoterGuides:      []VoterGuide{{ID: "dsa-san-diego-2026-primary-guide", Title: "DSA San Diego 2026 Primary Guide", ChapterID: "dsa-san-diego", ElectionID: "2026-ca-primary", SourceURL: "https://dsasandiego.org/guide", Confidence: "official_source", PublishedAt: "2026-05-05"}},
		Endorsements:     []Endorsement{{ID: "contreras-dsa-sd-2026-primary", CandidateID: "corinna-contreras", RaceID: "ca-48-us-house-2026-primary", ChapterID: "dsa-san-diego", VoterGuideID: "dsa-san-diego-2026-primary-guide", EndorsementType: "recommended", SourceURL: "https://dsasandiego.org/guide", SourceDate: "2026-05-05", Confidence: "official_source", Notes: "Preserve recommendation language."}},
		BallotMeasures:   []BallotMeasure{{ID: "san-diego-measure-a-2026-primary", ElectionID: "2026-ca-primary", Title: "San Diego Measure A", Jurisdiction: "San Diego, CA", MeasureNumber: "A"}},
		MeasurePositions: []MeasurePosition{{ID: "san-diego-measure-a-dsa-sd-yes-2026-primary", MeasureID: "san-diego-measure-a-2026-primary", ChapterID: "dsa-san-diego", Position: "yes", SourceURL: "https://dsasandiego.org/guide", Confidence: "official_source"}},
	}

	report := EvaluateQuality(data, QualityOptions{ReferenceDate: "2026-06-28", StaleAfterDays: 180})

	if report.HasFailures() {
		t.Fatalf("expected no failed findings, got:\n%s", report.String())
	}
	if got := report.String(); got != "PASS data quality checks\n" {
		t.Fatalf("report.String() = %q, want PASS banner", got)
	}
}

func TestQualityReportFlagsMissingAndWeakSourceContext(t *testing.T) {
	data := DirectoryData{
		Chapters:         []Chapter{{ID: "dsa-san-diego", Name: "DSA San Diego", State: "CA", Active: true}},
		Elections:        []Election{{ID: "2026-ca-primary", Name: "2026 California Primary"}},
		Races:            []Race{{ID: "ca-48-us-house-2026-primary", ElectionID: "2026-ca-primary", Office: "U.S. House", Level: "federal", Jurisdiction: "CA-48"}},
		Candidates:       []Candidate{{ID: "corinna-contreras", Name: "Corinna Contreras"}},
		VoterGuides:      []VoterGuide{{ID: "dsa-san-diego-2026-primary-guide", Title: "Guide", ChapterID: "dsa-san-diego", SourceURL: "", Confidence: "needs_review"}},
		Endorsements:     []Endorsement{{ID: "contreras-dsa-sd-2026-primary", CandidateID: "corinna-contreras", RaceID: "ca-48-us-house-2026-primary", ChapterID: "dsa-san-diego", EndorsementType: "endorsed", SourceURL: "", Confidence: "needs_review"}},
		MeasurePositions: []MeasurePosition{{ID: "measure-position", MeasureID: "missing-measure", ChapterID: "dsa-san-diego", Position: "yes", SourceURL: "", Confidence: "needs_review"}},
	}

	report := EvaluateQuality(data, QualityOptions{ReferenceDate: "2026-06-28", StaleAfterDays: 180})
	output := report.String()

	for _, want := range []string{
		"FAIL voterGuides[dsa-san-diego-2026-primary-guide] missing sourceUrl",
		"WARN voterGuides[dsa-san-diego-2026-primary-guide] confidence needs_review requires manual review",
		"FAIL endorsements[contreras-dsa-sd-2026-primary] missing sourceUrl",
		"WARN endorsements[contreras-dsa-sd-2026-primary] missing sourceDate",
		"FAIL measurePositions[measure-position] missing sourceUrl",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected report to contain %q, got:\n%s", want, output)
		}
	}
}

func TestQualityReportFlagsStaleSourceDatesAndBrokenCoverage(t *testing.T) {
	data := DirectoryData{
		Chapters:     []Chapter{{ID: "dsa-san-diego", Name: "DSA San Diego", State: "CA", Active: true}},
		Elections:    []Election{{ID: "2026-ca-primary", Name: "2026 California Primary"}},
		Races:        []Race{{ID: "ca-48-us-house-2026-primary", ElectionID: "2026-ca-primary", Office: "U.S. House", Level: "federal", Jurisdiction: "CA-48"}},
		Candidates:   []Candidate{{ID: "corinna-contreras", Name: "Corinna Contreras"}},
		VoterGuides:  []VoterGuide{{ID: "dsa-san-diego-2026-primary-guide", Title: "DSA San Diego 2026 Primary Guide", ChapterID: "dsa-san-diego", SourceURL: "https://dsasandiego.org/guide", Confidence: "official_source"}},
		Endorsements: []Endorsement{{ID: "contreras-dsa-sd-2026-primary", CandidateID: "corinna-contreras", RaceID: "ca-48-us-house-2026-primary", ChapterID: "dsa-san-diego", VoterGuideID: "missing-guide", EndorsementType: "recommended", SourceURL: "https://dsasandiego.org/guide", SourceDate: "2025-01-01", Confidence: "official_source"}},
	}

	report := EvaluateQuality(data, QualityOptions{ReferenceDate: "2026-06-28", StaleAfterDays: 180})
	output := report.String()

	for _, want := range []string{
		"WARN endorsements[contreras-dsa-sd-2026-primary] sourceDate is older than 180 days",
		"FAIL endorsements[contreras-dsa-sd-2026-primary] voterGuideId references missing voterGuides[missing-guide]",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected report to contain %q, got:\n%s", want, output)
		}
	}
}
