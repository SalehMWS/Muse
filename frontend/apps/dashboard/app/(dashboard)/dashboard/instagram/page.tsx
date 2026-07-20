import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { AccountList } from "@/features/instagram/components/account-list";
import { ConnectInstagramButton } from "@/features/instagram/components/connect-instagram-button";

export const metadata: Metadata = { title: "Instagram | NovaFlow" };

export default function InstagramPage() {
  return (
    <>
      <PageHeader
        title="Instagram"
        description="Connected accounts, token health and the OAuth connection to Meta."
        action={<ConnectInstagramButton />}
      />
      <AccountList />
    </>
  );
}
