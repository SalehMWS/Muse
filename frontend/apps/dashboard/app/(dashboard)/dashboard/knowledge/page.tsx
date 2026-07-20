import type { Metadata } from "next";
import { Construction } from "lucide-react";

import { EmptyState } from "@/components/shared/empty-state";
import { PageHeader } from "@/components/shared/page-header";

export const metadata: Metadata = { title: "Knowledge | NovaFlow" };

export default function KnowledgePage() {
  return (
    <>
      <PageHeader title="Knowledge" />
      <EmptyState
        icon={Construction}
        title="Not built yet"
        description="The knowledge API is live on the backend, but this screen has not been built. Content and authentication are the finished slices."
      />
    </>
  );
}
