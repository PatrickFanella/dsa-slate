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
  chapterId: string;
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

function validateUniqueIds(collectionName: string, records: { id: string }[], errors: string[]) {
  const seen = new Set<string>();

  for (const record of records) {
    if (seen.has(record.id)) {
      errors.push(`${collectionName}[${record.id}] duplicate id`);
      continue;
    }

    seen.add(record.id);
  }
}

function idSet(records: { id: string }[]): Set<string> {
  return new Set(records.map((record) => record.id));
}

function isAbsoluteHttpUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

function isIsoDate(value: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    return false;
  }

  const [yearText, monthText, dayText] = value.split('-');
  const year = Number(yearText);
  const month = Number(monthText);
  const day = Number(dayText);
  const date = new Date(Date.UTC(year, month - 1, day));

  date.setUTCFullYear(year);

  return date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day;
}

function validateOptionalUrl(recordName: string, fieldName: string, value: string | undefined, errors: string[]) {
  if (value && !isAbsoluteHttpUrl(value)) {
    errors.push(`${recordName}.${fieldName} must be an absolute http(s) URL`);
  }
}

function validateRequiredUrl(recordName: string, fieldName: string, value: string, errors: string[]) {
  if (!value.trim()) {
    errors.push(`${recordName} must include ${fieldName}`);
    return;
  }

  if (!isAbsoluteHttpUrl(value)) {
    errors.push(`${recordName}.${fieldName} must be an absolute http(s) URL`);
  }
}

function validateOptionalIsoDate(recordName: string, fieldName: string, value: string | undefined, errors: string[]) {
  if (value && !isIsoDate(value)) {
    errors.push(`${recordName}.${fieldName} must be YYYY-MM-DD`);
  }
}

function validateRequiredReference(
  recordName: string,
  fieldName: string,
  value: string,
  targetCollectionName: string,
  targetIds: Set<string>,
  errors: string[],
) {
  if (!targetIds.has(value)) {
    errors.push(`${recordName}.${fieldName} references missing ${targetCollectionName}[${value}]`);
  }
}

function validateOptionalReference(
  recordName: string,
  fieldName: string,
  value: string | undefined,
  targetCollectionName: string,
  targetIds: Set<string>,
  errors: string[],
) {
  if (value && !targetIds.has(value)) {
    errors.push(`${recordName}.${fieldName} references missing ${targetCollectionName}[${value}]`);
  }
}

export function validateDirectoryData(data: DirectoryData): string[] {
  const errors: string[] = [];

  validateUniqueIds('chapters', data.chapters, errors);
  validateUniqueIds('elections', data.elections, errors);
  validateUniqueIds('races', data.races, errors);
  validateUniqueIds('candidates', data.candidates, errors);
  validateUniqueIds('voterGuides', data.voterGuides, errors);
  validateUniqueIds('endorsements', data.endorsements, errors);
  validateUniqueIds('ballotMeasures', data.ballotMeasures, errors);
  validateUniqueIds('measurePositions', data.measurePositions, errors);

  const chapterIds = idSet(data.chapters);
  const electionIds = idSet(data.elections);
  const raceIds = idSet(data.races);
  const candidateIds = idSet(data.candidates);
  const voterGuideIds = idSet(data.voterGuides);
  const ballotMeasureIds = idSet(data.ballotMeasures);

  for (const race of data.races) {
    validateRequiredReference(`races[${race.id}]`, 'electionId', race.electionId, 'elections', electionIds, errors);
  }

  for (const chapter of data.chapters) {
    validateOptionalUrl(`chapters[${chapter.id}]`, 'websiteUrl', chapter.websiteUrl, errors);
    validateOptionalUrl(`chapters[${chapter.id}]`, 'officialDsaUrl', chapter.officialDsaUrl, errors);
  }

  for (const election of data.elections) {
    validateOptionalIsoDate(`elections[${election.id}]`, 'electionDate', election.electionDate, errors);
    validateOptionalUrl(`elections[${election.id}]`, 'officialUrl', election.officialUrl, errors);
  }

  for (const candidate of data.candidates) {
    validateOptionalUrl(`candidates[${candidate.id}]`, 'campaignUrl', candidate.campaignUrl, errors);
    validateOptionalUrl(`candidates[${candidate.id}]`, 'photoUrl', candidate.photoUrl, errors);
  }

  for (const guide of data.voterGuides) {
    if (typeof guide.chapterId !== 'string' || !guide.chapterId.trim()) {
      errors.push(`voterGuides[${guide.id}] must include chapterId`);
    } else {
      validateRequiredReference(`voterGuides[${guide.id}]`, 'chapterId', guide.chapterId, 'chapters', chapterIds, errors);
    }
    validateOptionalReference(`voterGuides[${guide.id}]`, 'electionId', guide.electionId, 'elections', electionIds, errors);
    validateRequiredUrl(`voterGuides[${guide.id}]`, 'sourceUrl', guide.sourceUrl, errors);
    validateOptionalIsoDate(`voterGuides[${guide.id}]`, 'publishedAt', guide.publishedAt, errors);
  }

  for (const endorsement of data.endorsements) {
    validateRequiredReference(
      `endorsements[${endorsement.id}]`,
      'candidateId',
      endorsement.candidateId,
      'candidates',
      candidateIds,
      errors,
    );
    validateRequiredReference(
      `endorsements[${endorsement.id}]`,
      'raceId',
      endorsement.raceId,
      'races',
      raceIds,
      errors,
    );
    validateRequiredReference(
      `endorsements[${endorsement.id}]`,
      'chapterId',
      endorsement.chapterId,
      'chapters',
      chapterIds,
      errors,
    );
    validateOptionalReference(
      `endorsements[${endorsement.id}]`,
      'voterGuideId',
      endorsement.voterGuideId,
      'voterGuides',
      voterGuideIds,
      errors,
    );
    validateRequiredUrl(`endorsements[${endorsement.id}]`, 'sourceUrl', endorsement.sourceUrl, errors);
    validateOptionalIsoDate(`endorsements[${endorsement.id}]`, 'sourceDate', endorsement.sourceDate, errors);
  }

  for (const measure of data.ballotMeasures) {
    validateRequiredReference(
      `ballotMeasures[${measure.id}]`,
      'electionId',
      measure.electionId,
      'elections',
      electionIds,
      errors,
    );
  }

  for (const position of data.measurePositions) {
    validateRequiredReference(
      `measurePositions[${position.id}]`,
      'measureId',
      position.measureId,
      'ballotMeasures',
      ballotMeasureIds,
      errors,
    );
    validateRequiredReference(
      `measurePositions[${position.id}]`,
      'chapterId',
      position.chapterId,
      'chapters',
      chapterIds,
      errors,
    );
    validateRequiredUrl(`measurePositions[${position.id}]`, 'sourceUrl', position.sourceUrl, errors);
  }

  return errors;
}
