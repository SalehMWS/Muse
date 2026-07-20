import { request } from "@/services/api-client";

import type { LoginInput, RegisterInput } from "../schemas/auth.schema";
import { toUser, type User, type UserDto } from "../types/user";

export const authService = {
  async login(input: LoginInput): Promise<User> {
    const data = await request<{ user: UserDto }>({
      url: "/auth/login",
      method: "POST",
      data: { email: input.email, password: input.password },
    });
    return toUser(data.user);
  },

  async register(input: RegisterInput): Promise<User> {
    const data = await request<{ user: UserDto }>({
      url: "/auth/register",
      method: "POST",
      data: {
        email: input.email,
        password: input.password,
        display_name: input.displayName,
      },
    });
    return toUser(data.user);
  },

  async logout(): Promise<void> {
    await request<null>({ url: "/auth/logout", method: "POST" });
  },

  async currentUser(signal?: AbortSignal): Promise<User> {
    const dto = await request<UserDto>({ url: "/auth/me", method: "GET", signal });
    return toUser(dto);
  },
};
