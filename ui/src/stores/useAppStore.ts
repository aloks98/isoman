import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AppStore {
  // Theme state
  isDark: boolean;
  toggleTheme: () => void;

  // View mode state
  viewMode: 'grid' | 'list';
  setViewMode: (mode: 'grid' | 'list') => void;

  // WebSocket state
  wsConnected: boolean;
  setWsConnected: (connected: boolean) => void;
}

export const useAppStore = create<AppStore>()(
  persist(
    (set) => ({
      // Theme state
      isDark: (() => {
        const stored = localStorage.getItem('darkMode');
        return stored ? stored === 'true' : true;
      })(),

      toggleTheme: () => {
        set((state) => {
          const newIsDark = !state.isDark;

          if (newIsDark) {
            document.documentElement.classList.add('dark');
          } else {
            document.documentElement.classList.remove('dark');
          }

          localStorage.setItem('darkMode', String(newIsDark));
          return { isDark: newIsDark };
        });
      },

      // View mode state
      viewMode: 'grid',
      setViewMode: (mode) => set({ viewMode: mode }),

      // WebSocket state
      wsConnected: false,
      setWsConnected: (connected) => set({ wsConnected: connected }),
    }),
    {
      name: 'isoman-app-store',
      partialize: (state) => ({
        isDark: state.isDark,
        viewMode: state.viewMode,
        // Don't persist wsConnected
      }),
    },
  ),
);

// Initialize theme on load
if (useAppStore.getState().isDark) {
  document.documentElement.classList.add('dark');
}
