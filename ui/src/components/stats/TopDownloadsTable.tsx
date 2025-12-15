import {
  type ColumnDef,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { useMemo } from 'react';
import { DataGrid, DataGridContainer } from '@/components/ui/data-grid';
import { DataGridTable } from '@/components/ui/data-grid-table';
import { formatBytes } from '@/lib/format';
import type { ISODownloadStat } from '@/types/stats';

interface TopDownloadsTableProps {
  downloads: ISODownloadStat[];
}

export function TopDownloadsTable({ downloads }: TopDownloadsTableProps) {
  const columns = useMemo<ColumnDef<ISODownloadStat>[]>(
    () => [
      {
        accessorKey: 'name',
        header: 'ISO',
        cell: ({ row }) => (
          <span className="font-medium">{row.original.name}</span>
        ),
      },
      {
        accessorKey: 'version',
        header: 'Version',
        cell: ({ row }) => (
          <span className="font-mono text-sm text-muted-foreground">
            {row.original.version}
          </span>
        ),
      },
      {
        accessorKey: 'arch',
        header: 'Arch',
        cell: ({ row }) => (
          <span className="font-mono text-sm text-muted-foreground">
            {row.original.arch}
          </span>
        ),
      },
      {
        accessorKey: 'size_bytes',
        header: 'Size',
        cell: ({ row }) => (
          <span className="font-mono text-sm text-muted-foreground">
            {formatBytes(row.original.size_bytes)}
          </span>
        ),
        meta: {
          headerClassName: 'text-right',
          cellClassName: 'text-right',
        },
      },
      {
        accessorKey: 'download_count',
        header: 'Downloads',
        cell: ({ row }) => (
          <span className="font-bold text-primary">
            {row.original.download_count.toLocaleString()}
          </span>
        ),
        meta: {
          headerClassName: 'text-right',
          cellClassName: 'text-right',
        },
      },
    ],
    [],
  );

  const table = useReactTable({
    data: downloads,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getRowId: (row) => row.id,
  });

  if (downloads.length === 0) {
    return (
      <div className="text-muted-foreground text-center py-8">
        No downloads recorded yet
      </div>
    );
  }

  return (
    <DataGridContainer>
      <DataGrid
        table={table}
        recordCount={downloads.length}
        tableLayout={{
          rowBorder: true,
          headerBackground: true,
          width: 'fixed',
        }}
        emptyMessage="No downloads recorded yet"
      >
        <DataGridTable />
      </DataGrid>
    </DataGridContainer>
  );
}
