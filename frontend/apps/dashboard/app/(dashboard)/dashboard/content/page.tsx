import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { ContentList } from "@/features/content/components/content-list";
import { CreateContentForm } from "@/features/content/components/create-content-form";

export const metadata: Metadata = { title: "Content | NovaFlow" };

export default function ContentPage() {
  return (
    <>
      <PageHeader
        title="Content"
        description="Drafts, scheduled posts and everything already published."
        action={<CreateContentForm />}
      />
      <ContentList />
    </>
  );
}
