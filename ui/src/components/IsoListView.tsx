import type { ISO } from '../types/iso';
import { Download, Trash2, RefreshCw, ExternalLink, CheckCircle, XCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface IsoListViewProps {
  isos: ISO[];
  onDelete: (id: string) => void;
  onRetry: (id: string) => void;
}

export function IsoListView({ isos, onDelete, onRetry }: IsoListViewProps) {
  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
  };

  const formatDate = (date: string): string => {
    return new Date(date).toLocaleDateString();
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'complete':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'failed':
        return <XCircle className="w-4 h-4 text-red-500" />;
      case 'downloading':
        return <Download className="w-4 h-4 text-blue-500 animate-pulse" />;
      default:
        return null;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'complete':
        return 'text-green-500';
      case 'failed':
        return 'text-red-500';
      case 'downloading':
        return 'text-blue-500';
      case 'verifying':
        return 'text-purple-500';
      default:
        return 'text-muted-foreground';
    }
  };

  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <table className="w-full">
        <thead className="bg-muted/50 border-b border-border">
          <tr>
            <th className="text-left px-4 py-3 text-sm font-medium">Status</th>
            <th className="text-left px-4 py-3 text-sm font-medium">Name</th>
            <th className="text-left px-4 py-3 text-sm font-medium">Version</th>
            <th className="text-left px-4 py-3 text-sm font-medium">Arch</th>
            <th className="text-left px-4 py-3 text-sm font-medium">Size</th>
            <th className="text-left px-4 py-3 text-sm font-medium">Progress</th>
            <th className="text-left px-4 py-3 text-sm font-medium">Created</th>
            <th className="text-right px-4 py-3 text-sm font-medium">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {isos.map((iso) => (
            <tr key={iso.id} className="hover:bg-accent/50 transition-colors">
              <td className="px-4 py-3">
                <div className="flex items-center gap-2">
                  {getStatusIcon(iso.status)}
                  <span className={`text-xs font-medium capitalize ${getStatusColor(iso.status)}`}>
                    {iso.status}
                  </span>
                </div>
              </td>
              <td className="px-4 py-3">
                <div>
                  <div className="font-medium">{iso.name}</div>
                  {iso.edition && <div className="text-xs text-muted-foreground">{iso.edition}</div>}
                </div>
              </td>
              <td className="px-4 py-3 font-mono text-sm">{iso.version}</td>
              <td className="px-4 py-3 font-mono text-sm">{iso.arch}</td>
              <td className="px-4 py-3 font-mono text-sm">
                {iso.size_bytes > 0 ? formatBytes(iso.size_bytes) : '-'}
              </td>
              <td className="px-4 py-3">
                {iso.status !== 'complete' && iso.status !== 'failed' && (
                  <div className="flex items-center gap-2">
                    <div className="w-24 h-1.5 bg-secondary rounded-full overflow-hidden">
                      <div
                        className={`h-full transition-all duration-300 ${
                          iso.status === 'downloading' ? 'bg-blue-500' : 'bg-purple-500'
                        }`}
                        style={{ width: `${iso.progress}%` }}
                      />
                    </div>
                    <span className="text-xs font-mono text-muted-foreground">{iso.progress}%</span>
                  </div>
                )}
              </td>
              <td className="px-4 py-3 font-mono text-xs text-muted-foreground">{formatDate(iso.created_at)}</td>
              <td className="px-4 py-3">
                <div className="flex items-center justify-end gap-1">
                  {iso.status === 'complete' && (
                    <Button asChild variant="ghost" mode="icon" size="sm">
                      <a href={iso.download_link} target="_blank" rel="noopener noreferrer" title="Download">
                        <Download className="w-4 h-4" />
                      </a>
                    </Button>
                  )}
                  {iso.status === 'failed' && (
                    <Button onClick={() => onRetry(iso.id)} variant="ghost" mode="icon" size="sm" title="Retry">
                      <RefreshCw className="w-4 h-4" />
                    </Button>
                  )}
                  <Button asChild variant="ghost" mode="icon" size="sm">
                    <a href={iso.download_url} target="_blank" rel="noopener noreferrer" title="View source">
                      <ExternalLink className="w-4 h-4" />
                    </a>
                  </Button>
                  <Button
                    onClick={() => onDelete(iso.id)}
                    variant="ghost"
                    mode="icon"
                    size="sm"
                    title="Delete"
                    className="text-destructive hover:text-destructive"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
