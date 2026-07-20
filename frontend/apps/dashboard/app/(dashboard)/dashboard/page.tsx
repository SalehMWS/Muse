import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { OverviewStats } from "@/features/analytics/components/overview-stats";

export const metadata: Metadata = { title: "Overview | NovaFlow" };

export default function DashboardPage() {
  return (
    <>
      <PageHeader title="Overview" description="A snapshot of your content pipeline." />
      <OverviewStats />
    </>
  );
}
