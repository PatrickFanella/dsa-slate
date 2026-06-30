package civic

import "context"

const (
	// SourceGoogleCivicInformationAPI identifies Google Civic as the upstream data source.
	SourceGoogleCivicInformationAPI = "google_civic_information_api"

	// RegistrationDisclaimer is returned with every lookup so the API does not imply it can confirm registration.
	RegistrationDisclaimer = "This lookup does not confirm that you are registered to vote. Use your state election office registration-status link to confirm."
)

// LookupRequest is the anonymous Civic lookup request body.
type LookupRequest struct {
	Address    string `json:"address"`
	ElectionID string `json:"electionId,omitempty"`
}

// LookupResponse is the privacy-safe lookup response returned to clients.
type LookupResponse struct {
	Status              string              `json:"status"`
	Source              string              `json:"source"`
	Disclaimer          string              `json:"disclaimer"`
	NormalizedInput     Address             `json:"normalizedInput"`
	Election            Election            `json:"election"`
	OtherElections      []Election          `json:"otherElections"`
	PollingLocations    []VotingLocation    `json:"pollingLocations"`
	EarlyVoteSites      []VotingLocation    `json:"earlyVoteSites"`
	DropOffLocations    []VotingLocation    `json:"dropOffLocations"`
	Contests            []Contest           `json:"contests"`
	StateElectionOffice StateElectionOffice `json:"stateElectionOffice"`
	FallbackLinks       FallbackLinks       `json:"fallbackLinks"`
}

// ElectionsResponse lists Google Civic-supported elections.
type ElectionsResponse struct {
	Source    string     `json:"source"`
	Elections []Election `json:"elections"`
}

// Address is a normalized civic address block.
type Address struct {
	LocationName string `json:"locationName,omitempty"`
	Line1        string `json:"line1,omitempty"`
	Line2        string `json:"line2,omitempty"`
	Line3        string `json:"line3,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	Zip          string `json:"zip,omitempty"`
}

// Election identifies a Google Civic election.
type Election struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	ElectionDay string `json:"electionDay,omitempty"`
}

// VotingLocation identifies a polling, early-vote, or drop-off site.
type VotingLocation struct {
	Address      Address  `json:"address"`
	PollingHours string   `json:"pollingHours,omitempty"`
	StartDate    string   `json:"startDate,omitempty"`
	EndDate      string   `json:"endDate,omitempty"`
	Sources      []Source `json:"sources"`
}

// Source identifies a location source.
type Source struct {
	Name     string `json:"name,omitempty"`
	Official bool   `json:"official"`
}

// Contest is a mapped contest from Google Civic.
type Contest struct {
	ID          string   `json:"id,omitempty"`
	Type        string   `json:"type,omitempty"`
	Office      string   `json:"office,omitempty"`
	BallotTitle string   `json:"ballotTitle,omitempty"`
	District    string   `json:"district,omitempty"`
	Candidates  []string `json:"candidates"`
}

// StateElectionOffice holds election-office and registration links.
type StateElectionOffice struct {
	Name                        string `json:"name,omitempty"`
	ElectionInfoURL             string `json:"electionInfoUrl,omitempty"`
	RegistrationURL             string `json:"registrationUrl,omitempty"`
	RegistrationConfirmationURL string `json:"registrationConfirmationUrl,omitempty"`
	AbsenteeVotingInfoURL       string `json:"absenteeVotingInfoUrl,omitempty"`
	BallotInfoURL               string `json:"ballotInfoUrl,omitempty"`
}

// FallbackLinks are official fallback resources kept outside the Google response.
type FallbackLinks struct {
	RegisterToVoteURL         string `json:"registerToVoteUrl"`
	NASSRegistrationStatusURL string `json:"nassRegistrationStatusUrl"`
	NASSPollingPlaceURL       string `json:"nassPollingPlaceUrl"`
}

// LookupService resolves anonymous Civic lookup requests.
type LookupService interface {
	Lookup(ctx context.Context, request LookupRequest) (LookupResponse, error)
	Elections(ctx context.Context) (ElectionsResponse, error)
}
