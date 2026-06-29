package directory

import (
	"context"
	"testing"
)

func TestMemoryStoreRacesFilter(t *testing.T) {
	store := NewMemoryStore(DirectoryData{Races: []Race{
		{ID: "race-1", ElectionID: "election-1", Office: "Mayor", Level: "local", Jurisdiction: "City"},
		{ID: "race-2", ElectionID: "election-2", Office: "Governor", Level: "statewide", Jurisdiction: "California"},
		{ID: "race-3", ElectionID: "election-2", Office: "Controller", Level: "statewide", Jurisdiction: "Los Angeles"},
	}})

	got, err := store.Races(context.Background(), RaceFilter{ElectionID: "election-2", Jurisdiction: "California", Level: "statewide"})
	if err != nil {
		t.Fatalf("Races returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "race-2" {
		t.Fatalf("Races = %#v, want race-2", got)
	}
}

func TestMemoryStoreRacesIgnoreEmptyFields(t *testing.T) {
	store := NewMemoryStore(DirectoryData{Races: []Race{
		{ID: "race-1", ElectionID: "election-1", Office: "Mayor", Level: "local", Jurisdiction: "City"},
		{ID: "race-2", ElectionID: "election-2", Office: "Governor", Level: "statewide", Jurisdiction: "California"},
	}})

	got, err := store.Races(context.Background(), RaceFilter{Level: "statewide"})
	if err != nil {
		t.Fatalf("Races returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "race-2" {
		t.Fatalf("Races = %#v, want race-2", got)
	}
}

func TestMemoryStoreEndorsementsFilter(t *testing.T) {
	store := NewMemoryStore(DirectoryData{
		Races: []Race{
			{ID: "race-1", ElectionID: "election-1", Office: "Mayor", Level: "local", Jurisdiction: "City"},
			{ID: "race-2", ElectionID: "election-2", Office: "Governor", Level: "statewide", Jurisdiction: "California"},
		},
		Endorsements: []Endorsement{
			{ID: "endorsement-1", CandidateID: "candidate-1", RaceID: "race-1", ChapterID: "chapter-1", EndorsementType: "endorsed", SourceURL: "https://example.com/e-1", Confidence: "official_source"},
			{ID: "endorsement-2", CandidateID: "candidate-2", RaceID: "race-2", ChapterID: "chapter-2", EndorsementType: "endorsed", SourceURL: "https://example.com/e-2", Confidence: "official_source"},
		},
	})

	got, err := store.Endorsements(context.Background(), EndorsementFilter{ChapterID: "chapter-2", ElectionID: "election-2", RaceID: "race-2"})
	if err != nil {
		t.Fatalf("Endorsements returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "endorsement-2" {
		t.Fatalf("Endorsements = %#v, want endorsement-2", got)
	}
}

func TestMemoryStoreMeasurePositionsFilter(t *testing.T) {
	store := NewMemoryStore(DirectoryData{MeasurePositions: []MeasurePosition{
		{ID: "pos-1", MeasureID: "measure-1", ChapterID: "chapter-1", Position: "support", SourceURL: "https://example.com/p1", Confidence: "official_source"},
		{ID: "pos-2", MeasureID: "measure-2", ChapterID: "chapter-1", Position: "oppose", SourceURL: "https://example.com/p2", Confidence: "official_source"},
	}})

	got, err := store.MeasurePositions(context.Background(), MeasurePositionFilter{ChapterID: "chapter-1", MeasureID: "measure-2"})
	if err != nil {
		t.Fatalf("MeasurePositions returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "pos-2" {
		t.Fatalf("MeasurePositions = %#v, want pos-2", got)
	}
}
