import { z } from "zod";

export const ingestDocumentSchema = z.object({
  title: z.string().min(1, "Title is required.").max(200, "Keep the title under 200 characters."),
  source: z.string().max(500, "Keep the source under 500 characters.").optional(),
  content: z
    .string()
    .min(20, "Add at least 20 characters so there is something worth indexing."),
});

export type IngestDocumentInput = z.infer<typeof ingestDocumentSchema>;

export const knowledgeQuerySchema = z.object({
  query: z.string().min(3, "Ask a question with at least 3 characters."),
  topK: z
    .number()
    .int("Pick a whole number of results.")
    .min(1, "Retrieve at least 1 result.")
    .max(20, "Retrieve at most 20 results.")
    .optional(),
});

export type KnowledgeQueryInput = z.infer<typeof knowledgeQuerySchema>;
