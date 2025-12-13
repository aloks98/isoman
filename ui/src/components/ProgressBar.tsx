import { Progress } from '@/components/ui/progress';

interface ProgressBarProps {
  progress: number;
  status: string;
  className?: string;
}

export function ProgressBar({ progress, status, className = '' }: ProgressBarProps) {
  const getStatusColor = () => {
    switch (status) {
      case 'downloading':
        return 'bg-blue-500';
      case 'verifying':
        return 'bg-purple-500';
      case 'complete':
        return 'bg-green-500';
      case 'failed':
        return 'bg-red-500';
      default:
        return 'bg-zinc-400';
    }
  };

  return (
    <div className={className}>
      <div className="flex justify-between items-center mb-2">
        <span className="text-sm text-muted-foreground capitalize">{status}</span>
        <span className="text-sm font-mono text-muted-foreground">{progress}%</span>
      </div>
      <Progress value={progress} indicatorClassName={getStatusColor()} />
    </div>
  );
}
