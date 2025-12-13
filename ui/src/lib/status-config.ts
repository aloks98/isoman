import type { ISOStatus } from '../types/iso';
import type { BadgeProps } from '@/components/ui/badge';

export interface StatusConfig {
  label: string;
  badgeVariant: BadgeProps['variant'];
  badgeAppearance: BadgeProps['appearance'];
  progressColor: string;
}

/**
 * Centralized status configuration for ISO states
 * Maps status to badge variants and progress bar colors
 */
export const STATUS_CONFIG: Record<ISOStatus, StatusConfig> = {
  complete: {
    label: 'Complete',
    badgeVariant: 'success',
    badgeAppearance: 'light',
    progressColor: 'bg-green-500',
  },
  failed: {
    label: 'Failed',
    badgeVariant: 'destructive',
    badgeAppearance: 'light',
    progressColor: 'bg-red-500',
  },
  downloading: {
    label: 'Downloading',
    badgeVariant: 'info',
    badgeAppearance: 'light',
    progressColor: 'bg-blue-500',
  },
  verifying: {
    label: 'Verifying',
    badgeVariant: 'warning',
    badgeAppearance: 'light',
    progressColor: 'bg-purple-500',
  },
  pending: {
    label: 'Pending',
    badgeVariant: 'secondary',
    badgeAppearance: 'light',
    progressColor: 'bg-zinc-400',
  },
};

/**
 * Gets the progress bar color class for a status
 * @param status - ISO status
 * @returns Tailwind CSS class name for background color
 */
export function getStatusColor(status: ISOStatus | string): string {
  if (status in STATUS_CONFIG) {
    return STATUS_CONFIG[status as ISOStatus].progressColor;
  }
  return 'bg-zinc-400'; // Default color
}
