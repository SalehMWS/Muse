import axios, { type AxiosInstance, type AxiosRequestConfig } from "axios";

import { normalizeError } from "@/lib/api/errors";
import type { ApiEnvelope } from "@/types/api";

const DEFAULT_TIMEOUT = 30_000;
const UPLOAD_TIMEOUT = 300_000;
const DOWNLOAD_TIMEOUT = 120_000;
const MAX_RETRIES = 3;

const RETRYABLE_METHODS = new Set(["get", "head", "options"]);

function createClient(): AxiosInstance {
  const instance = axios.create({
    baseURL: "/api/v1",
    timeout: DEFAULT_TIMEOUT,
    withCredentials: true,
    headers: { "Content-Type": "application/json" },
  });

  instance.interceptors.request.use((config) => {
    config.headers.set("X-Request-ID", crypto.randomUUID());
    config.headers.set("X-Locale", typeof navigator === "undefined" ? "en" : navigator.language);
    return config;
  });

  instance.interceptors.response.use(
    (response) => response,
    async (error) => {
      const config = error?.config as (AxiosRequestConfig & { __retryCount?: number }) | undefined;
      const status = error?.response?.status;
      const method = (config?.method ?? "get").toLowerCase();

      const retryable =
        config &&
        RETRYABLE_METHODS.has(method) &&
        (status === undefined || status === 429 || status >= 500);

      if (retryable) {
        const attempt = (config.__retryCount ?? 0) + 1;
        if (attempt <= MAX_RETRIES) {
          config.__retryCount = attempt;
          await delay(backoffMs(attempt, error?.response?.headers?.["retry-after"]));
          return instance.request(config);
        }
      }

      return Promise.reject(normalizeError(error));
    },
  );

  return instance;
}

function backoffMs(attempt: number, retryAfter?: string): number {
  if (retryAfter) {
    const seconds = Number(retryAfter);
    if (Number.isFinite(seconds)) {
      return seconds * 1000;
    }
  }
  const base = 300 * 2 ** (attempt - 1);
  return base + Math.random() * 200;
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export const apiClient = createClient();

export async function request<TData>(config: AxiosRequestConfig): Promise<TData> {
  const response = await apiClient.request<ApiEnvelope<TData>>(config);
  const envelope = response.data;

  if (!envelope?.success) {
    throw normalizeError({
      response: { status: response.status, data: envelope, headers: response.headers },
      isAxiosError: true,
    });
  }

  return envelope.data as TData;
}

export const uploadConfig: AxiosRequestConfig = { timeout: UPLOAD_TIMEOUT };
export const downloadConfig: AxiosRequestConfig = { timeout: DOWNLOAD_TIMEOUT };
