import { NextResponse } from "next/server";

import { callBackend, setSessionCookies } from "@/lib/api/session";

export async function POST(request: Request) {
  const body = await request.text();

  const response = await callBackend("/auth/login", { method: "POST", body });
  const payload = await response.json();

  if (!response.ok || !payload?.success) {
    return NextResponse.json(payload, { status: response.status });
  }

  const { access_token, refresh_token, expires_in, user } = payload.data;
  await setSessionCookies({ access_token, refresh_token, expires_in });

  return NextResponse.json({ success: true, data: { user } });
}
