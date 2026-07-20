import type { Metadata } from "next";
import { Construction } from "lucide-react";

import { EmptyState } from "@/components/shared/empty-state";
import { PageHeader } from "@/components/shared/page-header";

export const metadata: Metadata = { title: "Settings | NovaFlow" };

export default function SettingsPage() {
  return (
    <>
      <PageHeader title="Settings" />
      <EmptyState
        icon={Construction}
        title="Not built yet"
        description="The settings API is live on the backend, but this screen has not been built. Content and authentication are the finished slices."
      />
    </>
  );
}
