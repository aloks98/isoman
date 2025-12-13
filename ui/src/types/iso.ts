/**
 * ISO model matching backend structure
 */
export interface ISO {
  id: string;
  name: string;
  version: string;
  arch: string;
  edition: string;
  file_type: string;
  filename: string;
  file_path: string;
  download_link: string;
  size_bytes: number;
  checksum: string;
  checksum_type: string;
  download_url: string;
  checksum_url: string;
  status: ISOStatus;
  progress: number;
  error_message: string;
  created_at: string;
  completed_at: string | null;
}

/**
 * ISO status enum matching backend
 */
export type ISOStatus = 'pending' | 'downloading' | 'verifying' | 'complete' | 'failed';

/**
 * Request payload for creating a new ISO download
 */
export interface CreateISORequest {
  name: string;
  version: string;
  arch: string;
  edition?: string;
  download_url: string;
  checksum_url?: string;
  checksum_type?: 'sha256' | 'sha512' | 'md5';
}

/**
 * Uniform API response structure
 */
export interface APIResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: APIError;
  message?: string;
}

/**
 * API error structure
 */
export interface APIError {
  code: string;
  message: string;
  details?: string;
}

/**
 * WebSocket message format for progress updates
 */
export interface WSProgressMessage {
  type: 'progress';
  payload: {
    id: string;
    progress: number;
    status: ISOStatus;
  };
}
