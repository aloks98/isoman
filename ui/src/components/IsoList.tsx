import { LayoutGrid, List, Loader2, Server } from 'lucide-react';
import { useState } from 'react';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import type { CreateISORequest, ISO } from '../types/iso';
import { AddIsoForm } from './AddIsoForm';
import { IsoCard } from './IsoCard';
import { IsoListView } from './IsoListView';
import { WebSocketStatus } from './WebSocketStatus';

interface IsoListProps {
  isos: ISO[];
  isLoading: boolean;
  error: Error | null;
  viewMode: 'grid' | 'list';
  onViewModeChange: (mode: 'grid' | 'list') => void;
  onCreateISO: (request: CreateISORequest) => Promise<void>;
  onDeleteISO: (id: string) => void;
  onRetryISO: (id: string) => void;
  onEditISO: (iso: ISO) => void;
}

export function IsoList({
  isos,
  isLoading,
  error,
  viewMode,
  onViewModeChange,
  onCreateISO,
  onDeleteISO,
  onRetryISO,
  onEditISO,
}: IsoListProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [isoToDelete, setIsoToDelete] = useState<ISO | null>(null);

  const handleDelete = (id: string) => {
    const iso = isos.find((i) => i.id === id);
    if (iso) {
      setIsoToDelete(iso);
      setDeleteDialogOpen(true);
    }
  };

  const confirmDelete = () => {
    if (!isoToDelete) return;
    onDeleteISO(isoToDelete.id);
    setDeleteDialogOpen(false);
    setIsoToDelete(null);
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
          <p className="text-destructive">
            {error instanceof Error ? error.message : 'Failed to load ISOs'}
          </p>
        </div>
      )}

      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">ISO Downloads</h2>
          <div className="flex items-center gap-3">
            <p className="text-muted-foreground">
              Manage your Linux ISO downloads
            </p>
            <WebSocketStatus />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center border border-border rounded-md">
            <Button
              onClick={() => onViewModeChange('grid')}
              variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
              mode="icon"
              size="sm"
              className="rounded-none rounded-l-md"
              title="Grid view"
            >
              <LayoutGrid className="w-4 h-4" />
            </Button>
            <Button
              onClick={() => onViewModeChange('list')}
              variant={viewMode === 'list' ? 'secondary' : 'ghost'}
              mode="icon"
              size="sm"
              className="rounded-none rounded-r-md"
              title="List view"
            >
              <List className="w-4 h-4" />
            </Button>
          </div>
          <AddIsoForm onSubmit={onCreateISO} />
        </div>
      </div>

      {isos.length === 0 ? (
        <div className="flex flex-col items-center justify-center min-h-[300px] gap-4 border-2 border-dashed border-border rounded-lg">
          <Server className="w-16 h-16 text-muted-foreground/50" />
          <div className="text-center">
            <p className="text-lg font-medium">No ISOs yet</p>
            <p className="text-sm text-muted-foreground">
              Add your first ISO download to get started
            </p>
          </div>
        </div>
      ) : viewMode === 'grid' ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {isos.map((iso) => (
            <IsoCard
              key={iso.id}
              iso={iso}
              onDelete={handleDelete}
              onRetry={onRetryISO}
              onEdit={onEditISO}
            />
          ))}
        </div>
      ) : (
        <IsoListView
          isos={isos}
          onDelete={handleDelete}
          onRetry={onRetryISO}
          onEdit={onEditISO}
        />
      )}

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete ISO?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete{' '}
              <strong>{isoToDelete?.name}</strong>? This action cannot be undone
              and will remove the ISO file from your server.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete}>
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
