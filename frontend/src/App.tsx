import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { Toaster } from 'react-hot-toast';
import { I18nextProvider } from 'react-i18next';
import i18n from './i18n';
import Layout from './components/Layout/Layout';
import Dashboard from './components/Dashboard/Dashboard';
import Teams from './components/Teams/Teams';
import Users from './components/Users/Users';
import PullRequests from './components/PullRequests/PullRequests';
import Statistics from './components/Statistics/Statistics';
import Settings from './components/Settings/Settings';
import Login from './components/Auth/Login';
import { useThemeStore, useAuthStore, useAppStore } from './store';
import { registerServiceWorker } from './utils';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes
      retry: 3,
      refetchOnWindowFocus: false,
    },
  },
});

// Protected Route wrapper
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

function App() {
  const { isDark } = useThemeStore();
  const { language } = useAppStore();

  useEffect(() => {
    // Apply dark mode class on mount
    if (isDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [isDark]);

  useEffect(() => {
    // Change language
    i18n.changeLanguage(language);
  }, [language]);

  useEffect(() => {
    // Register service worker for PWA
    registerServiceWorker();
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <I18nextProvider i18n={i18n}>
        <BrowserRouter>
          <div className="App">
            <Routes>
              {/* Auth Routes */}
              <Route path="/login" element={<Login />} />
              
              {/* Protected Routes */}
              <Route
                path="/"
                element={
                  <ProtectedRoute>
                    <Layout />
                  </ProtectedRoute>
                }
              >
                <Route index element={<Dashboard />} />
                <Route path="teams" element={<Teams />} />
                <Route path="users" element={<Users />} />
                <Route path="pull-requests" element={<PullRequests />} />
                <Route path="statistics" element={<Statistics />} />
                <Route path="settings" element={<Settings />} />
              </Route>
              
              {/* Catch all */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
            
            {/* Toast notifications */}
            <Toaster
              position="top-right"
              toastOptions={{
                duration: 4000,
                style: {
                  background: isDark ? '#1F2937' : '#FFFFFF',
                  color: isDark ? '#F9FAFB' : '#111827',
                  border: `1px solid ${isDark ? '#374151' : '#E5E7EB'}`,
                },
              }}
            />
          </div>
        </BrowserRouter>
        <ReactQueryDevtools initialIsOpen={false} />
      </I18nextProvider>
    </QueryClientProvider>
  );
}

export default App;