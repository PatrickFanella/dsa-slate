package civic

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestGoogleClientLookup(t *testing.T) {
	const rawAddress = "123 Main St, San Diego, CA"
	const responseBody = `{
  "normalizedInput": {"line1": "123 MAIN ST", "city": "SAN DIEGO", "state": "CA", "zip": "92101"},
  "election": {"id": "2000", "name": "VIP Test Election", "electionDay": "2026-06-02"},
  "otherElections": [{"id": "2001", "name": "Other Election", "electionDay": "2026-11-03"}],
  "pollingLocations": [{
    "address": {"line1": "100 POLLING PL", "city": "SAN DIEGO", "state": "CA", "zip": "92101"},
    "pollingHours": "7am-8pm",
    "sources": [{"name": "County Registrar", "official": true}]
  }],
  "earlyVoteSites": [{
    "address": {"line1": "200 EARLY AVE", "city": "SAN DIEGO", "state": "CA", "zip": "92101"},
    "startDate": "2026-05-25",
    "endDate": "2026-06-01",
    "sources": [{"name": "County Registrar", "official": true}]
  }],
  "dropOffLocations": [],
  "contests": [{"id": "contest-1", "office": "Mayor", "candidates": [{"name": "Ada"}, {"name": "Bea"}]}],
  "state": [{
    "name": "California",
    "electionAdministrationBody": {
      "name": "California Secretary of State",
      "electionInfoUrl": "https://www.sos.ca.gov/elections/",
      "registrationUrl": "https://registertovote.ca.gov/",
      "electionRegistrationConfirmationUrl": "https://voterstatus.sos.ca.gov/",
      "absenteeVotingInfoUrl": "https://example.com/absentee",
      "ballotInfoUrl": "https://example.com/ballot"
    }
  }]
}`

	var gotRequest *http.Request
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotRequest = r.Clone(r.Context())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})
	client := NewGoogleClient("fake-key", &http.Client{Transport: transport})

	got, err := client.Lookup(context.Background(), LookupRequest{Address: rawAddress, ElectionID: "2000"})
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}

	if gotRequest == nil {
		t.Fatal("request was not captured")
	}
	if gotRequest.URL.Path != "/civicinfo/v2/voterinfo" {
		t.Fatalf("path = %q, want /civicinfo/v2/voterinfo", gotRequest.URL.Path)
	}
	query := gotRequest.URL.Query()
	if query.Get("address") != rawAddress {
		t.Fatalf("address query = %q, want %q", query.Get("address"), rawAddress)
	}
	if query.Get("key") != "fake-key" {
		t.Fatalf("key query = %q, want fake-key", query.Get("key"))
	}
	if query.Get("officialOnly") != "true" {
		t.Fatalf("officialOnly query = %q, want true", query.Get("officialOnly"))
	}
	if query.Get("electionId") != "2000" {
		t.Fatalf("electionId query = %q, want 2000", query.Get("electionId"))
	}

	if !strings.Contains(got.Disclaimer, "does not confirm") {
		t.Fatalf("disclaimer = %q, want registration disclaimer", got.Disclaimer)
	}
	if got.Status != "success" || got.Source != SourceGoogleCivicInformationAPI {
		t.Fatalf("unexpected status/source: %#v", got)
	}
	if got.NormalizedInput.Line1 != "123 MAIN ST" || got.NormalizedInput.City != "SAN DIEGO" {
		t.Fatalf("normalized input = %#v, want mapped Google response", got.NormalizedInput)
	}
	if got.NormalizedInput.Line1 == rawAddress {
		t.Fatal("raw request address leaked into normalized input")
	}
	if got.Election.ID != "2000" || got.Election.Name != "VIP Test Election" {
		t.Fatalf("election = %#v, want mapped election", got.Election)
	}
	if len(got.PollingLocations) != 1 || got.PollingLocations[0].Address.Line1 != "100 POLLING PL" {
		t.Fatalf("polling locations = %#v, want mapped location", got.PollingLocations)
	}
	if len(got.EarlyVoteSites) != 1 || got.EarlyVoteSites[0].StartDate != "2026-05-25" {
		t.Fatalf("early vote sites = %#v, want mapped site", got.EarlyVoteSites)
	}
	if got.StateElectionOffice.Name != "California Secretary of State" {
		t.Fatalf("state office = %#v, want mapped office", got.StateElectionOffice)
	}
	if got.StateElectionOffice.RegistrationConfirmationURL != "https://voterstatus.sos.ca.gov/" {
		t.Fatalf("state office confirmation url = %q", got.StateElectionOffice.RegistrationConfirmationURL)
	}
	if got.FallbackLinks.RegisterToVoteURL != "https://vote.gov/" {
		t.Fatalf("fallback links = %#v, want official fallbacks", got.FallbackLinks)
	}
}

func TestGoogleClientLookupRejectsBlankAddress(t *testing.T) {
	client := NewGoogleClient("fake-key", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("unexpected external call")
		return nil, nil
	})})

	_, err := client.Lookup(context.Background(), LookupRequest{Address: "   "})
	if err == nil || err.Error() != "address is required" {
		t.Fatalf("err = %v, want address validation error", err)
	}
}
