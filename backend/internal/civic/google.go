package civic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const googleCivicLookupURL = "https://www.googleapis.com/civicinfo/v2/voterinfo"

// HTTPClient is the minimal client interface used by GoogleClient.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// GoogleClient calls Google Civic Information API voter lookup.
type GoogleClient struct {
	apiKey     string
	httpClient HTTPClient
}

// NewGoogleClient constructs a Google Civic lookup client.
func NewGoogleClient(apiKey string, httpClient HTTPClient) *GoogleClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &GoogleClient{apiKey: strings.TrimSpace(apiKey), httpClient: httpClient}
}

// Lookup performs a Civic voter-info lookup and maps the response into DSA Slate DTOs.
func (c *GoogleClient) Lookup(ctx context.Context, request LookupRequest) (LookupResponse, error) {
	address := strings.TrimSpace(request.Address)
	if address == "" {
		return LookupResponse{}, errors.New("address is required")
	}
	if c.apiKey == "" {
		return LookupResponse{}, errors.New("google civic api key is required")
	}

	lookupURL, err := url.Parse(googleCivicLookupURL)
	if err != nil {
		return LookupResponse{}, fmt.Errorf("parse google civic url: %w", err)
	}
	query := lookupURL.Query()
	query.Set("address", address)
	query.Set("key", c.apiKey)
	query.Set("officialOnly", "true")
	if electionID := strings.TrimSpace(request.ElectionID); electionID != "" {
		query.Set("electionId", electionID)
	}
	lookupURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lookupURL.String(), nil)
	if err != nil {
		return LookupResponse{}, fmt.Errorf("build google civic request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return LookupResponse{}, fmt.Errorf("do google civic request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return LookupResponse{}, fmt.Errorf("google civic lookup returned %s", resp.Status)
	}

	var payload googleLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return LookupResponse{}, fmt.Errorf("decode google civic response: %w", err)
	}

	return mapGoogleLookupResponse(payload), nil
}

type googleLookupResponse struct {
	NormalizedInput  googleAddress    `json:"normalizedInput"`
	Election         googleElection   `json:"election"`
	OtherElections   []googleElection `json:"otherElections"`
	PollingLocations []googleLocation `json:"pollingLocations"`
	EarlyVoteSites   []googleLocation `json:"earlyVoteSites"`
	DropOffLocations []googleLocation `json:"dropOffLocations"`
	Contests         []googleContest  `json:"contests"`
	State            []googleState    `json:"state"`
}

type googleAddress struct {
	LocationName string `json:"locationName"`
	Line1        string `json:"line1"`
	Line2        string `json:"line2"`
	Line3        string `json:"line3"`
	City         string `json:"city"`
	State        string `json:"state"`
	Zip          string `json:"zip"`
}

type googleElection struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ElectionDay string `json:"electionDay"`
}

type googleSource struct {
	Name     string `json:"name"`
	Official bool   `json:"official"`
}

type googleLocation struct {
	Address      googleAddress  `json:"address"`
	PollingHours string         `json:"pollingHours"`
	StartDate    string         `json:"startDate"`
	EndDate      string         `json:"endDate"`
	Sources      []googleSource `json:"sources"`
}

type googleCandidate struct {
	Name string `json:"name"`
}

type googleContest struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Office      string            `json:"office"`
	BallotTitle string            `json:"ballotTitle"`
	District    string            `json:"district"`
	Candidates  []googleCandidate `json:"candidates"`
}

type googleState struct {
	Name                       string               `json:"name"`
	ElectionAdministrationBody googleElectionOffice `json:"electionAdministrationBody"`
}

type googleElectionOffice struct {
	Name                                string `json:"name"`
	ElectionInfoURL                     string `json:"electionInfoUrl"`
	RegistrationURL                     string `json:"registrationUrl"`
	ElectionRegistrationConfirmationURL string `json:"electionRegistrationConfirmationUrl"`
	AbsenteeVotingInfoURL               string `json:"absenteeVotingInfoUrl"`
	BallotInfoURL                       string `json:"ballotInfoUrl"`
}

func mapGoogleLookupResponse(payload googleLookupResponse) LookupResponse {
	response := LookupResponse{
		Status:           "success",
		Source:           SourceGoogleCivicInformationAPI,
		Disclaimer:       RegistrationDisclaimer,
		NormalizedInput:  mapAddress(payload.NormalizedInput),
		Election:         mapElection(payload.Election),
		OtherElections:   mapElections(payload.OtherElections),
		PollingLocations: mapLocations(payload.PollingLocations),
		EarlyVoteSites:   mapLocations(payload.EarlyVoteSites),
		DropOffLocations: mapLocations(payload.DropOffLocations),
		Contests:         mapContests(payload.Contests),
		FallbackLinks: FallbackLinks{
			RegisterToVoteURL:         "https://vote.gov/",
			NASSRegistrationStatusURL: "https://www.nass.org/can-I-vote/voter-registration-status",
			NASSPollingPlaceURL:       "https://www.nass.org/can-i-vote/find-your-polling-place",
		},
	}

	if office := mapStateElectionOffice(payload.State); office != (StateElectionOffice{}) {
		response.StateElectionOffice = office
	}

	return normalizeLookupResponse(response)
}

func mapAddress(input googleAddress) Address {
	return Address{
		LocationName: strings.TrimSpace(input.LocationName),
		Line1:        strings.TrimSpace(input.Line1),
		Line2:        strings.TrimSpace(input.Line2),
		Line3:        strings.TrimSpace(input.Line3),
		City:         strings.TrimSpace(input.City),
		State:        strings.TrimSpace(input.State),
		Zip:          strings.TrimSpace(input.Zip),
	}
}

func mapElection(input googleElection) Election {
	return Election{
		ID:          strings.TrimSpace(input.ID),
		Name:        strings.TrimSpace(input.Name),
		ElectionDay: strings.TrimSpace(input.ElectionDay),
	}
}

func mapElections(input []googleElection) []Election {
	items := make([]Election, 0, len(input))
	for _, election := range input {
		items = append(items, mapElection(election))
	}
	return items
}

func mapLocation(input googleLocation) VotingLocation {
	return VotingLocation{
		Address:      mapAddress(input.Address),
		PollingHours: strings.TrimSpace(input.PollingHours),
		StartDate:    strings.TrimSpace(input.StartDate),
		EndDate:      strings.TrimSpace(input.EndDate),
		Sources:      mapSources(input.Sources),
	}
}

func mapLocations(input []googleLocation) []VotingLocation {
	items := make([]VotingLocation, 0, len(input))
	for _, location := range input {
		items = append(items, mapLocation(location))
	}
	return items
}

func mapSources(input []googleSource) []Source {
	items := make([]Source, 0, len(input))
	for _, source := range input {
		items = append(items, Source{Name: strings.TrimSpace(source.Name), Official: source.Official})
	}
	return items
}

func mapContests(input []googleContest) []Contest {
	items := make([]Contest, 0, len(input))
	for _, contest := range input {
		candidates := make([]string, 0, len(contest.Candidates))
		for _, candidate := range contest.Candidates {
			candidates = append(candidates, strings.TrimSpace(candidate.Name))
		}
		items = append(items, Contest{
			ID:          strings.TrimSpace(contest.ID),
			Type:        strings.TrimSpace(contest.Type),
			Office:      strings.TrimSpace(contest.Office),
			BallotTitle: strings.TrimSpace(contest.BallotTitle),
			District:    strings.TrimSpace(contest.District),
			Candidates:  candidates,
		})
	}
	return items
}

func mapStateElectionOffice(states []googleState) StateElectionOffice {
	for _, state := range states {
		office := state.ElectionAdministrationBody
		mapped := StateElectionOffice{
			Name:                        strings.TrimSpace(office.Name),
			ElectionInfoURL:             strings.TrimSpace(office.ElectionInfoURL),
			RegistrationURL:             strings.TrimSpace(office.RegistrationURL),
			RegistrationConfirmationURL: strings.TrimSpace(office.ElectionRegistrationConfirmationURL),
			AbsenteeVotingInfoURL:       strings.TrimSpace(office.AbsenteeVotingInfoURL),
			BallotInfoURL:               strings.TrimSpace(office.BallotInfoURL),
		}
		if mapped != (StateElectionOffice{}) {
			return mapped
		}
	}
	return StateElectionOffice{}
}

func normalizeLookupResponse(response LookupResponse) LookupResponse {
	if response.OtherElections == nil {
		response.OtherElections = make([]Election, 0)
	}
	if response.PollingLocations == nil {
		response.PollingLocations = make([]VotingLocation, 0)
	}
	if response.EarlyVoteSites == nil {
		response.EarlyVoteSites = make([]VotingLocation, 0)
	}
	if response.DropOffLocations == nil {
		response.DropOffLocations = make([]VotingLocation, 0)
	}
	if response.Contests == nil {
		response.Contests = make([]Contest, 0)
	}
	return response
}
