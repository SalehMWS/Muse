import Link from "next/link";
import { ArrowRight, CalendarClock, Sparkles, Library } from "lucide-react";

import { Button } from "@/components/ui/button";

const highlights = [
  {
    icon: Sparkles,
    title: "AI captions",
    body: "Generate on-brand captions and hashtags from a title or a prompt.",
  },
  {
    icon: CalendarClock,
    title: "Scheduled publishing",
    body: "Queue one-off or recurring posts and let the workers publish them.",
  },
  {
    icon: Library,
    title: "Knowledge base",
    body: "Ground every caption in your own documents with RAG retrieval.",
  },
];

export default function LandingPage() {
  return (
    <main className="mx-auto flex min-h-dvh max-w-5xl flex-col justify-center gap-16 px-6 py-20">
      <section className="flex flex-col gap-6">
        <span className="w-fit rounded-full border border-border px-3 py-1 text-xs font-medium text-muted-foreground">
          Instagram automation
        </span>
        <h1 className="max-w-2xl text-4xl font-semibold tracking-tight">
          Plan, generate and publish Instagram content from one place.
        </h1>
        <p className="max-w-xl text-base text-muted-foreground">
          NovaFlow connects your Instagram account, drafts captions with your own knowledge base, and
          publishes on a schedule through a queue that retries on failure.
        </p>
        <div className="flex flex-wrap gap-3">
          <Button asChild size="lg">
            <Link href="/register">
              Get started <ArrowRight />
            </Link>
          </Button>
          <Button asChild size="lg" variant="outline">
            <Link href="/login">Sign in</Link>
          </Button>
        </div>
      </section>

      <section className="grid gap-6 sm:grid-cols-3">
        {highlights.map(({ icon: Icon, title, body }) => (
          <div key={title} className="flex flex-col gap-2">
            <Icon className="size-5 text-primary" aria-hidden />
            <h2 className="text-sm font-semibold">{title}</h2>
            <p className="text-sm text-muted-foreground">{body}</p>
          </div>
        ))}
      </section>
    </main>
  );
}
