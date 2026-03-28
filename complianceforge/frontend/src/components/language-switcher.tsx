'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { Globe, ChevronDown, Check } from 'lucide-react';

// ── Supported languages ──────────────────────────────────
interface Language {
  code: string;
  label: string;
  nativeLabel: string;
  flag: string; // ISO 3166-1 country code for flag display
}

const SUPPORTED_LANGUAGES: Language[] = [
  { code: 'en-GB', label: 'English', nativeLabel: 'English', flag: 'GB' },
  { code: 'de-DE', label: 'German', nativeLabel: 'Deutsch', flag: 'DE' },
  { code: 'fr-FR', label: 'French', nativeLabel: 'Fran\u00e7ais', flag: 'FR' },
  { code: 'es-ES', label: 'Spanish', nativeLabel: 'Espa\u00f1ol', flag: 'ES' },
  { code: 'it-IT', label: 'Italian', nativeLabel: 'Italiano', flag: 'IT' },
  { code: 'nl-NL', label: 'Dutch', nativeLabel: 'Nederlands', flag: 'NL' },
  { code: 'pt-PT', label: 'Portuguese', nativeLabel: 'Portugu\u00eas', flag: 'PT' },
  { code: 'pl-PL', label: 'Polish', nativeLabel: 'Polski', flag: 'PL' },
  { code: 'sv-SE', label: 'Swedish', nativeLabel: 'Svenska', flag: 'SE' },
];

const DEFAULT_LANGUAGE = 'en-GB';
const STORAGE_KEY = 'complianceforge-language';

// ── Helper: get persisted language ───────────────────────
function getPersistedLanguage(): string {
  if (typeof window === 'undefined') return DEFAULT_LANGUAGE;
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored && SUPPORTED_LANGUAGES.some((l) => l.code === stored)) {
      return stored;
    }
  } catch {
    // localStorage may be unavailable (e.g. private browsing).
  }
  return DEFAULT_LANGUAGE;
}

function persistLanguage(code: string): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, code);
  } catch {
    // Silently ignore storage errors.
  }
}

// ── Props ────────────────────────────────────────────────
interface LanguageSwitcherProps {
  /** Called when the user selects a new language. */
  onLanguageChange?: (languageCode: string) => void;
  /** Override the current language (controlled mode). */
  currentLanguage?: string;
  /** Render as a compact icon-only button (for narrow spaces). */
  compact?: boolean;
  /** Additional CSS class names. */
  className?: string;
}

// ── Component ────────────────────────────────────────────
export function LanguageSwitcher({
  onLanguageChange,
  currentLanguage,
  compact = false,
  className = '',
}: LanguageSwitcherProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [internalLanguage, setInternalLanguage] = useState<string>(DEFAULT_LANGUAGE);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Resolve controlled vs uncontrolled language.
  const activeLanguage = currentLanguage ?? internalLanguage;

  // Hydrate from localStorage on mount.
  useEffect(() => {
    if (!currentLanguage) {
      setInternalLanguage(getPersistedLanguage());
    }
  }, [currentLanguage]);

  // Close dropdown when clicking outside.
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Close dropdown on Escape key.
  useEffect(() => {
    function handleEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') setIsOpen(false);
    }
    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => document.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen]);

  const handleSelect = useCallback(
    (code: string) => {
      setIsOpen(false);

      if (!currentLanguage) {
        setInternalLanguage(code);
      }
      persistLanguage(code);

      // Set the lang attribute on the document root element.
      document.documentElement.setAttribute('lang', code);

      onLanguageChange?.(code);
    },
    [currentLanguage, onLanguageChange],
  );

  const activeLang = SUPPORTED_LANGUAGES.find((l) => l.code === activeLanguage) ?? SUPPORTED_LANGUAGES[0];

  return (
    <div ref={dropdownRef} className={`relative inline-block text-left ${className}`}>
      {/* Trigger button */}
      <button
        type="button"
        onClick={() => setIsOpen((prev) => !prev)}
        className={`
          inline-flex items-center justify-center gap-2 rounded-lg border border-gray-200
          bg-white px-3 py-2 text-sm font-medium text-gray-700 shadow-sm
          transition-colors hover:bg-gray-50 focus:outline-none focus:ring-2
          focus:ring-blue-500 focus:ring-offset-1
          ${compact ? 'px-2' : ''}
        `}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        aria-label="Select language"
      >
        <Globe className="h-4 w-4 text-gray-500" />
        {!compact && (
          <>
            <span>{activeLang.nativeLabel}</span>
            <ChevronDown className={`h-3.5 w-3.5 text-gray-400 transition-transform ${isOpen ? 'rotate-180' : ''}`} />
          </>
        )}
      </button>

      {/* Dropdown */}
      {isOpen && (
        <div
          className="absolute right-0 z-50 mt-2 w-56 origin-top-right rounded-lg border border-gray-200 bg-white py-1 shadow-lg ring-1 ring-black/5 focus:outline-none"
          role="listbox"
          aria-label="Available languages"
        >
          {SUPPORTED_LANGUAGES.map((lang) => {
            const isActive = lang.code === activeLanguage;
            return (
              <button
                key={lang.code}
                type="button"
                role="option"
                aria-selected={isActive}
                onClick={() => handleSelect(lang.code)}
                className={`
                  flex w-full items-center gap-3 px-4 py-2.5 text-left text-sm transition-colors
                  ${isActive ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-50'}
                `}
              >
                {/* Country flag via regional indicator symbols */}
                <span className="text-base" role="img" aria-label={lang.label}>
                  {countryCodeToFlag(lang.flag)}
                </span>

                <span className="flex-1">
                  <span className="block font-medium">{lang.nativeLabel}</span>
                  <span className={`block text-xs ${isActive ? 'text-blue-500' : 'text-gray-400'}`}>
                    {lang.label} ({lang.code})
                  </span>
                </span>

                {isActive && <Check className="h-4 w-4 text-blue-600" />}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

// ── Utility: convert ISO 3166-1 alpha-2 to flag emoji ────
function countryCodeToFlag(code: string): string {
  const OFFSET = 0x1f1e6 - 65; // Regional Indicator 'A'
  return String.fromCodePoint(
    code.charCodeAt(0) + OFFSET,
    code.charCodeAt(1) + OFFSET,
  );
}
