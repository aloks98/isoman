import { Outlet } from 'react-router';
import { Footer } from './Footer';
import { Header } from './Header';

export function MainLayout() {
  return (
    <>
      {/* Step 1: Main view - h-screen w-screen */}
      <div className="h-screen w-screen flex flex-col">
        {/* Step 2: Header with fixed height */}
        <Header />
        {/* Step 3: Scrollable main content - takes remaining height */}
        <main className="flex-1 overflow-y-auto">
          <div className="container mx-auto px-4 py-8">
            <Outlet />
          </div>
        </main>
      </div>
      {/* Step 4: Footer outside main view - revealed by page scroll */}
      <Footer />
    </>
  );
}
