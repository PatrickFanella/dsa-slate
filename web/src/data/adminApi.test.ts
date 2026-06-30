import { describe, expect, it, vi } from 'vitest';
import { emptyAdminChapterDraftForm, listAdminChapterDrafts, publishAdminChapterDraft, saveAdminChapterDraft } from './adminApi';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('adminApi', () => {
  it('does not send CSRF header for admin reads', async () => {
    const fetcher = vi.fn(async () => jsonResponse([]));

    await listAdminChapterDrafts(fetcher);

    const [, options] = fetcher.mock.calls[0] as unknown as [RequestInfo | URL, RequestInit?];
    expect(new Headers(options?.headers).get('X-DSA-Slate-CSRF')).toBeNull();
  });

  it('sends CSRF header for admin writes with JSON bodies', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ id: 'draft-1' }));
    const draft = emptyAdminChapterDraftForm();
    draft.id = 'draft-1';
    draft.chapterId = 'chapter-1';
    draft.name = 'Chapter 1';
    draft.state = 'CA';

    await saveAdminChapterDraft(draft, fetcher);

    const [, options] = fetcher.mock.calls[0] as unknown as [RequestInfo | URL, RequestInit?];
    const headers = new Headers(options?.headers);
    expect(headers.get('X-DSA-Slate-CSRF')).toBe('1');
    expect(headers.get('Content-Type')).toBe('application/json');
  });

  it('sends CSRF header for admin actions without bodies', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ status: 'published' }));

    await publishAdminChapterDraft('draft-1', fetcher);

    const [, options] = fetcher.mock.calls[0] as unknown as [RequestInfo | URL, RequestInit?];
    expect(new Headers(options?.headers).get('X-DSA-Slate-CSRF')).toBe('1');
  });
});
