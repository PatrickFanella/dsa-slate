package directory

import "context"

type RaceFilter struct {
	ElectionID   string
	Jurisdiction string
	Level        string
}

type VoterGuideFilter struct {
	ChapterID  string
	ElectionID string
}

type EndorsementFilter struct {
	ChapterID  string
	ElectionID string
	RaceID     string
}

type BallotMeasureFilter struct {
	ElectionID string
}

type MeasurePositionFilter struct {
	ChapterID string
	MeasureID string
}

type Store interface {
	Directory(ctx context.Context) (DirectoryData, error)
	Chapters(ctx context.Context) ([]Chapter, error)
	Elections(ctx context.Context) ([]Election, error)
	Races(ctx context.Context, filter RaceFilter) ([]Race, error)
	Candidates(ctx context.Context) ([]Candidate, error)
	VoterGuides(ctx context.Context, filter VoterGuideFilter) ([]VoterGuide, error)
	Endorsements(ctx context.Context, filter EndorsementFilter) ([]Endorsement, error)
	BallotMeasures(ctx context.Context, filter BallotMeasureFilter) ([]BallotMeasure, error)
	MeasurePositions(ctx context.Context, filter MeasurePositionFilter) ([]MeasurePosition, error)
}

type Chapter struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	State          string `json:"state"`
	WebsiteURL     string `json:"websiteUrl,omitempty"`
	OfficialDSAURL string `json:"officialDsaUrl,omitempty"`
	CoverageNotes  string `json:"coverageNotes,omitempty"`
	Active         bool   `json:"active"`
}

type Election struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ElectionDate string `json:"electionDate,omitempty"`
	ElectionType string `json:"electionType,omitempty"`
	Jurisdiction string `json:"jurisdiction,omitempty"`
	OfficialURL  string `json:"officialUrl,omitempty"`
}

type Race struct {
	ID           string `json:"id"`
	ElectionID   string `json:"electionId"`
	Office       string `json:"office"`
	District     string `json:"district,omitempty"`
	Level        string `json:"level"`
	Jurisdiction string `json:"jurisdiction"`
}

type Candidate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CampaignURL string `json:"campaignUrl,omitempty"`
	Party       string `json:"party,omitempty"`
	Incumbent   bool   `json:"incumbent"`
	PhotoURL    string `json:"photoUrl,omitempty"`
}

type VoterGuide struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ChapterID   string `json:"chapterId"`
	ElectionID  string `json:"electionId,omitempty"`
	SourceURL   string `json:"sourceUrl"`
	Confidence  string `json:"confidence"`
	PublishedAt string `json:"publishedAt,omitempty"`
	Summary     string `json:"summary,omitempty"`
	ReviewedAt  string `json:"reviewedAt,omitempty"`
	ReviewNotes string `json:"reviewNotes,omitempty"`
}

type Endorsement struct {
	ID              string `json:"id"`
	CandidateID     string `json:"candidateId"`
	RaceID          string `json:"raceId"`
	ChapterID       string `json:"chapterId"`
	VoterGuideID    string `json:"voterGuideId,omitempty"`
	EndorsementType string `json:"endorsementType"`
	SourceURL       string `json:"sourceUrl"`
	SourceDate      string `json:"sourceDate,omitempty"`
	Confidence      string `json:"confidence"`
	Notes           string `json:"notes,omitempty"`
	ReviewedAt      string `json:"reviewedAt,omitempty"`
	ReviewNotes     string `json:"reviewNotes,omitempty"`
}

type BallotMeasure struct {
	ID            string `json:"id"`
	ElectionID    string `json:"electionId"`
	Title         string `json:"title"`
	Jurisdiction  string `json:"jurisdiction"`
	MeasureNumber string `json:"measureNumber,omitempty"`
	Summary       string `json:"summary,omitempty"`
}

type MeasurePosition struct {
	ID          string `json:"id"`
	MeasureID   string `json:"measureId"`
	ChapterID   string `json:"chapterId"`
	Position    string `json:"position"`
	SourceURL   string `json:"sourceUrl"`
	Confidence  string `json:"confidence"`
	ReviewedAt  string `json:"reviewedAt,omitempty"`
	ReviewNotes string `json:"reviewNotes,omitempty"`
}

type DirectoryData struct {
	Chapters         []Chapter         `json:"chapters"`
	Elections        []Election        `json:"elections"`
	Races            []Race            `json:"races"`
	Candidates       []Candidate       `json:"candidates"`
	VoterGuides      []VoterGuide      `json:"voterGuides"`
	Endorsements     []Endorsement     `json:"endorsements"`
	BallotMeasures   []BallotMeasure   `json:"ballotMeasures"`
	MeasurePositions []MeasurePosition `json:"measurePositions"`
}

func NormalizeDirectoryData(data DirectoryData) DirectoryData {
	if data.Chapters == nil {
		data.Chapters = make([]Chapter, 0)
	}
	if data.Elections == nil {
		data.Elections = make([]Election, 0)
	}
	if data.Races == nil {
		data.Races = make([]Race, 0)
	}
	if data.Candidates == nil {
		data.Candidates = make([]Candidate, 0)
	}
	if data.VoterGuides == nil {
		data.VoterGuides = make([]VoterGuide, 0)
	}
	if data.Endorsements == nil {
		data.Endorsements = make([]Endorsement, 0)
	}
	if data.BallotMeasures == nil {
		data.BallotMeasures = make([]BallotMeasure, 0)
	}
	if data.MeasurePositions == nil {
		data.MeasurePositions = make([]MeasurePosition, 0)
	}

	return data
}
