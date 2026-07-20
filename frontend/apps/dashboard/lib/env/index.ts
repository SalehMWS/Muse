import { z } from "zod";

const serverSchema = z.object({
  API_BASE_URL: z.url().default("http://127.0.0.1:8099"),
  NODE_ENV: z.enum(["development", "test", "production"]).default("development"),
});

const clientSchema = z.object({
  NEXT_PUBLIC_APP_NAME: z.string().default("NovaFlow"),
});

export const serverEnv = serverSchema.parse({
  API_BASE_URL: process.env.API_BASE_URL,
  NODE_ENV: process.env.NODE_ENV,
});

export const clientEnv = clientSchema.parse({
  NEXT_PUBLIC_APP_NAME: process.env.NEXT_PUBLIC_APP_NAME,
});

export const isProduction = serverEnv.NODE_ENV === "production";
