import type { PaginationState, SortingState } from '@tanstack/react-table';
import {
  ChevronLeft,
  ChevronRight,
  LayoutGrid,
  List,
  Loader2,
  Server,
} from 'lucide-react';
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { CreateISORequest, ISO, PaginationInfo } from '../types/iso';
import { AddIsoForm } from './AddIsoForm';
import { IsoCard } from './IsoCard';
import { IsoListView } from './IsoListView';
import { WebSocketStatus } from './WebSocketStatus';

interface IsoListProps {
  isos: ISO[];
  isLoading: boolean;
  isFetching: boolean;
  error: Error | null;
  viewMode: 'grid' | 'list';
  onViewModeChange: (mode: 'grid' | 'list') => void;
  onCreateISO: (request: CreateISORequest) => Promise<void>;
  onDeleteISO: (id: string) => void;
  onRetryISO: (id: string) => void;
  onEditISO: (iso: ISO) => void;
  pagination: PaginationInfo;
  sorting: SortingState;
  onPaginationChange: (pagination: PaginationState) => void;
  onSortingChange: (sorting: SortingState) => void;
}

export function IsoList({
  isos,
  isLoading,
  isFetching,
  error,
  viewMode,
  onViewModeChange,
  onCreateISO,
  onDeleteISO,
  onRetryISO,
  onEditISO,
  pagination,
  sorting,
  onPaginationChange,
  onSortingChange,
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

      {pagination.total === 0 ? (
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
        <div className="space-y-4">
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
          {/* Pagination for grid view */}
          <GridPagination
            pagination={pagination}
            onPaginationChange={onPaginationChange}
          />
        </div>
      ) : (
        <IsoListView
          isos={isos}
          isFetching={isFetching}
          onDelete={handleDelete}
          onRetry={onRetryISO}
          onEdit={onEditISO}
          pagination={pagination}
          sorting={sorting}
          onPaginationChange={onPaginationChange}
          onSortingChange={onSortingChange}
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

interface GridPaginationProps {
  pagination: PaginationInfo;
  onPaginationChange: (pagination: PaginationState) => void;
}

function GridPagination({
  pagination,
  onPaginationChange,
}: GridPaginationProps) {
  const { page, page_size, total, total_pages } = pagination;

  const from = total === 0 ? 0 : (page - 1) * page_size + 1;
  const to = Math.min(page * page_size, total);

  const handlePageSizeChange = (value: string) => {
    onPaginationChange({
      pageIndex: 0, // Reset to first page when changing page size
      pageSize: Number(value),
    });
  };

  const handlePreviousPage = () => {
    if (page > 1) {
      onPaginationChange({
        pageIndex: page - 2, // Convert to 0-based
        pageSize: page_size,
      });
    }
  };

  const handleNextPage = () => {
    if (page < total_pages) {
      onPaginationChange({
        pageIndex: page, // Current page is already 1-based, so this is next page in 0-based
        pageSize: page_size,
      });
    }
  };

  if (total === 0) return null;

  return (
    <div className="flex items-center justify-between py-4 border-t border-border">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <span>Rows per page:</span>
        <Select value={String(page_size)} onValueChange={handlePageSizeChange}>
          <SelectTrigger className="w-[70px] h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {[10, 25, 50, 100].map((size) => (
              <SelectItem key={size} value={String(size)}>
                {size}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="flex items-center gap-4">
        <span className="text-sm text-muted-foreground">
          {from} - {to} of {total}
        </span>
        <div className="flex items-center gap-1">
          <Button
            variant="outline"
            mode="icon"
            size="sm"
            onClick={handlePreviousPage}
            disabled={page <= 1}
          >
            <ChevronLeft className="w-4 h-4" />
          </Button>
          <Button
            variant="outline"
            mode="icon"
            size="sm"
            onClick={handleNextPage}
            disabled={page >= total_pages}
          >
            <ChevronRight className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
