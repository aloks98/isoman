import {
  type ColumnDef,
  getCoreRowModel,
  type OnChangeFn,
  type PaginationState,
  type SortingState,
  useReactTable,
} from '@tanstack/react-table';
import {
  Check,
  Copy,
  Download,
  Edit,
  ExternalLink,
  MoreVertical,
  RefreshCw,
  Trash2,
} from 'lucide-react';
import { useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { DataGrid, DataGridContainer } from '@/components/ui/data-grid';
import { DataGridColumnHeader } from '@/components/ui/data-grid-column-header';
import { DataGridPagination } from '@/components/ui/data-grid-pagination';
import { DataGridTable } from '@/components/ui/data-grid-table';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Skeleton } from '@/components/ui/skeleton';
import { useCopyWithFeedback } from '@/hooks/useCopyWithFeedback';
import { formatBytes, formatDateShort } from '@/lib/format';
import { getFullChecksumUrl, getFullDownloadUrl } from '@/lib/iso-utils';
import { getStatusColor } from '@/lib/status-config';
import type { ISO, PaginationInfo } from '../types/iso';
import { StatusBadge } from './StatusBadge';

interface IsoListViewProps {
  isos: ISO[];
  isFetching: boolean;
  onDelete: (id: string) => void;
  onRetry: (id: string) => void;
  onEdit: (iso: ISO) => void;
  pagination: PaginationInfo;
  sorting: SortingState;
  onPaginationChange: (pagination: PaginationState) => void;
  onSortingChange: (sorting: SortingState) => void;
}

export function IsoListView({
  isos,
  isFetching,
  onDelete,
  onRetry,
  onEdit,
  pagination,
  sorting,
  onPaginationChange,
  onSortingChange,
}: IsoListViewProps) {
  const { copyToClipboard, copiedKey } = useCopyWithFeedback();

  // Derive TanStack Table pagination state from server pagination (0-based pageIndex)
  const paginationState: PaginationState = {
    pageIndex: pagination.page - 1,
    pageSize: pagination.page_size,
  };

  // Handle pagination changes from DataGrid
  const handlePaginationChange: OnChangeFn<PaginationState> = (updater) => {
    const newState =
      typeof updater === 'function' ? updater(paginationState) : updater;
    onPaginationChange(newState);
  };

  // Handle sorting changes from DataGrid
  const handleSortingChange: OnChangeFn<SortingState> = (updater) => {
    const newState = typeof updater === 'function' ? updater(sorting) : updater;
    onSortingChange(newState);
  };

  const columns = useMemo<ColumnDef<ISO>[]>(
    () => [
      {
        accessorKey: 'name',
        header: ({ column }) => (
          <DataGridColumnHeader column={column} title="Name" />
        ),
        cell: ({ row }) => {
          const iso = row.original;
          return (
            <div>
              <div className="font-medium">{iso.name}</div>
              {iso.edition && (
                <div className="text-xs text-muted-foreground">
                  {iso.edition}
                </div>
              )}
            </div>
          );
        },
        meta: {
          skeleton: (
            <div className="space-y-1">
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-3 w-20" />
            </div>
          ),
        },
      },
      {
        accessorKey: 'status',
        header: ({ column }) => (
          <DataGridColumnHeader column={column} title="Status" />
        ),
        cell: ({ row }) => <StatusBadge status={row.original.status} />,
        meta: {
          skeleton: <Skeleton className="h-5 w-20 rounded-full" />,
        },
      },
      {
        accessorKey: 'version',
        header: ({ column }) => (
          <DataGridColumnHeader column={column} title="Version" />
        ),
        cell: ({ row }) => (
          <span className="font-mono text-sm">{row.original.version}</span>
        ),
        meta: {
          skeleton: <Skeleton className="h-4 w-16" />,
        },
      },
      {
        accessorKey: 'arch',
        header: 'Arch',
        enableSorting: false,
        cell: ({ row }) => (
          <span className="font-mono text-sm">{row.original.arch}</span>
        ),
        meta: {
          skeleton: <Skeleton className="h-4 w-14" />,
        },
      },
      {
        accessorKey: 'size_bytes',
        header: ({ column }) => (
          <DataGridColumnHeader column={column} title="Size" />
        ),
        cell: ({ row }) => (
          <span className="font-mono text-sm">
            {row.original.size_bytes > 0
              ? formatBytes(row.original.size_bytes)
              : '-'}
          </span>
        ),
        meta: {
          skeleton: <Skeleton className="h-4 w-16" />,
        },
      },
      {
        accessorKey: 'progress',
        header: 'Progress',
        enableSorting: false,
        cell: ({ row }) => {
          const iso = row.original;
          if (iso.status === 'complete' || iso.status === 'failed') return null;
          return (
            <div className="flex items-center gap-2">
              <div className="w-24 h-1.5 bg-secondary rounded-full overflow-hidden">
                <div
                  className={`h-full transition-all duration-300 ${getStatusColor(iso.status)}`}
                  style={{ width: `${iso.progress}%` }}
                />
              </div>
              <span className="text-xs font-mono text-muted-foreground">
                {iso.progress}%
              </span>
            </div>
          );
        },
        meta: {
          skeleton: (
            <div className="flex items-center gap-2">
              <Skeleton className="h-1.5 w-24 rounded-full" />
              <Skeleton className="h-3 w-8" />
            </div>
          ),
        },
      },
      {
        accessorKey: 'created_at',
        header: ({ column }) => (
          <DataGridColumnHeader column={column} title="Created" />
        ),
        cell: ({ row }) => (
          <span className="font-mono text-xs text-muted-foreground">
            {formatDateShort(row.original.created_at)}
          </span>
        ),
        meta: {
          skeleton: <Skeleton className="h-3 w-24" />,
        },
      },
      {
        id: 'actions',
        header: () => <div className="text-right">Actions</div>,
        enableSorting: false,
        cell: ({ row }) => {
          const iso = row.original;
          const checksumUrl = getFullChecksumUrl(
            iso.download_link,
            iso.checksum_type,
          );
          const downloadUrl = getFullDownloadUrl(iso.download_link);
          const copyKey = `${iso.id}`;

          return (
            <div className="flex items-center justify-end gap-1">
              {iso.status === 'complete' && (
                <Button asChild variant="ghost" mode="icon" size="sm">
                  <a
                    href={iso.download_link}
                    target="_blank"
                    rel="noopener noreferrer"
                    title="Download"
                  >
                    <Download className="w-4 h-4" />
                  </a>
                </Button>
              )}
              {iso.status === 'failed' && (
                <Button
                  onClick={() => onRetry(iso.id)}
                  variant="ghost"
                  mode="icon"
                  size="sm"
                  title="Retry"
                >
                  <RefreshCw className="w-4 h-4" />
                </Button>
              )}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" mode="icon" size="sm">
                    <MoreVertical className="w-4 h-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem
                    onClick={() =>
                      copyToClipboard(downloadUrl, `${copyKey}-download`)
                    }
                  >
                    {copiedKey === `${copyKey}-download` ? (
                      <Check className="w-4 h-4 mr-2 text-green-500" />
                    ) : (
                      <Copy className="w-4 h-4 mr-2" />
                    )}
                    Copy URL
                  </DropdownMenuItem>
                  {checksumUrl && (
                    <DropdownMenuItem
                      onClick={() =>
                        copyToClipboard(checksumUrl, `${copyKey}-checksum`)
                      }
                    >
                      {copiedKey === `${copyKey}-checksum` ? (
                        <Check className="w-4 h-4 mr-2 text-green-500" />
                      ) : (
                        <Copy className="w-4 h-4 mr-2" />
                      )}
                      Copy Checksum URL
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => onEdit(iso)}>
                    <Edit className="w-4 h-4 mr-2" />
                    Edit
                  </DropdownMenuItem>
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
            </div>
          );
        },
        meta: {
          headerClassName: 'text-right',
          skeleton: (
            <div className="flex items-center justify-end gap-1">
              <Skeleton className="h-8 w-8 rounded" />
              <Skeleton className="h-8 w-8 rounded" />
            </div>
          ),
        },
      },
    ],
    [copiedKey, onDelete, onRetry, onEdit, copyToClipboard],
  );

  const table = useReactTable({
    data: isos,
    columns,
    getRowId: (row) => row.id,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualSorting: true,
    pageCount: pagination.total_pages,
    state: {
      sorting,
      pagination: paginationState,
    },
    onSortingChange: handleSortingChange,
    onPaginationChange: handlePaginationChange,
  });

  return (
    <DataGridContainer>
      <DataGrid
        table={table}
        recordCount={pagination.total}
        isLoading={isFetching}
        loadingMode="skeleton"
        tableLayout={{
          rowBorder: true,
          headerBackground: true,
          width: 'fixed',
        }}
        emptyMessage="No ISOs found"
      >
        <DataGridTable />
        {pagination.total > 0 && (
          <div className="px-4 py-2 border-t border-border">
            <DataGridPagination sizes={[10, 25, 50, 100]} />
          </div>
        )}
      </DataGrid>
    </DataGridContainer>
  );
}
