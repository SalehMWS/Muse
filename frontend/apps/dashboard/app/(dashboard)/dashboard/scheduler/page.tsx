import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { CreateScheduleForm } from "@/features/scheduler/components/create-schedule-form";
import { ScheduleList } from "@/features/scheduler/components/schedule-list";

export const metadata: Metadata = { title: "Scheduler | NovaFlow" };

export default function SchedulerPage() {
  return (
    <>
      <PageHeader
        title="Scheduler"
        description="Queue content for publishing at a fixed time or on a repeating cron schedule."
      />
      <div className="flex flex-col gap-8">
        <CreateScheduleForm />
        <ScheduleList />
      </div>
    </>
  );
}
