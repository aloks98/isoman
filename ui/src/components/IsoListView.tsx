import { useMemo } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  flexRender,
  type ColumnDef,
  type SortingState,
  type ColumnFiltersState,
} from '@tanstack/react-table';
import type { ISO } from '../types/iso';
import { Download, Trash2, RefreshCw, ExternalLink, Copy, Check, MoreVertical, ArrowUpDown } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import { formatBytes, formatDateShort } from '@/lib/format';
import { getFullDownloadUrl, getFullChecksumUrl } from '@/lib/iso-utils';
import { StatusBadge } from './StatusBadge';
import { useCopyWithFeedback } from '@/hooks/useCopyWithFeedback';
import { getStatusColor } from '@/lib/status-config';
import { useState } from 'react';

interface IsoListViewProps {
  isos: ISO[];
  onDelete: (id: string) => void;
  onRetry: (id: string) => void;
}

export function IsoListView({ isos, onDelete, onRetry }: IsoListViewProps) {
  const { copyToClipboard, copiedKey } = useCopyWithFeedback();
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

  const columns = useMemo<ColumnDef<ISO>[]>(
    () => [
      {
        accessorKey: 'name',
        header: ({ column }) => {
          return (
            <Button
              variant="ghost"
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
              className="h-8 px-2 -ml-2"
            >
              Name
              <ArrowUpDown className="ml-2 h-4 w-4" />
            </Button>
          );
        },
        cell: ({ row }) => {
          const iso = row.original;
          return (
            <div>
              <div className="font-medium">{iso.name}</div>
              {iso.edition && <div className="text-xs text-muted-foreground">{iso.edition}</div>}
            </div>
          );
        },
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => <StatusBadge status={row.original.status} />,
      },
      {
        accessorKey: 'version',
        header: ({ column }) => {
          return (
            <Button
              variant="ghost"
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
              className="h-8 px-2 -ml-2"
            >
              Version
              <ArrowUpDown className="ml-2 h-4 w-4" />
            </Button>
          );
        },
        cell: ({ row }) => <span className="font-mono text-sm">{row.original.version}</span>,
      },
      {
        accessorKey: 'arch',
        header: 'Arch',
        cell: ({ row }) => <span className="font-mono text-sm">{row.original.arch}</span>,
      },
      {
        accessorKey: 'size_bytes',
        header: ({ column }) => {
          return (
            <Button
              variant="ghost"
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
              className="h-8 px-2 -ml-2"
            >
              Size
              <ArrowUpDown className="ml-2 h-4 w-4" />
            </Button>
          );
        },
        cell: ({ row }) => (
          <span className="font-mono text-sm">
            {row.original.size_bytes > 0 ? formatBytes(row.original.size_bytes) : '-'}
          </span>
        ),
      },
      {
        accessorKey: 'progress',
        header: 'Progress',
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
              <span className="text-xs font-mono text-muted-foreground">{iso.progress}%</span>
            </div>
          );
        },
      },
      {
        accessorKey: 'created_at',
        header: ({ column }) => {
          return (
            <Button
              variant="ghost"
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
              className="h-8 px-2 -ml-2"
            >
              Created
              <ArrowUpDown className="ml-2 h-4 w-4" />
            </Button>
          );
        },
        cell: ({ row }) => (
          <span className="font-mono text-xs text-muted-foreground">{formatDateShort(row.original.created_at)}</span>
        ),
      },
      {
        id: 'actions',
        header: () => <div className="text-right">Actions</div>,
        cell: ({ row }) => {
          const iso = row.original;
          const checksumUrl = getFullChecksumUrl(iso.download_link, iso.checksum_type);
          const downloadUrl = getFullDownloadUrl(iso.download_link);
          const copyKey = `${iso.id}`;

          return (
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
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" mode="icon" size="sm">
                    <MoreVertical className="w-4 h-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem
                    onClick={() => copyToClipboard(downloadUrl, `${copyKey}-download`)}
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
                      onClick={() => copyToClipboard(checksumUrl, `${copyKey}-checksum`)}
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
                  <DropdownMenuItem asChild>
                    <a href={iso.download_url} target="_blank" rel="noopener noreferrer">
                      <ExternalLink className="w-4 h-4 mr-2" />
                      View Source
                    </a>
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => onDelete(iso.id)} className="text-destructive">
                    <Trash2 className="w-4 h-4 mr-2" />
                    Delete ISO
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          );
        },
      },
    ],
    [copiedKey, onDelete, onRetry, copyToClipboard]
  );

  const table = useReactTable({
    data: isos,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    state: {
      sorting,
      columnFilters,
    },
    initialState: {
      pagination: {
        pageSize: 10,
      },
    },
  });

  return (
    <div className="space-y-4">
      <div className="border border-border rounded-lg overflow-hidden">
        <table className="w-full">
          <thead className="bg-muted/50 border-b border-border">
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th key={header.id} className="text-left px-4 py-3 text-sm font-medium">
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody className="divide-y divide-border">
            {table.getRowModel().rows.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="px-4 py-8 text-center text-muted-foreground">
                  No ISOs found
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="hover:bg-accent/50 transition-colors">
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="px-4 py-3">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {table.getPageCount() > 1 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            Showing {table.getState().pagination.pageIndex * table.getState().pagination.pageSize + 1} to{' '}
            {Math.min(
              (table.getState().pagination.pageIndex + 1) * table.getState().pagination.pageSize,
              table.getFilteredRowModel().rows.length
            )}{' '}
            of {table.getFilteredRowModel().rows.length} entries
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.previousPage()}
              disabled={!table.getCanPreviousPage()}
            >
              Previous
            </Button>
            <Button variant="outline" size="sm" onClick={() => table.nextPage()} disabled={!table.getCanNextPage()}>
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
