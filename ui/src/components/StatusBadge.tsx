import { Badge } from '@/components/ui/badge';
import { STATUS_CONFIG } from '@/lib/status-config';
import type { ISOStatus } from '../types/iso';

interface StatusBadgeProps {
  status: ISOStatus | string;
  className?: string;
}

/**
 * Reusable badge component for ISO status display
 * Automatically applies correct styling based on status using badge variants
 */
export function StatusBadge({ status, className = '' }: StatusBadgeProps) {
  // Handle known statuses
  if (status in STATUS_CONFIG) {
    const config = STATUS_CONFIG[status as ISOStatus];

    return (
      <Badge
        variant={config.badgeVariant}
        appearance={config.badgeAppearance}
        className={className}
      >
        {config.label}
      </Badge>
    );
  }

  // Fallback for unknown statuses
  return (
    <Badge variant="outline" className={className}>
      {status}
    </Badge>
  );
}
