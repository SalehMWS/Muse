import type { Metadata } from "next";
import { Construction } from "lucide-react";

import { EmptyState } from "@/components/shared/empty-state";
import { PageHeader } from "@/components/shared/page-header";

export const metadata: Metadata = { title: "Instagram | NovaFlow" };

export default function InstagramPage() {
  return (
    <>
      <PageHeader title="Instagram" />
      <EmptyState
        icon={Construction}
        title="Not built yet"
        description="The instagram API is live on the backend, but this screen has not been built. Content and authentication are the finished slices."
      />
    </>
  );
}
