import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { ConnectedAccounts } from "@/features/analytics/components/connected-accounts";
import { ContentOverTimeChart } from "@/features/analytics/components/content-over-time-chart";
import { ContentStatusChart } from "@/features/analytics/components/content-status-chart";
import { OverviewStats } from "@/features/analytics/components/overview-stats";
import { WorkerHealth } from "@/features/analytics/components/worker-health";

export const metadata: Metadata = { title: "Analytics | NovaFlow" };

export default function AnalyticsPage() {
  return (
    <>
      <PageHeader
        title="Analytics"
        description="Derived from your content library, the background job queue and your connected accounts. Content figures are based on your 100 most recent items."
      />

      <div className="flex flex-col gap-8">
        <section className="flex flex-col gap-4" aria-labelledby="content-overview-heading">
          <h2 id="content-overview-heading" className="sr-only">
            Content overview
          </h2>
          <OverviewStats />
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <ConnectedAccounts />
          </div>
        </section>

        <section className="flex flex-col gap-4" aria-labelledby="content-breakdown-heading">
          <h2 id="content-breakdown-heading" className="text-lg font-semibold tracking-tight">
            Content breakdown
          </h2>
          <div className="grid gap-4 lg:grid-cols-2">
            <ContentStatusChart />
            <ContentOverTimeChart />
          </div>
        </section>

        <WorkerHealth />
      </div>
    </>
  );
}
