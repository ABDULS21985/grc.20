package middleware

// CORS is handled by the go-chi/cors package in the router.
// This file documents the CORS configuration:
//
// Allowed Origins: Configured via APP_CORS_ORIGINS env variable
// Allowed Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
// Allowed Headers: Accept, Authorization, Content-Type, X-Request-ID
// Exposed Headers: Link, X-Total-Count, X-Request-ID
// Max Age: 300 seconds (5 minutes)
// Credentials: Allowed
//
// The CORS configuration is designed for European-facing SPAs and
// supports the authentication flow with JWT Bearer tokens.
