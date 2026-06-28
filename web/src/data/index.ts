import { californiaDirectoryData } from './california';
import type { ConfidenceLevel, EndorsementType } from './schema';

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

export function getEndorsementCards(): EndorsementCardData[] {
  return directoryData.endorsements.map((endorsement) => {
    const candidate = directoryData.candidates.find((item) => item.id === endorsement.candidateId);
    const race = directoryData.races.find((item) => item.id === endorsement.raceId);
    const election = directoryData.elections.find((item) => item.id === race?.electionId);
    const chapter = directoryData.chapters.find((item) => item.id === endorsement.chapterId);

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

export function getMeasureCards(): MeasureCardData[] {
  return directoryData.measurePositions.map((position) => {
    const measure = directoryData.ballotMeasures.find((item) => item.id === position.measureId);
    const election = directoryData.elections.find((item) => item.id === measure?.electionId);
    const chapter = directoryData.chapters.find((item) => item.id === position.chapterId);

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

export function getVoterGuideCards(): VoterGuideCardData[] {
  return directoryData.voterGuides.map((guide) => {
    const chapter = directoryData.chapters.find((item) => item.id === guide.chapterId);
    const election = directoryData.elections.find((item) => item.id === guide.electionId);

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
