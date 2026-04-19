import { useEffect } from 'react';

const BASE_TITLE = 'ISOMan';

/**
 * Updates the browser tab title to reflect active download progress.
 * Shows "(45%) ISOMan" when downloads are in progress, or just "ISOMan" when idle.
 * Useful for homelabbers who monitor downloads across many tabs.
 */
export function useDocumentTitle(isos: { status: string; progress: number }[]) {
  useEffect(() => {
    const downloading = isos.filter((iso) => iso.status === 'downloading');

    if (downloading.length === 0) {
      document.title = BASE_TITLE;
      return;
    }

    // Average progress across all active downloads
    const avgProgress = Math.round(
      downloading.reduce((sum, iso) => sum + iso.progress, 0) /
        downloading.length,
    );

    document.title = `(${avgProgress}%) ${BASE_TITLE}`;

    return () => {
      document.title = BASE_TITLE;
    };
  }, [isos]);
}
