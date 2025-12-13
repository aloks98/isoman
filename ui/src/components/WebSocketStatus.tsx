import { Wifi, WifiOff } from 'lucide-react';

interface WebSocketStatusProps {
  isConnected: boolean;
}

export function WebSocketStatus({ isConnected }: WebSocketStatusProps) {
  return (
    <div className="flex items-center gap-2 text-xs">
      {isConnected ? (
        <>
          <Wifi className="w-4 h-4 text-green-500" />
          <span className="text-muted-foreground">Live</span>
        </>
      ) : (
        <>
          <WifiOff className="w-4 h-4 text-destructive" />
          <span className="text-destructive">Disconnected</span>
        </>
      )}
    </div>
  );
}
