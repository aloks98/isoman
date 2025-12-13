import { useState, useEffect } from 'react';
import { useCopyToClipboard } from 'usehooks-ts';

/**
 * Enhanced clipboard hook with automatic feedback reset
 * Wraps usehooks-ts useCopyToClipboard with auto-reset timer
 *
 * @param resetDelay - Time in milliseconds before resetting copied state (default: 2000)
 * @returns Object with copyToClipboard function and copiedKey state
 */
export function useCopyWithFeedback(resetDelay = 2000) {
  const [, copy] = useCopyToClipboard();
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  /**
   * Copy text to clipboard and track which item was copied
   * @param text - Text to copy
   * @param key - Unique identifier for the copied item (for UI feedback)
   */
  const copyToClipboard = async (text: string, key: string) => {
    await copy(text);
    setCopiedKey(key);
  };

  // Auto-reset copied state after delay
  useEffect(() => {
    if (copiedKey) {
      const timer = setTimeout(() => setCopiedKey(null), resetDelay);
      return () => clearTimeout(timer);
    }
  }, [copiedKey, resetDelay]);

  return {
    copyToClipboard,
    copiedKey,
  };
}
