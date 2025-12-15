/**
 * Aggregated statistics for the dashboard
 */
export interface Stats {
  total_isos: number;
  completed_isos: number;
  failed_isos: number;
  pending_isos: number;
  total_size_bytes: number;
  total_downloads: number;
  bandwidth_saved: number;
  isos_by_arch: Record<string, number>;
  isos_by_edition: Record<string, number>;
  isos_by_status: Record<string, number>;
  top_downloaded: ISODownloadStat[];
}

/**
 * Download statistics for a single ISO
 */
export interface ISODownloadStat {
  id: string;
  name: string;
  version: string;
  arch: string;
  download_count: number;
  size_bytes: number;
}

/**
 * Download trends over time
 */
export interface DownloadTrend {
  period: 'daily' | 'weekly';
  data: TrendDataPoint[];
}

/**
 * Single data point in a trend
 */
export interface TrendDataPoint {
  date: string;
  count: number;
}
