import { version } from '../../package.json';

export function Footer() {
  return (
    <footer className="border-t border-border bg-card text-card-foreground">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span>ISOMan - Linux ISO Manager</span>
          <span className="font-mono">v{version}</span>
        </div>
      </div>
    </footer>
  );
}
