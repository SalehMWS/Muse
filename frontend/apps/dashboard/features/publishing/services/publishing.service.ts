import { request } from "@/services/api-client";

import {
  toInstagramAccount,
  toPublication,
  type InstagramAccount,
  type InstagramAccountDto,
  type Publication,
  type PublicationDto,
  type PublishInput,
} from "../types/publication";

export const publishingService = {
  async listPublications(contentId: string, signal?: AbortSignal): Promise<Publication[]> {
    const dtos = await request<PublicationDto[]>({
      url: `/contents/${contentId}/publications`,
      method: "GET",
      signal,
    });
    return (dtos ?? []).map(toPublication);
  },

  async publish(contentId: string, input: PublishInput): Promise<Publication> {
    const dto = await request<PublicationDto>({
      url: `/contents/${contentId}/publish`,
      method: "POST",
      data: {
        instagram_account_id: input.instagramAccountId,
        ...(input.mediaType ? { media_type: input.mediaType } : {}),
      },
    });
    return toPublication(dto);
  },

  async listAccounts(signal?: AbortSignal): Promise<InstagramAccount[]> {
    const dtos = await request<InstagramAccountDto[]>({
      url: "/instagram/accounts",
      method: "GET",
      signal,
    });
    return (dtos ?? []).map(toInstagramAccount);
  },
};
