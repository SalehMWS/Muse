import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { AppearanceCard } from "@/features/settings/components/appearance-card";
import { ProfileCard } from "@/features/settings/components/profile-card";
import { SessionCard } from "@/features/settings/components/session-card";

export const metadata: Metadata = { title: "Settings | NovaFlow" };

export default function SettingsPage() {
  return (
    <>
      <PageHeader
        title="Settings"
        description="Your account details and how NovaFlow looks on this device."
      />

      <div className="flex max-w-2xl flex-col gap-6">
        <ProfileCard />
        <AppearanceCard />
        <SessionCard />
      </div>
    </>
  );
}
