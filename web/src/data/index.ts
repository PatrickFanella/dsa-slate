import { californiaDirectoryData } from './california';
import { validateDirectoryData, type ConfidenceLevel, type DirectoryData, type EndorsementType } from './schema';

export const directoryData = californiaDirectoryData;

export type EndorsementCardData = {
  id: string;
  candidateName: string;
  office: string;
  district?: string;
  electionName: string;
  chapterName: string;
  endorsementType: EndorsementType;
  confidence: ConfidenceLevel;
  sourceUrl: string;
  sourceDate?: string;
};

export type MeasureCardData = {
  id: string;
  title: string;
  jurisdiction: string;
  electionName: string;
  chapterName: string;
  position: string;
  confidence: ConfidenceLevel;
  sourceUrl: string;
};

export type VoterGuideCardData = {
  id: string;
  title: string;
  chapterName: string;
  electionName?: string;
  confidence: ConfidenceLevel;
  sourceUrl: string;
  publishedAt?: string;
  summary?: string;
};

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
      id: endorsement.id,
      candidateName: candidate.name,
      office: race.office,
      district: race.district,
      electionName: election.name,
      chapterName: chapter.name,
      endorsementType: endorsement.endorsementType,
      confidence: endorsement.confidence,
      sourceUrl: endorsement.sourceUrl,
      sourceDate: endorsement.sourceDate,
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
      id: position.id,
      title: measure.title,
      jurisdiction: measure.jurisdiction,
      electionName: election.name,
      chapterName: chapter.name,
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
      id: guide.id,
      title: guide.title,
      chapterName: chapter.name,
      electionName: election?.name,
      confidence: guide.confidence,
      sourceUrl: guide.sourceUrl,
      publishedAt: guide.publishedAt,
      summary: guide.summary,
    };
  });
}
