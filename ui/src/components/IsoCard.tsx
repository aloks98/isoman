import { ProgressBar } from './ProgressBar';
import type { ISO } from '../types/iso';
import { Download, Trash2, RefreshCw, ExternalLink, CheckCircle, XCircle } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

interface IsoCardProps {
  iso: ISO;
  onDelete: (id: string) => void;
  onRetry: (id: string) => void;
}

export function IsoCard({ iso, onDelete, onRetry }: IsoCardProps) {
  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
  };

  const formatDate = (date: string): string => {
    return new Date(date).toLocaleString();
  };

  const getStatusIcon = () => {
    switch (iso.status) {
      case 'complete':
        return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'failed':
        return <XCircle className="w-5 h-5 text-red-500" />;
      case 'downloading':
        return <Download className="w-5 h-5 text-blue-500 animate-pulse" />;
      default:
        return null;
    }
  };

  return (
    <Card>
      <CardContent>
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              {getStatusIcon()}
              <h3 className="text-lg font-semibold text-foreground">{iso.name}</h3>
            </div>
            <div className="flex gap-2 text-sm text-muted-foreground font-mono">
              <span>{iso.version}</span>
              <span>•</span>
              <span>{iso.arch}</span>
              {iso.edition && (
                <>
                  <span>•</span>
                  <span>{iso.edition}</span>
                </>
              )}
              <span>•</span>
              <span className="uppercase">{iso.file_type}</span>
            </div>
          </div>
        </div>

        {iso.status !== 'complete' && iso.status !== 'failed' && (
          <div className="mb-4">
            <ProgressBar progress={iso.progress} status={iso.status} />
          </div>
        )}

        {iso.error_message && (
          <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md">
            <p className="text-sm text-destructive">{iso.error_message}</p>
          </div>
        )}

        <div className="space-y-2 mb-4 text-sm">
          {iso.size_bytes > 0 && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Size:</span>
              <span className="font-mono">{formatBytes(iso.size_bytes)}</span>
            </div>
          )}
          {iso.checksum && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Checksum:</span>
              <span className="font-mono text-xs truncate max-w-[200px]" title={iso.checksum}>
                {iso.checksum.substring(0, 16)}...
              </span>
            </div>
          )}
          <div className="flex justify-between">
            <span className="text-muted-foreground">Created:</span>
            <span className="font-mono text-xs">{formatDate(iso.created_at)}</span>
          </div>
          {iso.completed_at && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Completed:</span>
              <span className="font-mono text-xs">{formatDate(iso.completed_at)}</span>
            </div>
          )}
        </div>

        <div className="flex gap-2">
          {iso.status === 'complete' && (
            <Button asChild className="flex-1">
              <a href={iso.download_link} target="_blank" rel="noopener noreferrer">
                <Download />
                Download
              </a>
            </Button>
          )}
          {iso.status === 'failed' && (
            <Button onClick={() => onRetry(iso.id)} className="flex-1">
              <RefreshCw />
              Retry
            </Button>
          )}
          <Button onClick={() => onDelete(iso.id)} variant="destructive" mode="icon">
            <Trash2 />
          </Button>
          <Button asChild variant="outline" mode="icon">
            <a href={iso.download_url} target="_blank" rel="noopener noreferrer" title="View source">
              <ExternalLink />
            </a>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
