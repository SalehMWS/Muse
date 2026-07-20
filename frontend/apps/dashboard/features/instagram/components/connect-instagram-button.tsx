"use client";

import { Camera } from "lucide-react";

import { Button } from "@/components/ui/button";

import { useConnectInstagram } from "../mutations/use-connect-instagram";

export function ConnectInstagramButton({ label = "Connect account" }: { label?: string }) {
  const connect = useConnectInstagram();

  return (
    <Button loading={connect.isPending} onClick={() => connect.mutate()}>
      <Camera aria-hidden /> {label}
    </Button>
  );
}
