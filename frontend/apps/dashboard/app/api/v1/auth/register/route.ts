import { NextResponse } from "next/server";

import { callBackend, setSessionCookies } from "@/lib/api/session";

export async function POST(request: Request) {
  const body = await request.text();

  const registerResponse = await callBackend("/auth/register", { method: "POST", body });
  const registerPayload = await registerResponse.json();

  if (!registerResponse.ok || !registerPayload?.success) {
    return NextResponse.json(registerPayload, { status: registerResponse.status });
  }

  const { email, password } = JSON.parse(body) as { email: string; password: string };

  const loginResponse = await callBackend("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
  const loginPayload = await loginResponse.json();

  if (!loginResponse.ok || !loginPayload?.success) {
    return NextResponse.json(
      { success: true, data: { user: registerPayload.data, requiresLogin: true } },
      { status: 201 },
    );
  }

  const { access_token, refresh_token, expires_in, user } = loginPayload.data;
  await setSessionCookies({ access_token, refresh_token, expires_in });

  return NextResponse.json({ success: true, data: { user } }, { status: 201 });
}
