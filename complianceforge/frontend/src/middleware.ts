import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

/**
 * Protected routes that require authentication.
 * Users without a valid cf_token cookie will be redirected to /auth/login.
 */
const PROTECTED_ROUTES = [
  '/dashboard',
  '/frameworks',
  '/risks',
  '/policies',
  '/audits',
  '/incidents',
  '/vendors',
  '/assets',
  '/settings',
  '/controls',
  '/reports',
];

/**
 * Routes that are always accessible without authentication.
 */
const PUBLIC_PREFIXES = [
  '/auth',
  '/api',
  '/_next',
  '/favicon.ico',
];

function isPublicPath(pathname: string): boolean {
  return PUBLIC_PREFIXES.some((prefix) => pathname.startsWith(prefix));
}

function isProtectedPath(pathname: string): boolean {
  return PROTECTED_ROUTES.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Allow public routes without any auth check
  if (isPublicPath(pathname)) {
    return NextResponse.next();
  }

  // Check for authentication token on protected routes
  if (isProtectedPath(pathname)) {
    const token = request.cookies.get('cf_token')?.value;

    if (!token) {
      const loginUrl = new URL('/auth/login', request.url);
      loginUrl.searchParams.set('redirect', pathname);
      return NextResponse.redirect(loginUrl);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except static files and images.
     * This ensures middleware runs on page navigations but not on
     * static asset requests for performance.
     */
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};
