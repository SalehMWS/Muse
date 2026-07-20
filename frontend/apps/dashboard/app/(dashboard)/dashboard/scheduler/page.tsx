import type { Metadata } from "next";
import { Construction } from "lucide-react";

import { EmptyState } from "@/components/shared/empty-state";
import { PageHeader } from "@/components/shared/page-header";

export const metadata: Metadata = { title: "Scheduler | NovaFlow" };

export default function SchedulerPage() {
  return (
    <>
      <PageHeader title="Scheduler" />
      <EmptyState
        icon={Construction}
        title="Not built yet"
        description="The scheduler API is live on the backend, but this screen has not been built. Content and authentication are the finished slices."
      />
    </>
  );
}
