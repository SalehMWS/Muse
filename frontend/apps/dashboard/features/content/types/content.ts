export type ContentStatus = "draft" | "scheduled" | "published" | "archived";

export interface Content {
  id: string;
  title: string;
  caption: string;
  status: ContentStatus;
  language: string;
  contentType: string;
  visibility: string;
  tags: string[];
  publishedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface ContentDto {
  id: string;
  title: string;
  caption: string;
  status: string;
  language: string;
  content_type: string;
  visibility: string;
  tags: string[];
  published_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ContentListDto {
  items: ContentDto[];
  next_cursor?: string;
}

export function toContent(dto: ContentDto): Content {
  return {
    id: dto.id,
    title: dto.title,
    caption: dto.caption,
    status: dto.status as ContentStatus,
    language: dto.language,
    contentType: dto.content_type,
    visibility: dto.visibility,
    tags: dto.tags ?? [],
    publishedAt: dto.published_at ?? null,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  };
}

export interface ContentFilters {
  status?: ContentStatus;
  language?: string;
  type?: string;
  tag?: string;
  limit?: number;
  cursor?: string | null;
}
