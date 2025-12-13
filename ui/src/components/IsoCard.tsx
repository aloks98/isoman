import {
  Check,
  Copy,
  Download,
  ExternalLink,
  MoreVertical,
  RefreshCw,
  Trash2,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useCopyWithFeedback } from '@/hooks/useCopyWithFeedback';
import { formatBytes, formatDate } from '@/lib/format';
import { getFullChecksumUrl, getFullDownloadUrl } from '@/lib/iso-utils';
import type { ISO } from '../types/iso';
import { ProgressBar } from './ProgressBar';
import { StatusBadge } from './StatusBadge';

interface IsoCardProps {
  iso: ISO;
  onDelete: (id: string) => void;
  onRetry: (id: string) => void;
}

export function IsoCard({ iso, onDelete, onRetry }: IsoCardProps) {
  const { copyToClipboard, copiedKey } = useCopyWithFeedback();
  const checksumUrl = getFullChecksumUrl(iso.download_link, iso.checksum_type);
  const downloadUrl = getFullDownloadUrl(iso.download_link);

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 py-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3 mb-2">
            <h3 className="text-lg font-semibold text-foreground">
              {iso.name}
            </h3>
            <StatusBadge status={iso.status} />
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
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" mode="icon" size="sm">
              <MoreVertical className="w-4 h-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem asChild>
              <a
                href={iso.download_url}
                target="_blank"
                rel="noopener noreferrer"
              >
                <ExternalLink className="w-4 h-4 mr-2" />
                View Source
              </a>
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => onDelete(iso.id)}
              className="text-destructive"
            >
              <Trash2 className="w-4 h-4 mr-2" />
              Delete ISO
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent>
        {iso.status !== 'complete' && iso.status !== 'failed' && (
          <div className="mb-4">
            <ProgressBar progress={iso.progress} status={iso.status} />
          </div>
        )}

        {iso.error_message && (
          <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md">
            <p className="text-sm text-destructive break-words">
              {iso.error_message}
            </p>
          </div>
        )}

        <div className="space-y-3 mb-4 text-sm">
          {iso.size_bytes > 0 && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Size:</span>
              <span className="font-mono">{formatBytes(iso.size_bytes)}</span>
            </div>
          )}
          {iso.checksum && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Checksum:</span>
              <span
                className="font-mono text-xs truncate max-w-[200px]"
                title={iso.checksum}
              >
                {iso.checksum.substring(0, 16)}...
              </span>
            </div>
          )}
          <div className="flex justify-between">
            <span className="text-muted-foreground">Created:</span>
            <span className="font-mono text-xs">
              {formatDate(iso.created_at)}
            </span>
          </div>
          {iso.completed_at && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Completed:</span>
              <span className="font-mono text-xs">
                {formatDate(iso.completed_at)}
              </span>
            </div>
          )}
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex gap-2">
            {iso.status === 'complete' && (
              <Button asChild className="flex-1">
                <a
                  href={iso.download_link}
                  target="_blank"
                  rel="noopener noreferrer"
                >
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
          </div>
          {iso.status === 'complete' && (
            <div className="flex gap-2">
              <Button
                onClick={() => copyToClipboard(downloadUrl, 'download')}
                variant="outline"
                className="flex-1"
              >
                {copiedKey === 'download' ? (
                  <>
                    <Check className="w-4 h-4" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4" />
                    Copy URL
                  </>
                )}
              </Button>
              {checksumUrl && (
                <Button
                  onClick={() => copyToClipboard(checksumUrl, 'checksum')}
                  variant="outline"
                  className="flex-1"
                >
                  {copiedKey === 'checksum' ? (
                    <>
                      <Check className="w-4 h-4" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="w-4 h-4" />
                      Copy Checksum URL
                    </>
                  )}
                </Button>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
