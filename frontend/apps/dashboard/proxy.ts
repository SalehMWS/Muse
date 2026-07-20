import { NextResponse, type NextRequest } from "next/server";

const ACCESS_TOKEN_COOKIE = "novaflow_access_token";
const REFRESH_TOKEN_COOKIE = "novaflow_refresh_token";

const PUBLIC_PATHS = ["/", "/login", "/register", "/forgot-password", "/reset-password", "/terms", "/privacy"];

export function proxy(request: NextRequest) {
  const { pathname, search } = request.nextUrl;

  const hasSession =
    Boolean(request.cookies.get(ACCESS_TOKEN_COOKIE)) ||
    Boolean(request.cookies.get(REFRESH_TOKEN_COOKIE));

  const isPublic = PUBLIC_PATHS.includes(pathname);

  if (!hasSession && !isPublic) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("callbackUrl", `${pathname}${search}`);
    return NextResponse.redirect(loginUrl);
  }

  if (hasSession && (pathname === "/login" || pathname === "/register")) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)"],
};
