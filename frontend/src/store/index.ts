import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Migration: Clean up corrupted localStorage from old version
const migrateLocalStorage = () => {
  try {
    const appStorage = localStorage.getItem('app-storage');
    if (appStorage) {
      const data = JSON.parse(appStorage);
      // If selectedRows exists and is not properly serialized (empty object {}), remove it
      if (data.state && data.state.selectedRows && typeof data.state.selectedRows === 'object') {
        // Check if it's a corrupted Set (will be an empty object {})
        if (Object.keys(data.state.selectedRows).length === 0 && data.state.selectedRows.constructor === Object) {
          delete data.state.selectedRows;
          localStorage.setItem('app-storage', JSON.stringify(data));
        }
      }
    }
  } catch (error) {
    console.error('Migration error:', error);
  }
};

// Run migration on load
migrateLocalStorage();

interface ThemeState {
  isDark: boolean;
  toggleTheme: () => void;
}

interface AuthState {
  isAuthenticated: boolean;
  user: {
    id: number;
    username: string;
    name: string;
    role?: string;
  } | null;
  login: (user: AuthState['user']) => void;
  logout: () => void;
}

interface AppState {
  language: 'ru' | 'en';
  setLanguage: (lang: 'ru' | 'en') => void;
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
  selectedRows: Set<number>;
  setSelectedRows: (rows: Set<number>) => void;
  clearSelectedRows: () => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      isDark: false,
      toggleTheme: () => set((state) => {
        const newIsDark = !state.isDark;
        if (newIsDark) {
          document.documentElement.classList.add('dark');
        } else {
          document.documentElement.classList.remove('dark');
        }
        return { isDark: newIsDark };
      }),
    }),
    {
      name: 'theme-storage',
    }
  )
);

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      isAuthenticated: false,
      user: null,
      login: (user) => set({ isAuthenticated: true, user }),
      logout: () => {
        localStorage.removeItem('authToken');
        set({ isAuthenticated: false, user: null });
      },
    }),
    {
      name: 'auth-storage',
    }
  )
);

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      language: 'ru',
      setLanguage: (language) => set({ language }),
      sidebarCollapsed: false,
      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      selectedRows: new Set(),
      setSelectedRows: (rows) => set({ selectedRows: rows }),
      clearSelectedRows: () => set({ selectedRows: new Set() }),
    }),
    {
      name: 'app-storage',
      partialize: (state) => ({
        language: state.language,
        sidebarCollapsed: state.sidebarCollapsed,
        // Exclude selectedRows from persistence as Sets can't be serialized
      }),
    }
  )
);
