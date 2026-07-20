import { ApiError } from "@/lib/api/errors";
import { request } from "@/services/api-client";
import type { Paginated } from "@/types/pagination";

import { parseTags, type CreateContentInput } from "../schemas/create-content.schema";
import {
  toContent,
  type Content,
  type ContentDto,
  type ContentFilters,
  type ContentListDto,
  type UpdateContentInput,
} from "../types/content";
import { toMedia, type AttachMediaInput, type Media, type MediaDto } from "../types/media";

function toUpdatePayload(patch: UpdateContentInput): Record<string, unknown> {
  const payload: Record<string, unknown> = {};
  if (patch.title !== undefined) payload.title = patch.title;
  if (patch.caption !== undefined) payload.caption = patch.caption;
  if (patch.status !== undefined) payload.status = patch.status;
  if (patch.language !== undefined) payload.language = patch.language;
  if (patch.contentType !== undefined) payload.content_type = patch.contentType;
  if (patch.visibility !== undefined) payload.visibility = patch.visibility;
  if (patch.tags !== undefined) payload.tags = patch.tags;
  return payload;
}

export const contentService = {
  async list(filters: ContentFilters, signal?: AbortSignal): Promise<Paginated<Content>> {
    const data = await request<ContentListDto>({
      url: "/contents",
      method: "GET",
      signal,
      params: {
        status: filters.status,
        language: filters.language,
        type: filters.type,
        tag: filters.tag,
        limit: filters.limit,
        cursor: filters.cursor || undefined,
      },
    });

    return {
      items: (data.items ?? []).map(toContent),
      cursor: data.next_cursor ?? null,
      hasNext: Boolean(data.next_cursor),
    };
  },

  async get(id: string, signal?: AbortSignal): Promise<Content> {
    const dto = await request<ContentDto>({ url: `/contents/${id}`, method: "GET", signal });
    return toContent(dto);
  },

  async create(input: CreateContentInput): Promise<Content> {
    const dto = await request<ContentDto>({
      url: "/contents",
      method: "POST",
      data: {
        title: input.title,
        caption: input.caption ?? "",
        content_type: input.contentType,
        language: input.language,
        tags: parseTags(input.tags),
      },
    });
    return toContent(dto);
  },

  async update(id: string, patch: UpdateContentInput): Promise<Content> {
    const dto = await request<ContentDto>({
      url: `/contents/${id}`,
      method: "PATCH",
      data: toUpdatePayload(patch),
    });
    return toContent(dto);
  },

  async archive(id: string): Promise<void> {
    await request<null>({ url: `/contents/${id}`, method: "DELETE" });
  },

  async duplicate(id: string): Promise<Content> {
    const dto = await request<ContentDto>({ url: `/contents/${id}/duplicate`, method: "POST" });
    return toContent(dto);
  },

  async listMedia(id: string, signal?: AbortSignal): Promise<Media[]> {
    const dtos = await request<MediaDto[]>({
      url: `/contents/${id}/media`,
      method: "GET",
      signal,
    });
    return (dtos ?? []).map(toMedia).sort((a, b) => a.position - b.position);
  },

  async attachMedia(id: string, input: AttachMediaInput): Promise<Media> {
    const dto = await request<MediaDto>({
      url: `/contents/${id}/media`,
      method: "POST",
      data: {
        url: input.url,
        media_type: input.mediaType,
        position: input.position ?? 0,
      },
    });
    return toMedia(dto);
  },

  async deleteMedia(id: string, mediaId: string): Promise<void> {
    try {
      await request<null>({ url: `/contents/${id}/media/${mediaId}`, method: "DELETE" });
    } catch (error) {
      // The endpoint answers 204 with an empty body, which the envelope-aware
      // request helper reads as a failed envelope. Only that case is benign.
      if (error instanceof ApiError && error.status === 204) {
        return;
      }
      throw error;
    }
  },

  async generateCaption(id: string, prompt?: string): Promise<Content> {
    const dto = await request<ContentDto>({
      url: `/contents/${id}/caption`,
      method: "POST",
      data: { prompt: prompt ?? "" },
    });
    return toContent(dto);
  },
};
