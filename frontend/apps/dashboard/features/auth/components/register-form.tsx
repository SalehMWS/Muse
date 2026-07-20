"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useForm } from "react-hook-form";

import { TextField } from "@/components/forms/text-field";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ApiError } from "@/lib/api/errors";

import { useRegister } from "../mutations/use-register";
import { registerSchema, type RegisterInput } from "../schemas/auth.schema";

export function RegisterForm() {
  const signUp = useRegister();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: { displayName: "", email: "", password: "" },
  });

  const submissionError = signUp.error instanceof ApiError ? signUp.error.message : null;
  const pending = isSubmitting || signUp.isPending;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create your account</CardTitle>
        <CardDescription>Start planning and publishing in a few minutes.</CardDescription>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit((values) => signUp.mutate(values))} noValidate>
          {submissionError ? (
            <p role="alert" className="rounded-md border border-danger/40 bg-danger/10 px-3 py-2 text-sm text-danger">
              {submissionError}
            </p>
          ) : null}

          <TextField
            label="Display name"
            autoComplete="name"
            required
            error={errors.displayName?.message}
            {...register("displayName")}
          />

          <TextField
            label="Email"
            type="email"
            autoComplete="email"
            required
            error={errors.email?.message}
            {...register("email")}
          />

          <TextField
            label="Password"
            type="password"
            autoComplete="new-password"
            required
            hint="At least 12 characters, with upper and lower case, a number and a symbol."
            error={errors.password?.message}
            {...register("password")}
          />

          <Button type="submit" loading={pending} className="mt-2">
            Create account
          </Button>

          <p className="text-center text-sm text-muted-foreground">
            Already registered?{" "}
            <Link href="/login" className="text-primary underline-offset-4 hover:underline">
              Sign in
            </Link>
          </p>
        </form>
      </CardContent>
    </Card>
  );
}
