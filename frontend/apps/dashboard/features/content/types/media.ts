export type MediaType = "image" | "video";

export interface Media {
  id: string;
  contentId: string;
  url: string;
  mediaType: MediaType;
  position: number;
  createdAt: string;
}

export interface MediaDto {
  id: string;
  content_id: string;
  url: string;
  media_type: string;
  position: number;
  created_at: string;
}

export function toMedia(dto: MediaDto): Media {
  return {
    id: dto.id,
    contentId: dto.content_id,
    url: dto.url,
    mediaType: dto.media_type === "video" ? "video" : "image",
    position: dto.position ?? 0,
    createdAt: dto.created_at,
  };
}

export interface AttachMediaInput {
  url: string;
  mediaType: MediaType;
  position?: number;
}
