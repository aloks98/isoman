import { Server } from 'lucide-react';
import { Link } from 'react-router';

export function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] gap-4">
      <Server className="w-16 h-16 text-muted-foreground/50" />
      <div className="text-center">
        <p className="text-lg font-medium">Page Not Found</p>
        <Link to="/isos" className="text-sm text-primary hover:underline">
          Go back to ISOs
        </Link>
      </div>
    </div>
  );
}
