import { confidenceLabel, endorsementTypeLabel } from './schema';
import { type DirectoryCards } from './index';

function normalizeText(value: string): string {
  return value.toLowerCase().trim().replace(/\s+/g, ' ');
}

function normalizeQuery(query: string): string {
  return normalizeText(query);
}

function matches(query: string, fields: Array<string | undefined>): boolean {
  if (!query) {
    return true;
  }

  const haystack = normalizeText(fields.filter((field): field is string => Boolean(field)).join(' '));

  return haystack.includes(query);
}

export function filterDirectoryCards(cards: DirectoryCards, query: string): DirectoryCards {
  const normalizedQuery = normalizeQuery(query);

  if (!normalizedQuery) {
    return cards;
  }

  return {
    endorsements: cards.endorsements.filter((card) =>
      matches(normalizedQuery, [
        card.candidateName,
        card.office,
        card.district,
        card.electionName,
        card.chapterName,
        endorsementTypeLabel(card.endorsementType),
        card.endorsementType,
        confidenceLabel(card.confidence),
        card.sourceDate,
      ]),
    ),
    measures: cards.measures.filter((card) =>
      matches(normalizedQuery, [
        card.title,
        card.jurisdiction,
        card.electionName,
        card.chapterName,
        card.position,
        confidenceLabel(card.confidence),
      ]),
    ),
    voterGuides: cards.voterGuides.filter((card) =>
      matches(normalizedQuery, [
        card.title,
        card.chapterName,
        card.electionName,
        confidenceLabel(card.confidence),
        card.publishedAt,
        card.summary,
      ]),
    ),
  };
}
