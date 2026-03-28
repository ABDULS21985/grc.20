import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { format, formatDistanceToNow } from 'date-fns';

/**
 * Merge Tailwind CSS classes with clsx and tailwind-merge.
 * Handles conditional classes and resolves conflicts.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Format a date string as "dd MMM yyyy" (e.g. "28 Mar 2026").
 */
export function formatDate(date: string): string {
  return format(new Date(date), 'dd MMM yyyy');
}

/**
 * Format a date string as "dd MMM yyyy, HH:mm" (e.g. "28 Mar 2026, 14:30").
 */
export function formatDateTime(date: string): string {
  return format(new Date(date), 'dd MMM yyyy, HH:mm');
}

/**
 * Format a date string as a relative time (e.g. "3 hours ago", "in 2 days").
 */
export function formatRelativeTime(date: string): string {
  return formatDistanceToNow(new Date(date), { addSuffix: true });
}

/**
 * Format a number as currency using Intl.NumberFormat.
 * Defaults to EUR.
 */
export function formatCurrency(amount: number, currency = 'EUR'): string {
  return new Intl.NumberFormat('en-IE', {
    style: 'currency',
    currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

/**
 * Truncate a string to the given length, appending "..." if truncated.
 */
export function truncate(str: string, length: number): string {
  if (str.length <= length) return str;
  return str.slice(0, length) + '...';
}

/**
 * Get uppercase initials from a first and last name.
 */
export function getInitials(firstName: string, lastName: string): string {
  const first = firstName.charAt(0).toUpperCase();
  const last = lastName.charAt(0).toUpperCase();
  return `${first}${last}`;
}
