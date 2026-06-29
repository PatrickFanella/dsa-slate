package directory

import "context"

type MemoryStore struct {
	Data DirectoryData
}

func NewMemoryStore(data DirectoryData) *MemoryStore {
	return &MemoryStore{Data: NormalizeDirectoryData(data)}
}

func (s *MemoryStore) Directory(ctx context.Context) (DirectoryData, error) {
	return NormalizeDirectoryData(s.Data), nil
}

func (s *MemoryStore) Chapters(ctx context.Context) ([]Chapter, error) {
	return NormalizeDirectoryData(s.Data).Chapters, nil
}

func (s *MemoryStore) Elections(ctx context.Context) ([]Election, error) {
	return NormalizeDirectoryData(s.Data).Elections, nil
}

func (s *MemoryStore) Races(ctx context.Context, filter RaceFilter) ([]Race, error) {
	data := NormalizeDirectoryData(s.Data)
	races := make([]Race, 0)
	for _, race := range data.Races {
		if filter.ElectionID != "" && race.ElectionID != filter.ElectionID {
			continue
		}
		if filter.Jurisdiction != "" && race.Jurisdiction != filter.Jurisdiction {
			continue
		}
		if filter.Level != "" && race.Level != filter.Level {
			continue
		}
		races = append(races, race)
	}
	return races, nil
}

func (s *MemoryStore) Candidates(ctx context.Context) ([]Candidate, error) {
	return NormalizeDirectoryData(s.Data).Candidates, nil
}

func (s *MemoryStore) VoterGuides(ctx context.Context, filter VoterGuideFilter) ([]VoterGuide, error) {
	data := NormalizeDirectoryData(s.Data)
	guides := make([]VoterGuide, 0)
	for _, guide := range data.VoterGuides {
		if filter.ChapterID != "" && guide.ChapterID != filter.ChapterID {
			continue
		}
		if filter.ElectionID != "" && guide.ElectionID != filter.ElectionID {
			continue
		}
		guides = append(guides, guide)
	}
	return guides, nil
}

func (s *MemoryStore) Endorsements(ctx context.Context, filter EndorsementFilter) ([]Endorsement, error) {
	data := NormalizeDirectoryData(s.Data)
	endorsements := make([]Endorsement, 0)
	for _, endorsement := range data.Endorsements {
		if filter.ChapterID != "" && endorsement.ChapterID != filter.ChapterID {
			continue
		}
		if filter.ElectionID != "" {
			matched := false
			for _, race := range data.Races {
				if race.ID == endorsement.RaceID && race.ElectionID == filter.ElectionID {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if filter.RaceID != "" && endorsement.RaceID != filter.RaceID {
			continue
		}
		endorsements = append(endorsements, endorsement)
	}
	return endorsements, nil
}

func (s *MemoryStore) BallotMeasures(ctx context.Context, filter BallotMeasureFilter) ([]BallotMeasure, error) {
	data := NormalizeDirectoryData(s.Data)
	ballotMeasures := make([]BallotMeasure, 0)
	for _, measure := range data.BallotMeasures {
		if filter.ElectionID != "" && measure.ElectionID != filter.ElectionID {
			continue
		}
		ballotMeasures = append(ballotMeasures, measure)
	}
	return ballotMeasures, nil
}

func (s *MemoryStore) MeasurePositions(ctx context.Context, filter MeasurePositionFilter) ([]MeasurePosition, error) {
	data := NormalizeDirectoryData(s.Data)
	positions := make([]MeasurePosition, 0)
	for _, position := range data.MeasurePositions {
		if filter.ChapterID != "" && position.ChapterID != filter.ChapterID {
			continue
		}
		if filter.MeasureID != "" && position.MeasureID != filter.MeasureID {
			continue
		}
		positions = append(positions, position)
	}
	return positions, nil
}
