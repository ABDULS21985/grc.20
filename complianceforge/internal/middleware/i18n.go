package middleware

import (
	"context"
	"net/http"
	"strings"

	i18nPkg "github.com/complianceforge/platform/internal/i18n"
)

const (
	// ContextKeyLanguage is the context key used to store the resolved language.
	ContextKeyLanguage contextKey = "language"

	// ContextKeyUserLang stores the user-profile language preference.
	ContextKeyUserLang contextKey = "user_language"
)

// I18n returns middleware that resolves the user's preferred language from
// the Accept-Language header, an explicit user preference already set in
// the context (e.g. from the JWT claims), and falls back to en-GB.
//
// Resolution order:
//  1. Accept-Language header (first supported match)
//  2. User profile preference (if present in context)
//  3. en-GB (default)
func I18n() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := resolveLanguage(r)
			ctx := context.WithValue(r.Context(), ContextKeyLanguage, lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// resolveLanguage determines the best language from the request.
func resolveLanguage(r *http.Request) string {
	// 1. Try Accept-Language header.
	if accept := r.Header.Get("Accept-Language"); accept != "" {
		if lang := parseAcceptLanguage(accept); lang != "" {
			return lang
		}
	}

	// 2. Try user preference already stored in context (e.g. from auth middleware).
	if pref, ok := r.Context().Value(ContextKeyUserLang).(string); ok && pref != "" {
		if i18nPkg.IsSupported(pref) {
			return pref
		}
	}

	// 3. Default.
	return "en-GB"
}

// parseAcceptLanguage parses the Accept-Language header and returns the first
// supported language. It handles both full tags (de-DE) and bare prefixes (de).
func parseAcceptLanguage(header string) string {
	supported := i18nPkg.SupportedLanguages()

	// Build a prefix map for bare-language matching (e.g. "de" -> "de-DE").
	prefixMap := make(map[string]string, len(supported))
	for _, lang := range supported {
		parts := strings.SplitN(lang, "-", 2)
		if len(parts) > 0 {
			prefix := strings.ToLower(parts[0])
			// First match wins — keeps the most specific variant.
			if _, exists := prefixMap[prefix]; !exists {
				prefixMap[prefix] = lang
			}
		}
	}

	// Parse comma-separated entries, ignoring quality values for simplicity.
	// Entries are ordered by preference already in most clients.
	entries := strings.Split(header, ",")
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		// Strip quality factor (e.g. ";q=0.8").
		if idx := strings.Index(entry, ";"); idx != -1 {
			entry = strings.TrimSpace(entry[:idx])
		}
		if entry == "" || entry == "*" {
			continue
		}

		// Exact match.
		for _, lang := range supported {
			if strings.EqualFold(entry, lang) {
				return lang
			}
		}

		// Prefix match (e.g. "de" matches "de-DE").
		prefix := strings.ToLower(strings.SplitN(entry, "-", 2)[0])
		if lang, ok := prefixMap[prefix]; ok {
			return lang
		}
	}

	return ""
}

// GetLanguage extracts the resolved language from the request context.
// Returns "en-GB" if not set.
func GetLanguage(ctx context.Context) string {
	if lang, ok := ctx.Value(ContextKeyLanguage).(string); ok && lang != "" {
		return lang
	}
	return "en-GB"
}

// SetUserLanguagePreference is a helper that stores a user's language
// preference in the context. Call this from auth middleware after loading
// the user profile.
func SetUserLanguagePreference(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, ContextKeyUserLang, lang)
}
