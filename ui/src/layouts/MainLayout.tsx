import { Outlet } from 'react-router';
import { Header } from './Header';

export function MainLayout() {
  return (
    <div className="min-h-screen">
      <Header />
      <main className="container mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}
