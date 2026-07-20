import { z } from "zod";

export const createContentSchema = z.object({
  title: z.string().min(1, "Title is required.").max(200, "Keep the title under 200 characters."),
  caption: z.string().max(2200, "Instagram captions cap out at 2200 characters.").optional(),
  contentType: z.enum(["image", "carousel", "reel", "story"]),
  language: z.string().min(2).max(10),
  tags: z.string().optional(),
});

export type CreateContentInput = z.infer<typeof createContentSchema>;

export function parseTags(raw?: string): string[] {
  if (!raw) return [];
  return Array.from(
    new Set(
      raw
        .split(",")
        .map((tag) => tag.trim().replace(/^#/, "").toLowerCase())
        .filter(Boolean),
    ),
  );
}
