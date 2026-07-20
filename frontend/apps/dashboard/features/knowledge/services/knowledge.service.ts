import { ApiError } from "@/lib/api/errors";
import { request } from "@/services/api-client";

import type {
  IngestDocumentInput,
  KnowledgeQueryInput,
} from "../schemas/ingest-document.schema";
import {
  toDocument,
  toQueryResult,
  type DocumentDto,
  type KnowledgeDocument,
  type QueryResult,
  type QueryResultDto,
} from "../types/document";

const DEFAULT_TOP_K = 5;

export const knowledgeService = {
  async list(signal?: AbortSignal): Promise<KnowledgeDocument[]> {
    const data = await request<DocumentDto[]>({
      url: "/knowledge/documents",
      method: "GET",
      signal,
    });

    return (data ?? []).map(toDocument);
  },

  async ingest(input: IngestDocumentInput): Promise<KnowledgeDocument> {
    const dto = await request<DocumentDto>({
      url: "/knowledge/documents",
      method: "POST",
      data: {
        title: input.title,
        source: input.source ?? "",
        content: input.content,
      },
    });

    return toDocument(dto);
  },

  async remove(id: string): Promise<void> {
    try {
      await request<null>({ url: `/knowledge/documents/${id}`, method: "DELETE" });
    } catch (error) {
      // This endpoint answers with a bare 204 and no response envelope, which the
      // shared `request` helper reports as a failed envelope. Anything else is real.
      if (error instanceof ApiError && error.status === 204) {
        return;
      }
      throw error;
    }
  },

  async query(input: KnowledgeQueryInput, signal?: AbortSignal): Promise<QueryResult> {
    const dto = await request<QueryResultDto>({
      url: "/knowledge/query",
      method: "POST",
      signal,
      data: {
        query: input.query,
        top_k: input.topK ?? DEFAULT_TOP_K,
      },
    });

    return toQueryResult(dto);
  },
};
