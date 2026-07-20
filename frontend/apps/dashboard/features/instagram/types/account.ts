export type AccountStatus = "active" | "expired" | "revoked";

export interface Account {
  id: string;
  instagramUserId: string;
  username: string;
  accountType: string | null;
  status: AccountStatus;
  tokenExpiresAt: string;
  connectedAt: string;
  lastRefreshedAt: string | null;
  /** Days until the long-lived token expires. Negative once already expired. */
  expiresInDays: number;
}

export interface AccountDto {
  id: string;
  instagram_user_id: string;
  username: string;
  account_type?: string | null;
  status?: string;
  token_expires_at: string;
  connected_at: string;
  last_refreshed_at?: string | null;
  /** Not sent by the current backend — derived from token_expires_at when absent. */
  expires_in_days?: number;
}

export interface ConnectDto {
  authorization_url: string;
  state: string;
}

export interface Connect {
  authorizationUrl: string;
  state: string;
}

const MS_PER_DAY = 86_400_000;

/** Whole days from now until `iso`. Returns 0 when the timestamp is unusable. */
export function daysUntil(iso: string, now: number = Date.now()): number {
  const target = Date.parse(iso);
  if (Number.isNaN(target)) return 0;
  return Math.floor((target - now) / MS_PER_DAY);
}

export const EXPIRY_WARNING_DAYS = 7;

export function isExpiringSoon(account: Account): boolean {
  return account.expiresInDays <= EXPIRY_WARNING_DAYS;
}

export function toAccount(dto: AccountDto): Account {
  return {
    id: dto.id,
    instagramUserId: dto.instagram_user_id,
    username: dto.username,
    accountType: dto.account_type ?? null,
    status: (dto.status as AccountStatus) ?? "active",
    tokenExpiresAt: dto.token_expires_at,
    connectedAt: dto.connected_at,
    lastRefreshedAt: dto.last_refreshed_at ?? null,
    expiresInDays: dto.expires_in_days ?? daysUntil(dto.token_expires_at),
  };
}

export function toConnect(dto: ConnectDto): Connect {
  return { authorizationUrl: dto.authorization_url, state: dto.state };
}
