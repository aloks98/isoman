import { Progress } from '@/components/ui/progress';
import { getStatusColor } from '@/lib/status-config';

interface ProgressBarProps {
  progress: number;
  status: string;
  className?: string;
}

export function ProgressBar({ progress, status, className = '' }: ProgressBarProps) {
  return (
    <div className={className}>
      <div className="flex justify-between items-center mb-2">
        <span className="text-sm text-muted-foreground capitalize">{status}</span>
        <span className="text-sm font-mono text-muted-foreground">{progress}%</span>
      </div>
      <Progress value={progress} indicatorClassName={getStatusColor(status)} />
    </div>
  );
}
