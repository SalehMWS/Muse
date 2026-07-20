import { NextResponse } from "next/server";

import { callBackend, clearSessionCookies, readRefreshToken } from "@/lib/api/session";

export async function POST() {
  const refreshToken = await readRefreshToken();

  if (refreshToken) {
    try {
      await callBackend("/auth/logout", {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
    } catch {
      // The local session is cleared regardless of what the backend reports.
    }
  }

  await clearSessionCookies();

  return NextResponse.json({ success: true, data: null });
}
