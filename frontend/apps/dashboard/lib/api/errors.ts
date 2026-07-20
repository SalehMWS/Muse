import axios from "axios";

import type { ApiEnvelope, ApiErrorKind, NormalizedError } from "@/types/api";

const kindByStatus: Record<number, ApiErrorKind> = {
  400: "validation",
  401: "authentication",
  403: "authorization",
  404: "unexpected",
  409: "validation",
  422: "validation",
  429: "rate_limited",
};

export class ApiError extends Error implements NormalizedError {
  readonly kind: ApiErrorKind;
  readonly code: string;
  readonly status?: number;
  readonly retryAfter?: number;

  constructor(error: NormalizedError) {
    super(error.message);
    this.name = "ApiError";
    this.kind = error.kind;
    this.code = error.code;
    this.status = error.status;
    this.retryAfter = error.retryAfter;
  }
}

export function normalizeError(error: unknown): ApiError {
  if (error instanceof ApiError) {
    return error;
  }

  if (axios.isCancel(error)) {
    return new ApiError({
      kind: "network",
      code: "REQUEST_CANCELLED",
      message: "Request was cancelled.",
    });
  }

  if (axios.isAxiosError(error)) {
    if (error.code === "ECONNABORTED") {
      return new ApiError({
        kind: "timeout",
        code: "TIMEOUT",
        message: "The request took too long. Please try again.",
      });
    }

    if (!error.response) {
      return new ApiError({
        kind: "network",
        code: "NETWORK_ERROR",
        message: "Cannot reach the server. Check your connection and try again.",
      });
    }

    const status = error.response.status;
    const envelope = error.response.data as ApiEnvelope<unknown> | undefined;
    const retryAfterHeader = error.response.headers?.["retry-after"];

    return new ApiError({
      kind: kindByStatus[status] ?? (status >= 500 ? "unexpected" : "unexpected"),
      code: envelope?.error?.code ?? "UNEXPECTED_ERROR",
      message: envelope?.error?.message ?? friendlyMessage(status),
      status,
      retryAfter: retryAfterHeader ? Number(retryAfterHeader) : undefined,
    });
  }

  return new ApiError({
    kind: "unexpected",
    code: "UNEXPECTED_ERROR",
    message: "Something went wrong. Please try again.",
  });
}

function friendlyMessage(status: number): string {
  if (status === 401) return "Your session has expired. Please sign in again.";
  if (status === 403) return "You do not have permission to do that.";
  if (status === 404) return "We could not find what you were looking for.";
  if (status === 429) return "Too many requests. Please wait a moment and try again.";
  if (status >= 500) return "The server ran into a problem. Please try again.";
  return "Something went wrong. Please try again.";
}
