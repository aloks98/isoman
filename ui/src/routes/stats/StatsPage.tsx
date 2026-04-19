import { useQuery } from '@tanstack/react-query';
import {
  BarChart3,
  Download,
  HardDrive,
  Loader2,
  Package,
  TrendingUp,
} from 'lucide-react';
import { DistributionChart } from '@/components/stats/DistributionChart';
import { StatsCard } from '@/components/stats/StatsCard';
import { TopDownloadsTable } from '@/components/stats/TopDownloadsTable';
import { TrendsChart } from '@/components/stats/TrendsChart';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { getDownloadTrends, getStats } from '@/lib/api';
import { formatBytes } from '@/lib/format';

export function StatsPage() {
  const {
    data: stats,
    isLoading: statsLoading,
    error: statsError,
  } = useQuery({
    queryKey: ['stats'],
    queryFn: getStats,
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  const { data: trends, isLoading: trendsLoading } = useQuery({
    queryKey: ['stats', 'trends', 'daily'],
    queryFn: () => getDownloadTrends('daily', 30),
  });

  if (statsLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (statsError || !stats) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <p className="text-lg text-destructive">
            Could not load statistics
          </p>
          <p className="text-sm text-muted-foreground mt-2">
            Check that the backend is running, then refresh the page.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="animate-fade-in-up">
        <h1 className="text-2xl font-bold">Statistics</h1>
        <p className="text-muted-foreground mt-1">
          Overview of your ISO collection and download activity
        </p>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <div className="animate-fade-in-up" style={{ animationDelay: '0ms' }}>
          <StatsCard
            title="Total ISOs"
            value={stats.total_isos}
            icon={<Package className="h-4 w-4 text-blue-500" aria-hidden="true" />}
            description={`${stats.completed_isos} complete, ${stats.failed_isos} failed`}
          />
        </div>
        <div
          className="animate-fade-in-up"
          style={{ animationDelay: '75ms' }}
        >
          <StatsCard
            title="Total Storage"
            value={formatBytes(stats.total_size_bytes)}
            icon={<HardDrive className="h-4 w-4 text-amber-500" aria-hidden="true" />}
            description="Used by completed ISOs"
          />
        </div>
        <div
          className="animate-fade-in-up"
          style={{ animationDelay: '150ms' }}
        >
          <StatsCard
            title="Total Downloads"
            value={stats.total_downloads.toLocaleString()}
            icon={<Download className="h-4 w-4 text-emerald-500" aria-hidden="true" />}
            description="All-time download count"
          />
        </div>
        <div
          className="animate-fade-in-up"
          style={{ animationDelay: '225ms' }}
        >
          <StatsCard
            title="Bandwidth Saved"
            value={formatBytes(stats.bandwidth_saved)}
            icon={<TrendingUp className="h-4 w-4 text-cyan-500" aria-hidden="true" />}
            description="By caching ISOs locally"
          />
        </div>
      </div>

      {/* Charts Row */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5 text-blue-500" />
              ISOs by Architecture
            </CardTitle>
          </CardHeader>
          <CardContent>
            <DistributionChart data={stats.isos_by_arch} title="Architecture" />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5 text-emerald-500" />
              ISOs by Status
            </CardTitle>
          </CardHeader>
          <CardContent>
            <DistributionChart
              data={stats.isos_by_status}
              title="Status"
              colorByKey
            />
          </CardContent>
        </Card>
      </div>

      {/* Trends Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5 text-cyan-500" />
            Download Trends (Last 30 Days)
          </CardTitle>
        </CardHeader>
        <CardContent>
          {trendsLoading ? (
            <div className="flex items-center justify-center h-[300px]">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : trends ? (
            <TrendsChart data={trends.data} />
          ) : (
            <div className="text-muted-foreground text-center py-8">
              No trend data available
            </div>
          )}
        </CardContent>
      </Card>

      {/* Top Downloads Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Download className="h-5 w-5 text-amber-500" />
            Top Downloaded ISOs
          </CardTitle>
        </CardHeader>
        <CardContent>
          <TopDownloadsTable downloads={stats.top_downloaded} />
        </CardContent>
      </Card>

      {/* Edition Distribution (if any) */}
      {Object.keys(stats.isos_by_edition).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5 text-purple-500" />
              ISOs by Edition
            </CardTitle>
          </CardHeader>
          <CardContent>
            <DistributionChart data={stats.isos_by_edition} title="Edition" />
          </CardContent>
        </Card>
      )}
    </div>
  );
}
