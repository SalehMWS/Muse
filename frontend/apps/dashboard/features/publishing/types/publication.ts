export type PublicationStatus = "published" | "failed" | "pending";

export interface Publication {
  id: string;
  contentId: string;
  instagramAccountId: string;
  platform: string;
  status: PublicationStatus;
  platformPostId: string | null;
  permalink: string | null;
  publishedAt: string | null;
  createdAt: string;
}

export interface PublicationDto {
  id: string;
  content_id: string;
  instagram_account_id: string;
  platform: string;
  status: string;
  platform_post_id?: string | null;
  permalink?: string | null;
  published_at?: string | null;
  created_at: string;
}

export function toPublication(dto: PublicationDto): Publication {
  return {
    id: dto.id,
    contentId: dto.content_id,
    instagramAccountId: dto.instagram_account_id,
    platform: dto.platform,
    status: dto.status as PublicationStatus,
    platformPostId: dto.platform_post_id ?? null,
    permalink: dto.permalink ?? null,
    publishedAt: dto.published_at ?? null,
    createdAt: dto.created_at,
  };
}

/**
 * Minimal shape of an Instagram account, duplicated here so the publishing
 * feature stays independent of the instagram feature.
 */
export interface InstagramAccount {
  id: string;
  username: string;
  status: string;
}

export interface InstagramAccountDto {
  id: string;
  username: string;
  status: string;
}

export function toInstagramAccount(dto: InstagramAccountDto): InstagramAccount {
  return { id: dto.id, username: dto.username, status: dto.status };
}

export type PublishMediaType = "image" | "carousel" | "reel";

export interface PublishInput {
  instagramAccountId: string;
  mediaType?: PublishMediaType;
}
