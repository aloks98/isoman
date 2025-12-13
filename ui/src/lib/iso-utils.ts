/**
 * Constructs the checksum download URL for an ISO
 * @param downloadLink - ISO download link path
 * @param checksumType - Checksum type (sha256, sha512, md5)
 * @returns Full checksum URL or null if no checksum type
 */
export function getChecksumUrl(
  downloadLink: string,
  checksumType: string | null | undefined
): string | null {
  return checksumType ? `${downloadLink}.${checksumType}` : null;
}

/**
 * Gets the full download URL including origin
 * @param downloadLink - Relative download link path
 * @returns Full URL with origin
 */
export function getFullDownloadUrl(downloadLink: string): string {
  return `${window.location.origin}${downloadLink}`;
}

/**
 * Gets the full checksum URL including origin
 * @param downloadLink - Relative download link path
 * @param checksumType - Checksum type
 * @returns Full checksum URL with origin or null
 */
export function getFullChecksumUrl(
  downloadLink: string,
  checksumType: string | null | undefined
): string | null {
  const checksumPath = getChecksumUrl(downloadLink, checksumType);
  return checksumPath ? getFullDownloadUrl(checksumPath) : null;
}
