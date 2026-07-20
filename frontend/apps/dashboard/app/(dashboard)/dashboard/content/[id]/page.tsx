import type { Metadata } from "next";
import Link from "next/link";

import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";
import { ContentDetail } from "@/features/content/components/content-detail";
import { GenerateCaptionButton } from "@/features/content/components/generate-caption-button";
import { MediaManager } from "@/features/content/components/media-manager";
import { PublishPanel } from "@/features/publishing/components/publish-panel";

export const metadata: Metadata = { title: "Content detail | NovaFlow" };

export default async function ContentDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <>
      <PageHeader
        title="Content detail"
        description="Edit the caption, attach media and publish to a connected account."
        action={
          <Button asChild variant="outline">
            <Link href="/dashboard/content">Back to content</Link>
          </Button>
        }
      />

      <div className="flex flex-col gap-6">
        <ContentDetail contentId={id} />
        <GenerateCaptionButton contentId={id} />
        <MediaManager contentId={id} />
        <PublishPanel contentId={id} />
      </div>
    </>
  );
}
