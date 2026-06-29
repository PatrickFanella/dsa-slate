import { californiaDirectoryData } from './california';
import { validateDirectoryData, type ConfidenceLevel, type DirectoryData, type EndorsementType } from './schema';

export const directoryData = californiaDirectoryData;

export type DirectoryCardKind = 'endorsement' | 'measure' | 'voterGuide';
export type DirectoryCardKey = `${DirectoryCardKind}:${string}`;

export type EndorsementCardData = {
  kind: 'endorsement';
  key: DirectoryCardKey;
  id: string;
  candidateId: string;
  candidateName: string;
  candidateParty?: string;
  candidateCampaignUrl?: string;
  raceId: string;
  office: string;
  district?: string;
  raceLevel: DirectoryData['races'][number]['level'];
  raceJurisdiction: string;
  electionId: string;
  electionName: string;
  electionDate?: string;
  chapterId: string;
  chapterName: string;
  chapterWebsiteUrl?: string;
  voterGuideId?: string;
  endorsementType: EndorsementType;
  confidence: ConfidenceLevel;
  sourceUrl: string;
  sourceDate?: string;
  notes?: string;
};

export type MeasureCardData = {
  kind: 'measure';
  key: DirectoryCardKey;
  id: string;
  measureId: string;
  title: string;
  jurisdiction: string;
  measureNumber?: string;
  summary?: string;
  electionId: string;
  electionName: string;
  electionDate?: string;
  chapterId: string;
  chapterName: string;
  chapterWebsiteUrl?: string;
  position: string;
  confidence: ConfidenceLevel;
  sourceUrl: string;
};

export type VoterGuideCardData = {
  kind: 'voterGuide';
  key: DirectoryCardKey;
  id: string;
  title: string;
  chapterId: string;
  chapterName: string;
  chapterWebsiteUrl?: string;
  electionName?: string;
  electionId?: string;
  electionDate?: string;
  confidence: ConfidenceLevel;
  sourceUrl: string;
  publishedAt?: string;
  summary?: string;
};

export type DirectoryCardData = EndorsementCardData | MeasureCardData | VoterGuideCardData;

export type DirectoryCards = {
  endorsements: EndorsementCardData[];
  measures: MeasureCardData[];
  voterGuides: VoterGuideCardData[];
};

export function assertValidDirectoryData(data: DirectoryData) {
  const errors = validateDirectoryData(data);

  if (errors.length > 0) {
    throw new Error(`Directory data validation failed:\n${errors.join('\n')}`);
  }
}

export function buildDirectoryCards(data: DirectoryData): DirectoryCards {
  assertValidDirectoryData(data);

  return {
    endorsements: buildEndorsementCards(data),
    measures: buildMeasureCards(data),
    voterGuides: buildVoterGuideCards(data),
  };
}

export function getEndorsementCards(data: DirectoryData = directoryData): EndorsementCardData[] {
  assertValidDirectoryData(data);

  return buildEndorsementCards(data);
}

export function allDirectoryCards(cards: DirectoryCards): DirectoryCardData[] {
  return [...cards.voterGuides, ...cards.endorsements, ...cards.measures];
}

function buildEndorsementCards(data: DirectoryData): EndorsementCardData[] {
  return data.endorsements.map((endorsement) => {
    const candidate = data.candidates.find((item) => item.id === endorsement.candidateId);
    const race = data.races.find((item) => item.id === endorsement.raceId);
    const election = data.elections.find((item) => item.id === race?.electionId);
    const chapter = data.chapters.find((item) => item.id === endorsement.chapterId);

    if (!candidate || !race || !election || !chapter) {
      throw new Error(`Unable to build endorsement card ${endorsement.id}`);
    }

    return {
      kind: 'endorsement',
      key: `endorsement:${endorsement.id}`,
      id: endorsement.id,
      candidateId: endorsement.candidateId,
      candidateName: candidate.name,
      candidateParty: candidate.party,
      candidateCampaignUrl: candidate.campaignUrl,
      raceId: endorsement.raceId,
      office: race.office,
      district: race.district,
      raceLevel: race.level,
      raceJurisdiction: race.jurisdiction,
      electionId: race.electionId,
      electionName: election.name,
      electionDate: election.electionDate,
      chapterId: endorsement.chapterId,
      chapterName: chapter.name,
      chapterWebsiteUrl: chapter.websiteUrl,
      voterGuideId: endorsement.voterGuideId,
      endorsementType: endorsement.endorsementType,
      confidence: endorsement.confidence,
      sourceUrl: endorsement.sourceUrl,
      sourceDate: endorsement.sourceDate,
      notes: endorsement.notes,
    };
  });
}

export function getMeasureCards(data: DirectoryData = directoryData): MeasureCardData[] {
  assertValidDirectoryData(data);

  return buildMeasureCards(data);
}

function buildMeasureCards(data: DirectoryData): MeasureCardData[] {
  return data.measurePositions.map((position) => {
    const measure = data.ballotMeasures.find((item) => item.id === position.measureId);
    const election = data.elections.find((item) => item.id === measure?.electionId);
    const chapter = data.chapters.find((item) => item.id === position.chapterId);

    if (!measure || !election || !chapter) {
      throw new Error(`Unable to build measure card ${position.id}`);
    }

    return {
      kind: 'measure',
      key: `measure:${position.id}`,
      id: position.id,
      measureId: position.measureId,
      title: measure.title,
      jurisdiction: measure.jurisdiction,
      measureNumber: measure.measureNumber,
      summary: measure.summary,
      electionId: measure.electionId,
      electionName: election.name,
      electionDate: election.electionDate,
      chapterId: position.chapterId,
      chapterName: chapter.name,
      chapterWebsiteUrl: chapter.websiteUrl,
      position: position.position,
      confidence: position.confidence,
      sourceUrl: position.sourceUrl,
    };
  });
}

export function getVoterGuideCards(data: DirectoryData = directoryData): VoterGuideCardData[] {
  assertValidDirectoryData(data);

  return buildVoterGuideCards(data);
}

function buildVoterGuideCards(data: DirectoryData): VoterGuideCardData[] {
  return data.voterGuides.map((guide) => {
    const chapter = data.chapters.find((item) => item.id === guide.chapterId);
    const election = data.elections.find((item) => item.id === guide.electionId);

    if (!chapter) {
      throw new Error(`Unable to build voter guide card ${guide.id}`);
    }

    return {
      kind: 'voterGuide',
      key: `voterGuide:${guide.id}`,
      id: guide.id,
      title: guide.title,
      chapterId: guide.chapterId,
      chapterName: chapter.name,
      chapterWebsiteUrl: chapter.websiteUrl,
      electionId: guide.electionId,
      electionName: election?.name,
      electionDate: election?.electionDate,
      confidence: guide.confidence,
      sourceUrl: guide.sourceUrl,
      publishedAt: guide.publishedAt,
      summary: guide.summary,
    };
  });
}
