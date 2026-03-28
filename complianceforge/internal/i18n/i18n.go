// Package i18n provides internationalisation support for the ComplianceForge platform.
// It loads embedded JSON translation files, supports interpolation, pluralisation,
// and locale-aware formatting for dates, numbers and currencies.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

//go:embed locales/*.json
var localeFS embed.FS

// supportedLanguages lists all languages the platform supports.
var supportedLanguages = []string{
	"en-GB", "de-DE", "fr-FR", "es-ES", "it-IT",
	"nl-NL", "pt-PT", "pl-PL", "sv-SE",
}

// defaultLanguage is the fallback language when a requested key is missing.
const defaultLanguage = "en-GB"

// langFileMap maps BCP-47 language tags to their JSON file names.
var langFileMap = map[string]string{
	"en-GB": "locales/en.json",
	"de-DE": "locales/de.json",
	"fr-FR": "locales/fr.json",
	"es-ES": "locales/es.json",
	"it-IT": "locales/it.json",
	"nl-NL": "locales/nl.json",
	"pt-PT": "locales/pt.json",
	"pl-PL": "locales/pl.json",
	"sv-SE": "locales/sv.json",
}

// dateFormats stores locale-specific date/time format strings.
var dateFormats = map[string]string{
	"de-DE": "02.01.2006 15:04",
	"fr-FR": "02/01/2006 15h04",
	"en-GB": "02/01/2006 15:04",
	"es-ES": "02/01/2006 15:04",
	"it-IT": "02/01/2006 15:04",
	"nl-NL": "02-01-2006 15:04",
	"pt-PT": "02/01/2006 15:04",
	"pl-PL": "02.01.2006 15:04",
	"sv-SE": "2006-01-02 15:04",
}

// numberFormat stores decimal and thousands separators per locale.
type numberFormat struct {
	Decimal   string
	Thousands string
}

var numberFormats = map[string]numberFormat{
	"en-GB": {Decimal: ".", Thousands: ","},
	"de-DE": {Decimal: ",", Thousands: "."},
	"fr-FR": {Decimal: ",", Thousands: "\u202f"}, // narrow no-break space
	"es-ES": {Decimal: ",", Thousands: "."},
	"it-IT": {Decimal: ",", Thousands: "."},
	"nl-NL": {Decimal: ",", Thousands: "."},
	"pt-PT": {Decimal: ",", Thousands: "."},
	"pl-PL": {Decimal: ",", Thousands: "\u00a0"}, // no-break space
	"sv-SE": {Decimal: ",", Thousands: "\u00a0"}, // no-break space
}

// currencySymbols maps ISO-4217 codes to their symbols.
var currencySymbols = map[string]string{
	"EUR": "\u20ac",
	"GBP": "\u00a3",
	"USD": "$",
	"SEK": "kr",
	"PLN": "z\u0142",
	"CHF": "CHF",
}

// Translator holds all loaded translations and provides lookup methods.
type Translator struct {
	// translations maps lang -> flat-dotted-key -> translated string
	translations map[string]map[string]string
}

var (
	globalTranslator *Translator
	once             sync.Once
)

// Global returns the singleton Translator, initialised on first call.
func Global() *Translator {
	once.Do(func() {
		t, err := newTranslator()
		if err != nil {
			log.Fatal().Err(err).Msg("i18n: failed to load translations")
		}
		globalTranslator = t
	})
	return globalTranslator
}

// newTranslator loads all embedded JSON locale files into memory.
func newTranslator() (*Translator, error) {
	t := &Translator{
		translations: make(map[string]map[string]string),
	}

	for lang, file := range langFileMap {
		data, err := localeFS.ReadFile(file)
		if err != nil {
			log.Warn().Str("lang", lang).Str("file", file).Err(err).Msg("i18n: could not read locale file")
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			log.Warn().Str("lang", lang).Err(err).Msg("i18n: invalid JSON in locale file")
			continue
		}

		flat := make(map[string]string)
		flatten("", raw, flat)
		t.translations[lang] = flat
		log.Debug().Str("lang", lang).Int("keys", len(flat)).Msg("i18n: loaded locale")
	}

	if _, ok := t.translations[defaultLanguage]; !ok {
		return nil, fmt.Errorf("i18n: default language %s not loaded", defaultLanguage)
	}

	return t, nil
}

// flatten recursively converts nested JSON into dot-separated keys.
func flatten(prefix string, src map[string]interface{}, dst map[string]string) {
	for k, v := range src {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			dst[key] = val
		case map[string]interface{}:
			flatten(key, val, dst)
		default:
			// Numbers or booleans — store as string representation.
			dst[key] = fmt.Sprintf("%v", val)
		}
	}
}

// SupportedLanguages returns the list of supported BCP-47 language tags.
func SupportedLanguages() []string {
	out := make([]string, len(supportedLanguages))
	copy(out, supportedLanguages)
	return out
}

// IsSupported returns true if the given language tag is supported.
func IsSupported(lang string) bool {
	for _, l := range supportedLanguages {
		if strings.EqualFold(l, lang) {
			return true
		}
	}
	return false
}

// T translates a key for the given language. If the key is not found in the
// requested language, it falls back to en-GB. Interpolation placeholders of
// the form {{name}} are replaced from the optional args map.
func (t *Translator) T(lang, key string, args ...map[string]interface{}) string {
	result := t.lookup(lang, key)
	if len(args) > 0 && args[0] != nil {
		result = interpolate(result, args[0])
	}
	return result
}

// lookup finds a translation by language and key with fallback.
func (t *Translator) lookup(lang, key string) string {
	if msgs, ok := t.translations[lang]; ok {
		if val, ok := msgs[key]; ok {
			return val
		}
	}

	// Fallback to default language.
	if lang != defaultLanguage {
		if msgs, ok := t.translations[defaultLanguage]; ok {
			if val, ok := msgs[key]; ok {
				log.Debug().Str("lang", lang).Str("key", key).Msg("i18n: key missing, fell back to en-GB")
				return val
			}
		}
	}

	log.Warn().Str("lang", lang).Str("key", key).Msg("i18n: missing translation key")
	return key
}

// interpolate replaces {{placeholder}} tokens in the string with values from args.
func interpolate(s string, args map[string]interface{}) string {
	for k, v := range args {
		placeholder := "{{" + k + "}}"
		s = strings.ReplaceAll(s, placeholder, fmt.Sprintf("%v", v))
	}
	return s
}

// TPl handles pluralisation. It looks up key.one or key.other based on count,
// and injects {{count}} into the interpolation args automatically.
func (t *Translator) TPl(lang, key string, count int, args ...map[string]interface{}) string {
	suffix := ".other"
	if count == 1 {
		suffix = ".one"
	}

	merged := map[string]interface{}{"count": count}
	if len(args) > 0 && args[0] != nil {
		for k, v := range args[0] {
			merged[k] = v
		}
	}

	return t.T(lang, key+suffix, merged)
}

// FormatDate formats a time.Time value according to the locale's date format.
func FormatDate(lang string, t time.Time) string {
	if fmt, ok := dateFormats[lang]; ok {
		return t.Format(fmt)
	}
	return t.Format(dateFormats[defaultLanguage])
}

// FormatNumber formats a float64 with locale-appropriate separators.
// It uses two decimal places by default.
func FormatNumber(lang string, n float64) string {
	nf, ok := numberFormats[lang]
	if !ok {
		nf = numberFormats[defaultLanguage]
	}

	negative := n < 0
	if negative {
		n = -n
	}

	// Split into integer and fractional parts.
	intPart := int64(math.Floor(n))
	fracPart := int64(math.Round((n - float64(intPart)) * 100))

	// Format the integer part with thousands separators.
	intStr := formatIntWithSep(intPart, nf.Thousands)

	result := intStr + nf.Decimal + fmt.Sprintf("%02d", fracPart)
	if negative {
		result = "-" + result
	}
	return result
}

// formatIntWithSep inserts thousands separators into an integer string.
func formatIntWithSep(n int64, sep string) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var b strings.Builder
	start := len(s) % 3
	if start == 0 {
		start = 3
	}
	b.WriteString(s[:start])
	for i := start; i < len(s); i += 3 {
		b.WriteString(sep)
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// FormatCurrency formats a monetary amount with the locale's number format
// and prepends or appends the currency symbol.
func FormatCurrency(lang string, amount float64, currency string) string {
	formatted := FormatNumber(lang, amount)
	symbol, ok := currencySymbols[strings.ToUpper(currency)]
	if !ok {
		symbol = currency
	}

	// Symbol placement varies by locale.
	switch lang {
	case "en-GB":
		return symbol + formatted
	default:
		// Most European locales place the symbol after the number.
		return formatted + "\u00a0" + symbol
	}
}
