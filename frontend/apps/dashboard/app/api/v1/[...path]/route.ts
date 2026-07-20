import { NextResponse } from "next/server";

import {
  callBackend,
  clearSessionCookies,
  readAccessToken,
  readRefreshToken,
  setSessionCookies,
} from "@/lib/api/session";

type RouteContext = { params: Promise<{ path: string[] }> };

async function proxy(request: Request, context: RouteContext): Promise<Response> {
  const { path } = await context.params;
  const target = `/${path.join("/")}`;
  const search = new URL(request.url).search;
  const body = request.method === "GET" || request.method === "HEAD" ? undefined : await request.text();

  const accessToken = await readAccessToken();

  let response = await callBackend(`${target}${search}`, {
    method: request.method,
    body,
    accessToken,
  });

  if (response.status === 401) {
    const refreshed = await refreshSession();

    if (!refreshed) {
      await clearSessionCookies();
      return NextResponse.json(
        { success: false, error: { code: "UNAUTHORIZED", message: "Your session has expired." } },
        { status: 401 },
      );
    }

    response = await callBackend(`${target}${search}`, {
      method: request.method,
      body,
      accessToken: refreshed,
    });
  }

  const payload = await response.text();
  const nullBodyStatus = response.status === 204 || response.status === 205 || response.status === 304;

  return new NextResponse(nullBodyStatus || payload === "" ? null : payload, {
    status: response.status,
    headers: { "Content-Type": response.headers.get("Content-Type") ?? "application/json" },
  });
}

async function refreshSession(): Promise<string | null> {
  const refreshToken = await readRefreshToken();
  if (!refreshToken) {
    return null;
  }

  const response = await callBackend("/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!response.ok) {
    return null;
  }

  const payload = await response.json();
  if (!payload?.success) {
    return null;
  }

  const { access_token, refresh_token, expires_in } = payload.data;
  await setSessionCookies({ access_token, refresh_token, expires_in });

  return access_token as string;
}

export { proxy as GET, proxy as POST, proxy as PUT, proxy as PATCH, proxy as DELETE };
