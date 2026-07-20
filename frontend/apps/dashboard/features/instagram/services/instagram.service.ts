import { request } from "@/services/api-client";

import {
  toAccount,
  toConnect,
  type Account,
  type AccountDto,
  type Connect,
  type ConnectDto,
} from "../types/account";

export const instagramService = {
  async connect(signal?: AbortSignal): Promise<Connect> {
    const dto = await request<ConnectDto>({ url: "/instagram/connect", method: "GET", signal });
    return toConnect(dto);
  },

  async list(signal?: AbortSignal): Promise<Account[]> {
    const dtos = await request<AccountDto[]>({
      url: "/instagram/accounts",
      method: "GET",
      signal,
    });
    return (dtos ?? []).map(toAccount);
  },

  async refresh(id: string): Promise<Account> {
    const dto = await request<AccountDto>({
      url: `/instagram/accounts/${id}/refresh`,
      method: "POST",
    });
    return toAccount(dto);
  },

  async disconnect(id: string): Promise<void> {
    await request<null>({ url: `/instagram/accounts/${id}`, method: "DELETE" });
  },
};
