import React, { useState } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  HomeIcon,
  UsersIcon,
  UserGroupIcon,
  CodeBracketIcon,
  ChartBarIcon,
  Cog6ToothIcon,
  SunIcon,
  MoonIcon,
  Bars3Icon,
  XMarkIcon,
  GlobeAltIcon,
  ArrowLeftOnRectangleIcon,
} from '@heroicons/react/24/outline';
import { cn } from '../../utils';
import { useThemeStore, useAppStore, useAuthStore } from '../../store';
import { motion, AnimatePresence } from 'framer-motion';

const Layout: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const { isDark, toggleTheme } = useThemeStore();
  const { sidebarCollapsed, toggleSidebar, language, setLanguage } = useAppStore();
  const { isAuthenticated, user, logout } = useAuthStore();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const navigation = [
    { name: t('nav.dashboard'), href: '/', icon: HomeIcon },
    { name: t('nav.teams'), href: '/teams', icon: UserGroupIcon },
    { name: t('nav.users'), href: '/users', icon: UsersIcon },
    { name: t('nav.pullRequests'), href: '/pull-requests', icon: CodeBracketIcon },
    { name: t('nav.statistics'), href: '/statistics', icon: ChartBarIcon },
    { name: t('nav.settings'), href: '/settings', icon: Cog6ToothIcon },
  ];

  const handleLanguageChange = () => {
    const newLang = language === 'ru' ? 'en' : 'ru';
    setLanguage(newLang);
    i18n.changeLanguage(newLang);
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors">
      {/* Mobile menu button */}
      <div className="lg:hidden fixed top-4 left-4 z-50">
        <button
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          className="p-2 rounded-lg bg-white dark:bg-gray-800 shadow-lg"
        >
          {mobileMenuOpen ? (
            <XMarkIcon className="h-6 w-6 text-gray-600 dark:text-gray-300" />
          ) : (
            <Bars3Icon className="h-6 w-6 text-gray-600 dark:text-gray-300" />
          )}
        </button>
      </div>

      {/* Mobile menu overlay */}
      <AnimatePresence>
        {mobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/50 z-40 lg:hidden"
            onClick={() => setMobileMenuOpen(false)}
          />
        )}
      </AnimatePresence>

      {/* Sidebar */}
      <aside
        className={cn(
          'fixed left-0 top-0 h-full bg-white dark:bg-gray-800 shadow-xl transition-all duration-300 z-40',
          sidebarCollapsed && !mobileMenuOpen ? 'w-20' : 'w-64',
          mobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        )}
      >
        <div className="flex flex-col h-full">
          {/* Logo */}
          <div className="flex items-center justify-between h-16 px-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center">
              <CodeBracketIcon className="h-8 w-8 text-primary-600 dark:text-primary-400" />
              {(!sidebarCollapsed || mobileMenuOpen) && (
                <span className="ml-2 text-xl font-semibold text-gray-900 dark:text-white">
                  PR Reviewer
                </span>
              )}
            </div>
            <button
              onClick={toggleSidebar}
              className="hidden lg:block p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              <Bars3Icon className="h-5 w-5 text-gray-600 dark:text-gray-400" />
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-2 py-4 space-y-1">
            {navigation.map((item) => {
              const isActive = location.pathname === item.href;
              return (
                <Link
                  key={item.href}
                  to={item.href}
                  className={cn(
                    'flex items-center px-3 py-2 rounded-lg transition-colors',
                    isActive
                      ? 'bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  )}
                  onClick={() => setMobileMenuOpen(false)}
                >
                  <item.icon className="h-5 w-5 flex-shrink-0" />
                  {(!sidebarCollapsed || mobileMenuOpen) && (
                    <span className="ml-3">{item.name}</span>
                  )}
                </Link>
              );
            })}
          </nav>

          {/* Bottom actions */}
          <div className="border-t border-gray-200 dark:border-gray-700 p-4 space-y-2">
            {/* Theme toggle */}
            <button
              onClick={toggleTheme}
              className={cn(
                'flex items-center w-full px-3 py-2 rounded-lg',
                'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
              )}
            >
              {isDark ? (
                <SunIcon className="h-5 w-5" />
              ) : (
                <MoonIcon className="h-5 w-5" />
              )}
              {(!sidebarCollapsed || mobileMenuOpen) && (
                <span className="ml-3">
                  {isDark ? t('common.lightMode') : t('common.darkMode')}
                </span>
              )}
            </button>

            {/* Language toggle */}
            <button
              onClick={handleLanguageChange}
              className={cn(
                'flex items-center w-full px-3 py-2 rounded-lg',
                'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
              )}
            >
              <GlobeAltIcon className="h-5 w-5" />
              {(!sidebarCollapsed || mobileMenuOpen) && (
                <span className="ml-3">{language === 'ru' ? 'English' : 'Русский'}</span>
              )}
            </button>

            {/* User info / Logout */}
            {isAuthenticated && user && (
              <div className="border-t border-gray-200 dark:border-gray-700 pt-2 mt-2">
                {(!sidebarCollapsed || mobileMenuOpen) && (
                  <div className="px-3 py-2 text-sm text-gray-600 dark:text-gray-400">
                    {user.name}
                  </div>
                )}
                <button
                  onClick={handleLogout}
                  className={cn(
                    'flex items-center w-full px-3 py-2 rounded-lg',
                    'text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20'
                  )}
                >
                  <ArrowLeftOnRectangleIcon className="h-5 w-5" />
                  {(!sidebarCollapsed || mobileMenuOpen) && (
                    <span className="ml-3">{t('nav.logout')}</span>
                  )}
                </button>
              </div>
            )}
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div
        className={cn(
          'transition-all duration-300',
          sidebarCollapsed ? 'lg:ml-20' : 'lg:ml-64'
        )}
      >
        <main className="min-h-screen">
          <div className="p-4 sm:p-6 lg:p-8">
            <motion.div
              key={location.pathname}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3 }}
            >
              <Outlet />
            </motion.div>
          </div>
        </main>
      </div>
    </div>
  );
};

export default Layout;
