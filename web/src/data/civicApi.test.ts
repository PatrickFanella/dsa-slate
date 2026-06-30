import { describe, expect, it, vi } from 'vitest';
import { CivicLookupApiError, lookupCivicInformation } from './civicApi';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('lookupCivicInformation', () => {
  it('posts a trimmed address and omits a blank election id', async () => {
    const fetcher = vi.fn(async () =>
      jsonResponse({
        status: 'success',
        source: 'google_civic_information_api',
        disclaimer: 'This lookup does not confirm that you are registered to vote.',
        normalizedInput: { city: 'San Diego', state: 'CA', zip: '92101' },
        election: { id: '2000', name: 'VIP Test Election', electionDay: '2026-06-02' },
        otherElections: [],
        pollingLocations: [],
        earlyVoteSites: [],
        dropOffLocations: [],
        contests: [],
        stateElectionOffice: { name: 'California Secretary of State' },
        fallbackLinks: {
          registerToVoteUrl: 'https://vote.gov/',
          nassRegistrationStatusUrl: 'https://www.nass.org/can-I-vote/voter-registration-status',
          nassPollingPlaceUrl: 'https://www.nass.org/can-i-vote/find-your-polling-place',
        },
      }),
    );

    const response = await lookupCivicInformation({ address: ' 123 Main St, San Diego, CA ', electionId: '   ' }, fetcher);

    expect(fetcher).toHaveBeenCalledWith('/api/civic/lookup', {
      method: 'POST',
      headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
      body: JSON.stringify({ address: '123 Main St, San Diego, CA' }),
    });
    expect(response.election.id).toBe('2000');
    expect(response.normalizedInput.city).toBe('San Diego');
  });

  it.each([
    { status: 503, body: { error: 'civic_not_configured' }, code: 'civic_not_configured' as const, messages: [] },
    { status: 400, body: { error: 'validation_failed', messages: ['address is required'] }, code: 'validation_failed' as const, messages: ['address is required'] },
    { status: 500, body: { error: 'civic_lookup_failed', message: 'Google Civic lookup failed' }, code: 'civic_lookup_failed' as const, messages: ['Google Civic lookup failed'] },
  ])('throws a typed $code error', async ({ status, body, code, messages }) => {
    const fetcher = vi.fn(async () => jsonResponse(body, status));

    try {
      await lookupCivicInformation({ address: '123 Main St, San Diego, CA' }, fetcher);
      throw new Error('expected lookupCivicInformation to throw');
    } catch (error) {
      expect(error).toBeInstanceOf(CivicLookupApiError);
      const civicError = error as CivicLookupApiError;
      expect(civicError.status).toBe(status);
      expect(civicError.code).toBe(code);
      expect(civicError.messages).toEqual(messages);
    }
  });
});
