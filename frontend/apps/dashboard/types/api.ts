export interface ApiEnvelope<TData> {
  success: boolean;
  data?: TData;
  error?: ApiErrorBody;
  meta?: Record<string, unknown>;
}

export interface ApiErrorBody {
  code: string;
  message: string;
}

export type ApiErrorKind =
  | "validation"
  | "authentication"
  | "authorization"
  | "network"
  | "timeout"
  | "rate_limited"
  | "unexpected";

export interface NormalizedError {
  kind: ApiErrorKind;
  code: string;
  message: string;
  status?: number;
  retryAfter?: number;
}
