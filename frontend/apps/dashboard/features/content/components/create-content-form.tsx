"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Plus } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { TextField } from "@/components/forms/text-field";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";

import { useCreateContent } from "../mutations/use-create-content";
import { createContentSchema, type CreateContentInput } from "../schemas/create-content.schema";

export function CreateContentForm() {
  const [open, setOpen] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateContentInput>({
    resolver: zodResolver(createContentSchema),
    defaultValues: { title: "", caption: "", contentType: "image", language: "en", tags: "" },
  });

  const create = useCreateContent(() => {
    reset();
    setOpen(false);
  });

  if (!open) {
    return (
      <Button onClick={() => setOpen(true)}>
        <Plus /> New content
      </Button>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>New content</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          className="flex flex-col gap-4"
          onSubmit={handleSubmit((values) => create.mutate(values))}
          noValidate
        >
          <TextField label="Title" required error={errors.title?.message} {...register("title")} />

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="content-type">Type</Label>
            <select
              id="content-type"
              className="h-10 rounded-md border border-border bg-background px-3 text-sm"
              {...register("contentType")}
            >
              <option value="image">Image</option>
              <option value="carousel">Carousel</option>
              <option value="reel">Reel</option>
              <option value="story">Story</option>
            </select>
          </div>

          <TextField
            label="Tags"
            hint="Comma separated, for example: sunset, travel"
            error={errors.tags?.message}
            {...register("tags")}
          />

          <div className="flex gap-2">
            <Button type="submit" loading={create.isPending}>
              Create
            </Button>
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                reset();
                setOpen(false);
              }}
            >
              Cancel
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
