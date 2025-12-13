import { useState, useEffect } from 'react';
import { IsoCard } from './IsoCard';
import { IsoListView } from './IsoListView';
import { AddIsoForm } from './AddIsoForm';
import { WebSocketStatus } from './WebSocketStatus';
import type { ISO, CreateISORequest, WSProgressMessage } from '../types/iso';
import { listISOs, createISO, deleteISO, retryISO } from '@/lib/api';
import { useWebSocket } from '@/hooks/useWebSocket';
import { Loader2, Server, LayoutGrid, List } from 'lucide-react';
import { Button } from '@/components/ui/button';

export function IsoList() {
  const [isos, setIsos] = useState<ISO[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  // Fetch ISOs on mount
  useEffect(() => {
    fetchISOs();
  }, []);

  const fetchISOs = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await listISOs();
      setIsos(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load ISOs');
    } finally {
      setIsLoading(false);
    }
  };

  // Handle WebSocket progress updates
  const handleWebSocketMessage = (message: WSProgressMessage) => {
    if (message.type === 'progress') {
      setIsos((prevIsos) =>
        prevIsos.map((iso) =>
          iso.id === message.payload.id
            ? { ...iso, progress: message.payload.progress, status: message.payload.status }
            : iso,
        ),
      );
    }
  };

  // Set up WebSocket connection
  const { isConnected } = useWebSocket({ onMessage: handleWebSocketMessage });

  const handleCreate = async (request: CreateISORequest) => {
    try {
      const newISO = await createISO(request);
      setIsos((prev) => [newISO, ...prev]);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create ISO');
      throw err;
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this ISO?')) return;

    try {
      await deleteISO(id);
      setIsos((prev) => prev.filter((iso) => iso.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete ISO');
    }
  };

  const handleRetry = async (id: string) => {
    try {
      const updatedISO = await retryISO(id);
      setIsos((prev) => prev.map((iso) => (iso.id === id ? updatedISO : iso)));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to retry ISO');
    }
  };

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] gap-4">
        <Loader2 className="w-12 h-12 animate-spin text-primary" />
        <p className="text-muted-foreground">Loading ISOs...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
          <p className="text-destructive">{error}</p>
        </div>
      )}

      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">ISO Downloads</h2>
          <div className="flex items-center gap-3">
            <p className="text-muted-foreground">Manage your Linux ISO downloads</p>
            <WebSocketStatus isConnected={isConnected} />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center border border-border rounded-md">
            <Button
              onClick={() => setViewMode('grid')}
              variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
              mode="icon"
              size="sm"
              className="rounded-none rounded-l-md"
              title="Grid view"
            >
              <LayoutGrid className="w-4 h-4" />
            </Button>
            <Button
              onClick={() => setViewMode('list')}
              variant={viewMode === 'list' ? 'secondary' : 'ghost'}
              mode="icon"
              size="sm"
              className="rounded-none rounded-r-md"
              title="List view"
            >
              <List className="w-4 h-4" />
            </Button>
          </div>
          <AddIsoForm onSubmit={handleCreate} />
        </div>
      </div>

      {isos.length === 0 ? (
        <div className="flex flex-col items-center justify-center min-h-[300px] gap-4 border-2 border-dashed border-border rounded-lg">
          <Server className="w-16 h-16 text-muted-foreground/50" />
          <div className="text-center">
            <p className="text-lg font-medium">No ISOs yet</p>
            <p className="text-sm text-muted-foreground">Add your first ISO download to get started</p>
          </div>
        </div>
      ) : viewMode === 'grid' ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {isos.map((iso) => (
            <IsoCard key={iso.id} iso={iso} onDelete={handleDelete} onRetry={handleRetry} />
          ))}
        </div>
      ) : (
        <IsoListView isos={isos} onDelete={handleDelete} onRetry={handleRetry} />
      )}
    </div>
  );
}
