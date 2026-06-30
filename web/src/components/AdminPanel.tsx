import {
  useMemo,
  useState,
  type FormEvent,
  type InputHTMLAttributes,
  type ReactNode,
  type SelectHTMLAttributes,
  type TextareaHTMLAttributes,
} from 'react';
import {
  AdminApiError,
  adminChapterDraftToForm,
  adminDraftStatuses,
  adminEndorsementDraftToForm,
  archiveAdminChapterDraft,
  archiveAdminEndorsementDraft,
  emptyAdminChapterDraftForm,
  emptyAdminEndorsementDraftForm,
  loadAdminChapterDraft,
  loadAdminEndorsementDraft,
  listAdminChapterDrafts,
  listAdminEndorsementDrafts,
  publishAdminChapterDraft,
  publishAdminEndorsementDraft,
  saveAdminChapterDraft,
  saveAdminEndorsementDraft,
  type AdminChapterDraft,
  type AdminChapterDraftForm,
  type AdminEndorsementDraft,
  type AdminEndorsementDraftForm,
} from '../data/adminApi';
import { confidenceLabel, confidenceLevels, endorsementTypeLabel, endorsementTypes } from '../data/schema';

type Notice = {
  kind: 'info' | 'success' | 'error';
  title: string;
  messages?: string[];
};

function draftStatusLabel(value: string): string {
  return value
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

function errorToNotice(error: unknown): Notice {
  if (error instanceof AdminApiError) {
    const adminError = error;

    if (adminError.error === 'validation_failed') {
      return {
        kind: 'error',
        title: 'Validation failed',
        messages: adminError.messages.length > 0 ? adminError.messages : ['The draft payload was rejected.'],
      };
    }

    return {
      kind: 'error',
      title: `Request failed (${adminError.status})`,
      messages: adminError.messages.length > 0 ? adminError.messages : [adminError.error],
    };
  }

  return {
    kind: 'error',
    title: 'Request failed',
    messages: [error instanceof Error ? error.message : 'Unknown error'],
  };
}

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

function TextArea(props: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      {...props}
      className={`min-h-28 rounded-none border-2 border-border-strong bg-surface px-3 py-3 text-foreground placeholder:text-muted focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background ${props.className ?? ''}`}
    />
  );
}

function Select(props: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      {...props}
      className={`rounded-none border-2 border-border-strong bg-surface px-3 py-3 text-foreground focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background ${props.className ?? ''}`}
    />
  );
}

function DraftListCard({
  draft,
  onArchive,
  onLoad,
  onPublish,
}: {
  draft: AdminChapterDraft | AdminEndorsementDraft;
  onLoad: () => void;
  onPublish: () => void;
  onArchive: () => void;
}) {
  return (
    <li className="rounded-none border-2 border-border bg-surface-muted p-4 shadow-[4px_4px_0_var(--theme-shadow)]">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-sm font-black uppercase tracking-[0.24em] text-primary">{draft.status}</p>
          <h3 className="mt-1 text-lg font-black text-foreground">{draft.id}</h3>
          {'chapterId' in draft ? <p className="mt-2 text-sm text-muted">Chapter: {draft.chapterId}</p> : null}
          {'endorsementId' in draft ? <p className="mt-2 text-sm text-muted">Endorsement: {draft.endorsementId}</p> : null}
          {'reviewNotes' in draft && draft.reviewNotes ? <p className="mt-2 text-sm leading-6 text-muted">{draft.reviewNotes}</p> : null}
        </div>
        <div className="flex flex-wrap gap-2">
          <button
            className="rounded-none border-2 border-border-strong bg-surface px-3 py-2 text-xs font-black uppercase tracking-[0.24em] text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            type="button"
            onClick={onLoad}
          >
            Load
          </button>
          <button
            className="rounded-none border-2 border-primary bg-primary px-3 py-2 text-xs font-black uppercase tracking-[0.24em] text-primary-foreground transition hover:bg-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            type="button"
            onClick={onPublish}
          >
            Publish
          </button>
          <button
            className="rounded-none border-2 border-border-strong bg-surface px-3 py-2 text-xs font-black uppercase tracking-[0.24em] text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            type="button"
            onClick={onArchive}
          >
            Archive
          </button>
        </div>
      </div>
    </li>
  );
}

function SectionShell({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return (
    <section className="rounded-none border-2 border-border-strong bg-surface p-6 shadow-[8px_8px_0_var(--theme-shadow)] md:p-8">
      <div className="mb-5">
        <h2 className="text-2xl font-black text-foreground">{title}</h2>
        <p className="mt-2 text-sm leading-6 text-muted">{description}</p>
      </div>
      {children}
    </section>
  );
}

export function AdminPanel() {
  const [token, setToken] = useState('');
  const [notice, setNotice] = useState<Notice>({
    kind: 'info',
    title: 'Token-protected editorial drafting',
    messages: [
      'Open this page at /admin for local editorial drafting only.',
      'The bearer token is kept in memory, sent as Authorization: Bearer <token>, and not stored in localStorage.',
      'This UI is not a public submission form.',
    ],
  });
  const [chapterDraft, setChapterDraft] = useState<AdminChapterDraftForm>(() => emptyAdminChapterDraftForm());
  const [endorsementDraft, setEndorsementDraft] = useState<AdminEndorsementDraftForm>(() => emptyAdminEndorsementDraftForm());
  const [chapterDrafts, setChapterDrafts] = useState<AdminChapterDraft[]>([]);
  const [endorsementDrafts, setEndorsementDrafts] = useState<AdminEndorsementDraft[]>([]);

  const chapterStatusOptions = useMemo(
    () =>
      adminDraftStatuses.map((status) => (
        <option key={status} value={status}>
          {draftStatusLabel(status)}
        </option>
      )),
    [],
  );

  const endorsementTypeOptions = useMemo(
    () =>
      endorsementTypes.map((type) => (
        <option key={type} value={type}>
          {endorsementTypeLabel(type)}
        </option>
      )),
    [],
  );

  const confidenceOptions = useMemo(
    () =>
      confidenceLevels.map((level) => (
        <option key={level} value={level}>
          {confidenceLabel(level)}
        </option>
      )),
    [],
  );

  async function runAction(action: () => Promise<void>, successTitle: string) {
    if (!token.trim()) {
      setNotice({ kind: 'error', title: 'Token required', messages: ['Enter the admin bearer token before making requests.'] });
      return;
    }

    try {
      await action();
      setNotice({ kind: 'success', title: successTitle });
    } catch (error) {
      setNotice(errorToNotice(error));
    }
  }

  function resetChapterDraftFromResponse(draft: AdminChapterDraft) {
    setChapterDraft(adminChapterDraftToForm(draft));
  }

  function resetEndorsementDraftFromResponse(draft: AdminEndorsementDraft) {
    setEndorsementDraft(adminEndorsementDraftToForm(draft));
  }

  return (
    <main className="min-h-screen bg-background px-6 py-10 text-foreground">
      <section className="mx-auto flex max-w-6xl flex-col gap-8">
        <header className="rounded-none border-2 border-border-strong bg-surface p-8 shadow-[10px_10px_0_var(--theme-shadow)] md:p-12">
          <p className="text-sm font-black uppercase tracking-[0.28em] text-primary">DSA Slate Admin</p>
          <h1 className="mt-4 font-display text-4xl font-black tracking-tight text-foreground md:text-6xl">Editorial draft console</h1>
          <p className="mt-5 max-w-3xl text-lg leading-8 text-muted">
            Token-protected local drafting for chapter and endorsement records. This page is for editorial work only;
            it does not accept public submissions.
          </p>

          <div className="mt-8 grid gap-4 rounded-none border-2 border-border bg-surface-muted p-4 md:grid-cols-[1fr_auto] md:items-end">
            <FieldShell htmlFor="admin-token" label="Admin bearer token" hint="Not persisted. Kept in session memory only.">
              <TextInput
                id="admin-token"
                autoComplete="off"
                onChange={(event) => setToken(event.target.value)}
                placeholder="ADMIN_API_TOKEN"
                type="password"
                value={token}
              />
            </FieldShell>
            <button
              className="rounded-none border-2 border-primary bg-primary px-6 py-3 font-black text-primary-foreground transition hover:bg-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
              type="button"
              onClick={() =>
                setNotice({
                  kind: 'info',
                  title: 'Token copied into memory',
                  messages: ['Requests will send Authorization: Bearer <token> for this session.'],
                })
              }
            >
              Use token
            </button>
          </div>

          <div className="mt-4">
            <NoticeBanner notice={notice} />
          </div>
        </header>

        <section className="grid gap-6 lg:grid-cols-2">
          <SectionShell
            description="Draft chapter records before they are published to the public read tables."
            title="Chapter drafts"
          >
            <form
              className="grid gap-4"
              onSubmit={async (event: FormEvent<HTMLFormElement>) => {
                event.preventDefault();
                await runAction(async () => {
                  if (!chapterDraft.id.trim()) {
                    throw new Error('Enter a chapter draft ID before saving.');
                  }
                  const saved = await saveAdminChapterDraft(token, chapterDraft);
                  setChapterDraft(adminChapterDraftToForm(saved));
                  setChapterDrafts((current) => {
                    const next = current.filter((draft) => draft.id !== saved.id);
                    next.unshift(saved);
                    return next;
                  });
                }, 'Chapter draft saved');
              }}
            >
              <div className="grid gap-4 md:grid-cols-2">
                <FieldShell htmlFor="chapter-draft-id" label="ID">
                  <TextInput
                    id="chapter-draft-id"
                    onChange={(event) => setChapterDraft((current) => ({ ...current, id: event.target.value }))}
                    value={chapterDraft.id}
                  />
                </FieldShell>
                <FieldShell htmlFor="chapter-draft-chapter-id" label="Chapter ID">
                  <TextInput
                    id="chapter-draft-chapter-id"
                    onChange={(event) => setChapterDraft((current) => ({ ...current, chapterId: event.target.value }))}
                    value={chapterDraft.chapterId}
                  />
                </FieldShell>
                <FieldShell htmlFor="chapter-draft-name" label="Name">
                  <TextInput
                    id="chapter-draft-name"
                    onChange={(event) => setChapterDraft((current) => ({ ...current, name: event.target.value }))}
                    value={chapterDraft.name}
                  />
                </FieldShell>
                <FieldShell htmlFor="chapter-draft-state" label="State">
                  <TextInput
                    id="chapter-draft-state"
                    onChange={(event) => setChapterDraft((current) => ({ ...current, state: event.target.value }))}
                    value={chapterDraft.state}
                  />
                </FieldShell>
                <FieldShell htmlFor="chapter-draft-status" label="Draft status">
                  <Select
                    id="chapter-draft-status"
                    onChange={(event) => setChapterDraft((current) => ({ ...current, status: event.target.value as AdminChapterDraftForm['status'] }))}
                    value={chapterDraft.status}
                  >
                    {chapterStatusOptions}
                  </Select>
                </FieldShell>
                <FieldShell htmlFor="chapter-draft-reviewed-at" label="Reviewed at">
                  <TextInput
                    id="chapter-draft-reviewed-at"
                    onChange={(event) => setChapterDraft((current) => ({ ...current, reviewedAt: event.target.value }))}
                    type="date"
                    value={chapterDraft.reviewedAt}
                  />
                </FieldShell>
              </div>

              <FieldShell htmlFor="chapter-draft-website-url" label="Website URL">
                <TextInput
                  id="chapter-draft-website-url"
                  onChange={(event) => setChapterDraft((current) => ({ ...current, websiteUrl: event.target.value }))}
                  placeholder="https://example.org"
                  type="url"
                  value={chapterDraft.websiteUrl}
                />
              </FieldShell>

              <FieldShell htmlFor="chapter-draft-group-email" label="Group email">
                <TextInput
                  id="chapter-draft-group-email"
                  onChange={(event) => setChapterDraft((current) => ({ ...current, groupEmail: event.target.value }))}
                  placeholder="hello@example.org"
                  type="email"
                  value={chapterDraft.groupEmail}
                />
              </FieldShell>

              <FieldShell htmlFor="chapter-draft-review-notes" label="Review notes">
                <TextArea
                  id="chapter-draft-review-notes"
                  onChange={(event) => setChapterDraft((current) => ({ ...current, reviewNotes: event.target.value }))}
                  value={chapterDraft.reviewNotes}
                />
              </FieldShell>

              <div className="flex flex-wrap gap-3">
                <button
                  className="rounded-none border-2 border-primary bg-primary px-5 py-3 font-black text-primary-foreground transition hover:bg-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                  type="submit"
                >
                  Save draft
                </button>
                <button
                  className="rounded-none border-2 border-border-strong bg-surface px-5 py-3 font-black text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                  type="button"
                  onClick={async () =>
                    runAction(async () => {
                      const drafts = await listAdminChapterDrafts(token);
                      setChapterDrafts(drafts);
                    }, 'Chapter drafts loaded')
                  }
                >
                  List drafts
                </button>
                <button
                  className="rounded-none border-2 border-border-strong bg-surface px-5 py-3 font-black text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                  type="button"
                  onClick={async () =>
                    runAction(async () => {
                      if (!chapterDraft.id.trim()) {
                        throw new Error('Enter a chapter draft ID before loading it.');
                      }
                      const draft = await loadAdminChapterDraft(token, chapterDraft.id);
                      resetChapterDraftFromResponse(draft);
                    }, 'Chapter draft loaded')
                  }
                >
                  Load by ID
                </button>
              </div>
            </form>

            {chapterDrafts.length > 0 ? (
              <ul className="mt-6 space-y-3">
                {chapterDrafts.map((draft) => (
                  <DraftListCard
                    draft={draft}
                    key={draft.id}
                    onArchive={async () =>
                      runAction(async () => {
                        await archiveAdminChapterDraft(token, draft.id);
                        const refreshed = await listAdminChapterDrafts(token);
                        setChapterDrafts(refreshed);
                      }, 'Chapter draft archived')
                    }
                    onLoad={() => resetChapterDraftFromResponse(draft)}
                    onPublish={async () =>
                      runAction(async () => {
                        await publishAdminChapterDraft(token, draft.id);
                        const refreshed = await listAdminChapterDrafts(token);
                        setChapterDrafts(refreshed);
                      }, 'Chapter draft published')
                    }
                  />
                ))}
              </ul>
            ) : (
              <p className="mt-6 rounded-none border-2 border-dashed border-border bg-surface-muted p-4 text-sm leading-6 text-muted">
                Load chapter drafts to review, publish, or archive them.
              </p>
            )}
          </SectionShell>

          <SectionShell
            description="Draft endorsement records against the token-protected editorial API."
            title="Endorsement drafts"
          >
            <form
              className="grid gap-4"
              onSubmit={async (event: FormEvent<HTMLFormElement>) => {
                event.preventDefault();
                await runAction(async () => {
                  if (!endorsementDraft.id.trim()) {
                    throw new Error('Enter an endorsement draft ID before saving.');
                  }
                  const saved = await saveAdminEndorsementDraft(token, endorsementDraft);
                  setEndorsementDraft(adminEndorsementDraftToForm(saved));
                  setEndorsementDrafts((current) => {
                    const next = current.filter((draft) => draft.id !== saved.id);
                    next.unshift(saved);
                    return next;
                  });
                }, 'Endorsement draft saved');
              }}
            >
              <div className="grid gap-4 md:grid-cols-2">
                <FieldShell htmlFor="endorsement-draft-id" label="ID">
                  <TextInput
                    id="endorsement-draft-id"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, id: event.target.value }))}
                    value={endorsementDraft.id}
                  />
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-endorsement-id" label="Endorsement ID">
                  <TextInput
                    id="endorsement-draft-endorsement-id"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, endorsementId: event.target.value }))}
                    value={endorsementDraft.endorsementId}
                  />
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-candidate-id" label="Candidate ID">
                  <TextInput
                    id="endorsement-draft-candidate-id"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, candidateId: event.target.value }))}
                    value={endorsementDraft.candidateId}
                  />
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-race-id" label="Race ID">
                  <TextInput
                    id="endorsement-draft-race-id"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, raceId: event.target.value }))}
                    value={endorsementDraft.raceId}
                  />
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-chapter-id" label="Chapter ID">
                  <TextInput
                    id="endorsement-draft-chapter-id"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, chapterId: event.target.value }))}
                    value={endorsementDraft.chapterId}
                  />
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-voter-guide-id" label="Voter guide ID">
                  <TextInput
                    id="endorsement-draft-voter-guide-id"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, voterGuideId: event.target.value }))}
                    value={endorsementDraft.voterGuideId}
                  />
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-endorsement-type" label="Endorsement type">
                  <Select
                    id="endorsement-draft-endorsement-type"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, endorsementType: event.target.value as AdminEndorsementDraftForm['endorsementType'] }))}
                    value={endorsementDraft.endorsementType}
                  >
                    {endorsementTypeOptions}
                  </Select>
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-confidence" label="Confidence">
                  <Select
                    id="endorsement-draft-confidence"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, confidence: event.target.value as AdminEndorsementDraftForm['confidence'] }))}
                    value={endorsementDraft.confidence}
                  >
                    {confidenceOptions}
                  </Select>
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-status" label="Draft status">
                  <Select
                    id="endorsement-draft-status"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, status: event.target.value as AdminEndorsementDraftForm['status'] }))}
                    value={endorsementDraft.status}
                  >
                    {chapterStatusOptions}
                  </Select>
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-source-date" label="Source date">
                  <TextInput
                    id="endorsement-draft-source-date"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, sourceDate: event.target.value }))}
                    type="date"
                    value={endorsementDraft.sourceDate}
                  />
                </FieldShell>
                <FieldShell htmlFor="endorsement-draft-reviewed-at" label="Reviewed at">
                  <TextInput
                    id="endorsement-draft-reviewed-at"
                    onChange={(event) => setEndorsementDraft((current) => ({ ...current, reviewedAt: event.target.value }))}
                    type="date"
                    value={endorsementDraft.reviewedAt}
                  />
                </FieldShell>
              </div>

              <FieldShell htmlFor="endorsement-draft-source-url" label="Source URL">
                <TextInput
                  id="endorsement-draft-source-url"
                  onChange={(event) => setEndorsementDraft((current) => ({ ...current, sourceUrl: event.target.value }))}
                  placeholder="https://example.org/source"
                  type="url"
                  value={endorsementDraft.sourceUrl}
                />
              </FieldShell>

              <FieldShell htmlFor="endorsement-draft-notes" label="Notes">
                <TextArea
                  id="endorsement-draft-notes"
                  onChange={(event) => setEndorsementDraft((current) => ({ ...current, notes: event.target.value }))}
                  value={endorsementDraft.notes}
                />
              </FieldShell>

              <FieldShell htmlFor="endorsement-draft-review-notes" label="Review notes">
                <TextArea
                  id="endorsement-draft-review-notes"
                  onChange={(event) => setEndorsementDraft((current) => ({ ...current, reviewNotes: event.target.value }))}
                  value={endorsementDraft.reviewNotes}
                />
              </FieldShell>

              <div className="flex flex-wrap gap-3">
                <button
                  className="rounded-none border-2 border-primary bg-primary px-5 py-3 font-black text-primary-foreground transition hover:bg-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                  type="submit"
                >
                  Save draft
                </button>
                <button
                  className="rounded-none border-2 border-border-strong bg-surface px-5 py-3 font-black text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                  type="button"
                  onClick={async () =>
                    runAction(async () => {
                      const drafts = await listAdminEndorsementDrafts(token);
                      setEndorsementDrafts(drafts);
                    }, 'Endorsement drafts loaded')
                  }
                >
                  List drafts
                </button>
                <button
                  className="rounded-none border-2 border-border-strong bg-surface px-5 py-3 font-black text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                  type="button"
                  onClick={async () =>
                    runAction(async () => {
                      if (!endorsementDraft.id.trim()) {
                        throw new Error('Enter an endorsement draft ID before loading it.');
                      }
                      const draft = await loadAdminEndorsementDraft(token, endorsementDraft.id);
                      resetEndorsementDraftFromResponse(draft);
                    }, 'Endorsement draft loaded')
                  }
                >
                  Load by ID
                </button>
              </div>
            </form>

            {endorsementDrafts.length > 0 ? (
              <ul className="mt-6 space-y-3">
                {endorsementDrafts.map((draft) => (
                  <DraftListCard
                    draft={draft}
                    key={draft.id}
                    onArchive={async () =>
                      runAction(async () => {
                        await archiveAdminEndorsementDraft(token, draft.id);
                        const refreshed = await listAdminEndorsementDrafts(token);
                        setEndorsementDrafts(refreshed);
                      }, 'Endorsement draft archived')
                    }
                    onLoad={() => resetEndorsementDraftFromResponse(draft)}
                    onPublish={async () =>
                      runAction(async () => {
                        await publishAdminEndorsementDraft(token, draft.id);
                        const refreshed = await listAdminEndorsementDrafts(token);
                        setEndorsementDrafts(refreshed);
                      }, 'Endorsement draft published')
                    }
                  />
                ))}
              </ul>
            ) : (
              <p className="mt-6 rounded-none border-2 border-dashed border-border bg-surface-muted p-4 text-sm leading-6 text-muted">
                Load endorsement drafts to review, publish, or archive them.
              </p>
            )}
          </SectionShell>
        </section>
      </section>
    </main>
  );
}
