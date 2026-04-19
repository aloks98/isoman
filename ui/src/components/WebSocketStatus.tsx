import { WifiOff } from 'lucide-react';
import { useAppStore } from '@/stores';

export function WebSocketStatus() {
  const isConnected = useAppStore((state) => state.wsConnected);

  return (
    <div className="flex items-center gap-2 text-xs" role="status" aria-live="polite">
      {isConnected ? (
        <>
          <span className="relative flex h-2 w-2" aria-hidden="true">
            <span className="absolute inline-flex h-full w-full rounded-full bg-success opacity-75 animate-ping" />
            <span className="relative inline-flex h-2 w-2 rounded-full bg-success" />
          </span>
          <span className="text-muted-foreground">Live updates</span>
        </>
      ) : (
        <>
          <WifiOff className="w-3.5 h-3.5 text-destructive" />
          <span className="text-destructive">Disconnected</span>
        </>
      )}
    </div>
  );
}
