import { request } from "@/services/api-client";

import { toProfile, type Profile, type ProfileDto } from "../types/profile";

export const settingsService = {
  async profile(signal?: AbortSignal): Promise<Profile> {
    const dto = await request<ProfileDto>({ url: "/auth/me", method: "GET", signal });
    return toProfile(dto);
  },

  async signOut(): Promise<void> {
    await request<null>({ url: "/auth/logout", method: "POST" });
  },
};
