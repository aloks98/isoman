import './index.css';
import { IsoList } from './components/IsoList';
import { DarkModeToggle } from './components/DarkModeToggle';
import { HardDrive } from 'lucide-react';

const App = () => {
  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="border-b border-border bg-card text-card-foreground">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <HardDrive className="w-8 h-8 text-primary" />
              <div>
                <h1 className="text-2xl font-bold font-mono">ISO Manager</h1>
                <p className="text-sm text-muted-foreground">Download and manage Linux ISOs</p>
              </div>
            </div>
            <DarkModeToggle />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        <IsoList />
      </main>

      {/* Footer */}
      <footer className="border-t border-border mt-16 py-8">
        <div className="container mx-auto px-4 text-center text-sm text-muted-foreground font-mono">
          Built with Go + React + reui
        </div>
      </footer>
    </div>
  );
};

export default App;
