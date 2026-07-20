export type DocumentStatus = "pending" | "indexed" | "failed";

/**
 * Named `KnowledgeDocument` rather than `Document` on purpose: a bare `Document`
 * export shadows the DOM `Document` global inside every module that imports it.
 */
export interface KnowledgeDocument {
  id: string;
  title: string;
  source: string;
  status: DocumentStatus;
  chunkCount: number;
  lastError: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface DocumentDto {
  id: string;
  title: string;
  source?: string;
  status: string;
  chunk_count?: number;
  last_error?: string | null;
  created_at: string;
  updated_at?: string;
}

export function toDocument(dto: DocumentDto): KnowledgeDocument {
  return {
    id: dto.id,
    title: dto.title,
    source: dto.source ?? "",
    status: dto.status as DocumentStatus,
    chunkCount: dto.chunk_count ?? 0,
    lastError: dto.last_error ?? null,
    createdAt: dto.created_at,
    // The API does not currently serialise `updated_at` on the document payload,
    // so fall back to `created_at` to keep the model total.
    updatedAt: dto.updated_at ?? dto.created_at,
  };
}

export interface QueryHit {
  documentId: string;
  chunkIndex: number;
  content: string;
  score: number;
}

export interface QueryHitDto {
  document_id: string;
  chunk_index: number;
  content: string;
  score: number;
}

export interface QueryResult {
  hits: QueryHit[];
  context: string;
}

export interface QueryResultDto {
  hits?: QueryHitDto[] | null;
  context?: string;
}

export function toQueryHit(dto: QueryHitDto): QueryHit {
  return {
    documentId: dto.document_id,
    chunkIndex: dto.chunk_index,
    content: dto.content,
    score: dto.score,
  };
}

export function toQueryResult(dto: QueryResultDto): QueryResult {
  return {
    hits: (dto.hits ?? []).map(toQueryHit),
    context: dto.context ?? "",
  };
}
