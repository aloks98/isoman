import { format } from 'date-fns';

/**
 * Converts bytes to human-readable format
 * @param bytes - Number of bytes
 * @returns Formatted string (e.g., "1.23 GB")
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / k ** i).toFixed(2)} ${sizes[i]}`;
}

/**
 * Formats an ISO 8601 date string to localized date and time
 * @param date - ISO 8601 date string
 * @returns Formatted date string (e.g., "Jan 15, 2025, 10:30 AM")
 */
export function formatDate(date: string): string {
  return format(new Date(date), 'PPp'); // "Jan 15, 2025, 10:30 AM"
}

/**
 * Formats date to short format (date only, no time)
 * @param date - ISO 8601 date string
 * @returns Formatted date string (e.g., "Jan 15, 2025")
 */
export function formatDateShort(date: string): string {
  return format(new Date(date), 'PP'); // "Jan 15, 2025"
}
