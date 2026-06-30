import { useState, type FormEvent, type InputHTMLAttributes, type ReactNode } from 'react';
import {
  CivicLookupApiError,
  civicRegistrationDisclaimer,
  fetchCivicElections,
  lookupCivicInformation,
  type CivicAddress,
  type CivicElection,
  type CivicLookupResponse,
  type CivicVotingLocation,
} from '../data/civicApi';

type Notice = {
  kind: 'info' | 'success' | 'error';
  title: string;
  messages?: string[];
};

function NoticeBanner({ notice }: { notice?: Notice }) {
  if (!notice) {
    return null;
  }

  const toneClasses =
    notice.kind === 'success'
      ? 'border-primary bg-surface-muted'
      : notice.kind === 'info'
        ? 'border-border bg-surface-muted'
        : 'border-primary bg-surface';

  return (
    <section className={`rounded-none border-2 p-4 shadow-[4px_4px_0_var(--theme-shadow)] ${toneClasses}`} aria-live="polite">
      <p className="text-sm font-black uppercase tracking-[0.24em] text-primary">{notice.title}</p>
      {notice.messages?.length ? (
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm leading-6 text-foreground">
          {notice.messages.map((message) => (
            <li key={message}>{message}</li>
          ))}
        </ul>
      ) : null}
    </section>
  );
}

function FieldShell({
  children,
  label,
  htmlFor,
  hint,
}: {
  children: ReactNode;
  label: string;
  htmlFor: string;
  hint?: string;
}) {
  return (
    <label className="flex flex-col gap-2 text-sm font-bold text-foreground" htmlFor={htmlFor}>
      <span className="text-xs font-black uppercase tracking-[0.24em] text-primary">{label}</span>
      {children}
      {hint ? <span className="text-xs font-normal leading-5 text-muted">{hint}</span> : null}
    </label>
  );
}

function TextInput(props: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      className={`rounded-none border-2 border-border-strong bg-surface px-3 py-3 text-foreground placeholder:text-muted focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background ${props.className ?? ''}`}
    />
  );
}

function SectionShell({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return (
    <section className="rounded-none border-2 border-border-strong bg-surface p-6 shadow-[8px_8px_0_var(--theme-shadow)] md:p-8">
      <div className="mb-5">
        <p className="text-sm font-black uppercase tracking-[0.28em] text-primary">Public lookup</p>
        <h2 className="mt-3 text-3xl font-black text-foreground">{title}</h2>
        <p className="mt-3 text-sm leading-6 text-muted">{description}</p>
      </div>
      {children}
    </section>
  );
}

function formatAddressLines(address: CivicAddress): string[] {
  return [
    address.locationName,
    address.line1,
    address.line2,
    address.line3,
    [address.city, address.state, address.zip].filter(Boolean).join(', '),
  ].filter((line): line is string => Boolean(line && line.trim()));
}

function formatElectionLabel(election: { name?: string; electionDay?: string; id?: string }): string {
  return [election.name, election.electionDay, election.id ? `ID ${election.id}` : undefined].filter(Boolean).join(' · ');
}

function formatLocationDates(location: CivicVotingLocation): string | undefined {
  if (location.startDate && location.endDate) {
    return `${location.startDate} to ${location.endDate}`;
  }

  return location.startDate ?? location.endDate;
}

function LocationCard({ title, location }: { title: string; location: CivicVotingLocation }) {
  const locationLines = formatAddressLines(location.address);
  const dateRange = formatLocationDates(location);
  const officialSources = location.sources.filter((source) => source.official).map((source) => source.name).filter(Boolean);

  return (
    <article className="rounded-none border-2 border-border bg-surface-muted p-4 shadow-[4px_4px_0_var(--theme-shadow)]">
      <p className="text-xs font-black uppercase tracking-[0.24em] text-primary">{title}</p>
      <address className="mt-3 not-italic text-sm leading-6 text-foreground">
        {locationLines.length > 0 ? locationLines.map((line) => <div key={line}>{line}</div>) : <div>No address returned.</div>}
      </address>
      {location.pollingHours ? <p className="mt-3 text-sm leading-6 text-muted">Hours: {location.pollingHours}</p> : null}
      {dateRange ? <p className="mt-2 text-sm leading-6 text-muted">Dates: {dateRange}</p> : null}
      {officialSources.length > 0 ? <p className="mt-2 text-xs font-bold uppercase tracking-wide text-muted">Source: {officialSources.join(', ')}</p> : null}
    </article>
  );
}

function LocationList({ title, locations, emptyMessage }: { title: string; locations: CivicVotingLocation[]; emptyMessage: string }) {
  return (
    <section className="rounded-none border-2 border-border bg-surface p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
      <h3 className="text-lg font-black text-foreground">
        {title} <span className="text-muted">({locations.length})</span>
      </h3>
      {locations.length > 0 ? (
        <div className="mt-4 grid gap-3 md:grid-cols-2">
          {locations.map((location, index) => (
            <LocationCard key={`${title}-${index}`} location={location} title={title.slice(0, -1)} />
          ))}
        </div>
      ) : (
        <p className="mt-3 text-sm leading-6 text-muted">{emptyMessage}</p>
      )}
    </section>
  );
}

function LinkList({ title, links }: { title: string; links: Array<{ href?: string; label: string }> }) {
  const visibleLinks = links.filter((link): link is { href: string; label: string } => Boolean(link.href));

  return (
    <section className="rounded-none border-2 border-border bg-surface p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
      <h3 className="text-lg font-black text-foreground">{title}</h3>
      {visibleLinks.length > 0 ? (
        <ul className="mt-4 space-y-2 text-sm font-bold">
          {visibleLinks.map((link) => (
            <li key={link.label}>
              <a
                className="text-link underline decoration-2 underline-offset-4 hover:text-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                href={link.href}
              >
                {link.label}
              </a>
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-3 text-sm leading-6 text-muted">No links returned.</p>
      )}
    </section>
  );
}

function ElectionBlock({ title, election }: { title: string; election?: { name?: string; electionDay?: string; id?: string } }) {
  return (
    <section className="rounded-none border-2 border-border bg-surface p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
      <h3 className="text-lg font-black text-foreground">{title}</h3>
      {election && formatElectionLabel(election) ? (
        <p className="mt-3 text-sm leading-6 text-muted">{formatElectionLabel(election)}</p>
      ) : (
        <p className="mt-3 text-sm leading-6 text-muted">No election returned.</p>
      )}
    </section>
  );
}

function ResultSummary({ response }: { response: CivicLookupResponse }) {
  const normalizedInput = response.normalizedInput;
  const addressLines = formatAddressLines(normalizedInput);

  return (
    <section className="grid gap-4 md:grid-cols-2">
      <section className="rounded-none border-2 border-border bg-surface p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
        <h3 className="text-lg font-black text-foreground">Normalized input</h3>
        <div className="mt-3 space-y-1 text-sm leading-6 text-muted">
          {addressLines.length > 0 ? addressLines.map((line) => <div key={line}>{line}</div>) : <div>No normalized input returned.</div>}
        </div>
      </section>
      <ElectionBlock title="Election" election={response.election} />
    </section>
  );
}

function ResponseDetails({ response }: { response: CivicLookupResponse }) {
  return (
    <div className="mt-5 space-y-4">
      {response.otherElections.length > 0 ? (
        <section className="rounded-none border-2 border-border bg-surface p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
          <h3 className="text-lg font-black text-foreground">Other elections</h3>
          <ul className="mt-4 space-y-2 text-sm leading-6 text-muted">
            {response.otherElections.map((election) => (
              <li key={`${election.id ?? election.name ?? election.electionDay}`}>
                {formatElectionLabel(election) || 'Unnamed election'}
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      <ResultSummary response={response} />

      <div className="grid gap-4 lg:grid-cols-2">
        <LocationList emptyMessage="No polling locations were returned for this lookup." locations={response.pollingLocations} title="Polling locations" />
        <LocationList emptyMessage="No early vote sites were returned for this lookup." locations={response.earlyVoteSites} title="Early vote sites" />
        <LocationList emptyMessage="No drop-off locations were returned for this lookup." locations={response.dropOffLocations} title="Drop-off locations" />

        <section className="rounded-none border-2 border-border bg-surface p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
          <h3 className="text-lg font-black text-foreground">State election office</h3>
          <p className="mt-3 text-sm leading-6 text-muted">{response.stateElectionOffice.name || 'State election office links'}</p>
          <div className="mt-4 space-y-2 text-sm font-bold">
            {response.stateElectionOffice.electionInfoUrl ? (
              <a
                className="block text-link underline decoration-2 underline-offset-4 hover:text-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                href={response.stateElectionOffice.electionInfoUrl}
              >
                Election info
              </a>
            ) : null}
            {response.stateElectionOffice.registrationUrl ? (
              <a
                className="block text-link underline decoration-2 underline-offset-4 hover:text-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                href={response.stateElectionOffice.registrationUrl}
              >
                Registration status / vote registration
              </a>
            ) : null}
            {response.stateElectionOffice.registrationConfirmationUrl ? (
              <a
                className="block text-link underline decoration-2 underline-offset-4 hover:text-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                href={response.stateElectionOffice.registrationConfirmationUrl}
              >
                Registration confirmation
              </a>
            ) : null}
            {response.stateElectionOffice.absenteeVotingInfoUrl ? (
              <a
                className="block text-link underline decoration-2 underline-offset-4 hover:text-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                href={response.stateElectionOffice.absenteeVotingInfoUrl}
              >
                Absentee voting info
              </a>
            ) : null}
            {response.stateElectionOffice.ballotInfoUrl ? (
              <a
                className="block text-link underline decoration-2 underline-offset-4 hover:text-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                href={response.stateElectionOffice.ballotInfoUrl}
              >
                Ballot info
              </a>
            ) : null}
            {!response.stateElectionOffice.electionInfoUrl &&
            !response.stateElectionOffice.registrationUrl &&
            !response.stateElectionOffice.registrationConfirmationUrl &&
            !response.stateElectionOffice.absenteeVotingInfoUrl &&
            !response.stateElectionOffice.ballotInfoUrl ? (
              <p className="text-sm leading-6 text-muted">No state office links were returned.</p>
            ) : null}
          </div>
        </section>

        <LinkList
          links={[
            { href: response.fallbackLinks.registerToVoteUrl, label: 'Register to vote' },
            { href: response.fallbackLinks.nassRegistrationStatusUrl, label: 'NASS voter registration status' },
            { href: response.fallbackLinks.nassPollingPlaceUrl, label: 'NASS polling place lookup' },
          ]}
          title="Fallback links"
        />
      </div>
    </div>
  );
}

export function CivicLookupPanel() {
  const [address, setAddress] = useState('');
  const [electionId, setElectionId] = useState('');
  const [elections, setElections] = useState<CivicElection[]>([]);
  const [electionsLoaded, setElectionsLoaded] = useState(false);
  const [response, setResponse] = useState<CivicLookupResponse | undefined>(undefined);
  const [notice, setNotice] = useState<Notice>({
    kind: 'info',
    title: 'Anonymous lookup ready',
    messages: [
      'Address is used for this lookup only and is not stored by DSA Slate.',
      civicRegistrationDisclaimer,
      'Load supported elections, then choose one before looking up an address when possible.',
    ],
  });

  async function loadElections() {
    setNotice({ kind: 'info', title: 'Loading supported Civic elections…' });

    try {
      const nextElections = await fetchCivicElections();
      setElections(nextElections.elections);
      setElectionsLoaded(true);
      setNotice({
        kind: 'success',
        title: nextElections.elections.length > 0 ? 'Supported elections loaded' : 'No supported elections returned',
        messages:
          nextElections.elections.length > 0
            ? ['Choose an election from Google Civic before looking up an address when possible.']
            : ['Google Civic did not return any currently supported elections. You can still try an address lookup without selecting one.'],
      });
    } catch (error) {
      if (error instanceof CivicLookupApiError) {
        setNotice({
          kind: 'error',
          title: error.code === 'civic_not_configured' ? 'Civic lookup not configured' : 'Election list failed',
          messages: error.messages.length > 0 ? error.messages : ['Unable to load supported Civic elections.'],
        });
        return;
      }

      setNotice({ kind: 'error', title: 'Election list failed', messages: ['Unable to load supported Civic elections.'] });
    }
  }

  async function runLookup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setNotice({ kind: 'info', title: 'Looking up official Civic data…' });
    setResponse(undefined);

    try {
      const nextResponse = await lookupCivicInformation({ address, electionId });
      setResponse(nextResponse);
      setNotice({
        kind: 'success',
        title: 'Lookup complete',
        messages: [nextResponse.source, nextResponse.disclaimer],
      });
    } catch (error) {
      if (error instanceof CivicLookupApiError) {
        const errorMessages =
          error.messages.length > 0
            ? error.messages
            : error.code === 'civic_not_configured'
              ? ['Set GOOGLE_CIVIC_API_KEY on the backend to enable this public lookup.']
              : error.code === 'validation_failed'
                ? ['Enter a street address before looking up official election information.']
                : ['The Civic lookup request failed. Please try again.'];

        setNotice({
          kind: 'error',
          title:
            error.code === 'civic_not_configured'
              ? 'Civic lookup not configured'
              : error.code === 'validation_failed'
                ? 'Address required'
                : 'Lookup failed',
          messages: errorMessages,
        });
        return;
      }

      setNotice({
        kind: 'error',
        title: 'Lookup failed',
        messages: [error instanceof Error ? error.message : 'Unknown Civic lookup error.'],
      });
    }
  }

  return (
    <SectionShell
      description="Look up official polling locations and election-office links without storing the submitted address."
      title="Find official election info"
    >
      <div className="space-y-4">
        <div className="rounded-none border-2 border-border bg-surface-muted p-4 text-sm leading-6 text-muted">
          <p className="font-bold text-foreground">Privacy</p>
          <p className="mt-1">Address is used for this lookup only and is not stored by DSA Slate.</p>
          <p className="mt-2 font-bold text-foreground">Disclaimer</p>
          <p className="mt-1">This lookup does not confirm that you are registered to vote.</p>
        </div>

        <form className="grid gap-4 rounded-none border-2 border-border bg-surface-muted p-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] md:items-end" onSubmit={runLookup}>
          <FieldShell htmlFor="civic-address" label="Address" hint="Street address only; the backend does the official lookup.">
            <TextInput
              id="civic-address"
              autoComplete="street-address"
              onChange={(event) => setAddress(event.target.value)}
              placeholder="123 Main St, San Diego, CA"
              type="text"
              value={address}
            />
          </FieldShell>

          <div className="flex flex-col gap-2 text-sm font-bold text-foreground">
            <span className="text-xs font-black uppercase tracking-[0.24em] text-primary">Civic election</span>
            <div className="grid gap-2 md:grid-cols-[1fr_auto]">
              <select
                id="civic-election-id"
                className="rounded-none border-2 border-border-strong bg-surface px-3 py-3 text-foreground focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                value={electionId}
                onChange={(event) => setElectionId(event.target.value)}
              >
                <option value="">Let Civic choose</option>
                {elections.map((election) => (
                  <option key={election.id ?? `${election.name}-${election.electionDay}`} value={election.id ?? ''}>
                    {[election.name, election.electionDay, election.id ? `ID ${election.id}` : undefined].filter(Boolean).join(' · ')}
                  </option>
                ))}
              </select>
              <button
                className="rounded-none border-2 border-border-strong bg-surface px-4 py-3 font-black text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                type="button"
                onClick={loadElections}
              >
                Load supported elections
              </button>
            </div>
            <span className="text-xs font-normal leading-5 text-muted">
              Choose an election from Google Civic before looking up an address when possible.
              {electionsLoaded ? ` ${elections.length} supported elections loaded.` : ''}
            </span>
          </div>

          <button
            className="rounded-none border-2 border-primary bg-primary px-6 py-3 font-black text-primary-foreground transition hover:bg-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            type="submit"
          >
            Lookup
          </button>
        </form>

        <NoticeBanner notice={notice} />

        {response ? (
          <div className="rounded-none border-2 border-border bg-surface-muted p-4 shadow-[4px_4px_0_var(--theme-shadow)]">
            <p className="text-sm font-black uppercase tracking-[0.24em] text-primary">{response.source}</p>
            <p className="mt-2 text-sm leading-6 text-muted">{response.disclaimer}</p>
            <ResponseDetails response={response} />
          </div>
        ) : null}
      </div>
    </SectionShell>
  );
}
