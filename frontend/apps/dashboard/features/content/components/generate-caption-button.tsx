"use client";

import { Sparkles } from "lucide-react";
import { useId, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ApiError } from "@/lib/api/errors";

import { useGenerateCaption } from "../mutations/use-generate-caption";

export function GenerateCaptionButton({ contentId }: { contentId: string }) {
  const promptId = useId();
  const [prompt, setPrompt] = useState("");
  const generate = useGenerateCaption(contentId, () => setPrompt(""));

  const errorMessage = generate.isError
    ? generate.error instanceof ApiError
      ? generate.error.message
      : "Could not generate a caption."
    : null;

  return (
    <Card>
      <CardHeader>
        <CardTitle>AI caption</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={promptId}>Prompt</Label>
          <Input
            id={promptId}
            value={prompt}
            placeholder="Optional: steer the tone, for example: playful, short, no emoji"
            onChange={(event) => setPrompt(event.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            Leave this empty to caption from the content itself. Generating replaces the current
            caption and tags.
          </p>
        </div>

        <div>
          <Button
            loading={generate.isPending}
            onClick={() => generate.mutate(prompt.trim() || undefined)}
          >
            <Sparkles aria-hidden /> Generate caption
          </Button>
        </div>

        {errorMessage ? (
          <p role="alert" className="text-sm text-danger">
            {errorMessage}
          </p>
        ) : null}
      </CardContent>
    </Card>
  );
}
