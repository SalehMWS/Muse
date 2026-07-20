import { request } from "@/services/api-client";
import type { Paginated } from "@/types/pagination";

import { parseTags, type CreateContentInput } from "../schemas/create-content.schema";
import {
  toContent,
  type Content,
  type ContentDto,
  type ContentFilters,
  type ContentListDto,
} from "../types/content";

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

  async archive(id: string): Promise<void> {
    await request<null>({ url: `/contents/${id}`, method: "DELETE" });
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
