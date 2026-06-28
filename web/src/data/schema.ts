export const endorsementTypes = [
  'endorsed',
  'recommended',
  'opposed',
  'no_recommendation',
  'write_in',
  'coalition',
  'unknown',
] as const;

export const confidenceLevels = [
  'official_source',
  'official_social_source',
  'campaign_source',
  'press_source',
  'needs_review',
] as const;

export type EndorsementType = (typeof endorsementTypes)[number];
export type ConfidenceLevel = (typeof confidenceLevels)[number];

export type Chapter = {
  id: string;
  name: string;
  state: string;
  websiteUrl?: string;
  officialDsaUrl?: string;
  coverageNotes?: string;
  active?: boolean;
};

export type Election = {
  id: string;
  name: string;
  electionDate?: string;
  electionType?: string;
  jurisdiction?: string;
  officialUrl?: string;
};

export type Race = {
  id: string;
  electionId: string;
  office: string;
  district?: string;
  level: 'federal' | 'statewide' | 'state_legislature' | 'local' | 'unknown';
  jurisdiction: string;
};

export type Candidate = {
  id: string;
  name: string;
  campaignUrl?: string;
  party?: string;
  incumbent?: boolean;
  photoUrl?: string;
};

export type VoterGuide = {
  id: string;
  title: string;
  chapterId?: string;
  electionId?: string;
  sourceUrl: string;
  confidence: ConfidenceLevel;
  publishedAt?: string;
  summary?: string;
};

export type Endorsement = {
  id: string;
  candidateId: string;
  raceId: string;
  chapterId: string;
  voterGuideId?: string;
  endorsementType: EndorsementType;
  sourceUrl: string;
  sourceDate?: string;
  confidence: ConfidenceLevel;
  notes?: string;
};

export type BallotMeasure = {
  id: string;
  electionId: string;
  title: string;
  jurisdiction: string;
  measureNumber?: string;
  summary?: string;
};

export type MeasurePosition = {
  id: string;
  measureId: string;
  chapterId: string;
  position: 'yes' | 'no' | 'no_recommendation' | 'opposed' | 'unknown';
  sourceUrl: string;
  confidence: ConfidenceLevel;
};

export type DirectoryData = {
  chapters: Chapter[];
  elections: Election[];
  races: Race[];
  candidates: Candidate[];
  voterGuides: VoterGuide[];
  endorsements: Endorsement[];
  ballotMeasures: BallotMeasure[];
  measurePositions: MeasurePosition[];
};

const endorsementTypeLabels: Record<EndorsementType, string> = {
  endorsed: 'Endorsed',
  recommended: 'Recommended',
  opposed: 'Opposed',
  no_recommendation: 'No recommendation',
  write_in: 'Write-in',
  coalition: 'Coalition',
  unknown: 'Unknown',
};

const confidenceLevelLabels: Record<ConfidenceLevel, string> = {
  official_source: 'Official source',
  official_social_source: 'Official social source',
  campaign_source: 'Campaign source',
  press_source: 'Press source',
  needs_review: 'Needs review',
};

export function endorsementTypeLabel(type: EndorsementType): string {
  return endorsementTypeLabels[type];
}

export function confidenceLabel(level: ConfidenceLevel): string {
  return confidenceLevelLabels[level];
}

export function validateDirectoryData(data: DirectoryData): string[] {
  const errors: string[] = [];

  for (const guide of data.voterGuides) {
    if (!guide.sourceUrl.trim()) {
      errors.push(`voterGuides[${guide.id}] must include sourceUrl`);
    }
  }

  for (const endorsement of data.endorsements) {
    if (!endorsement.sourceUrl.trim()) {
      errors.push(`endorsements[${endorsement.id}] must include sourceUrl`);
    }
  }

  for (const position of data.measurePositions) {
    if (!position.sourceUrl.trim()) {
      errors.push(`measurePositions[${position.id}] must include sourceUrl`);
    }
  }

  return errors;
}
