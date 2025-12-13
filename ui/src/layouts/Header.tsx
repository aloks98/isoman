import { FolderOpen, HardDrive } from 'lucide-react';
import { Link } from 'react-router';
import { DarkModeToggle } from '@/components/DarkModeToggle';
import { Button } from '@/components/ui/button';

export function Header() {
  return (
    <header className="border-b border-border bg-card text-card-foreground">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <Link
            to="/isos"
            className="flex items-center gap-3 hover:opacity-80 transition-opacity"
          >
            <HardDrive className="w-8 h-8 text-primary" />
            <div>
              <h1 className="text-2xl font-bold font-mono">ISOMan</h1>
              <p className="text-sm text-muted-foreground">
                Download and manage Linux ISOs
              </p>
            </div>
          </Link>
          <div className="flex items-center gap-2">
            <Button asChild variant="outline">
              <a href="/images/" target="_blank" rel="noopener noreferrer">
                <FolderOpen className="w-4 h-4" />
                Browse Files
              </a>
            </Button>
            <DarkModeToggle />
          </div>
        </div>
      </div>
    </header>
  );
}
