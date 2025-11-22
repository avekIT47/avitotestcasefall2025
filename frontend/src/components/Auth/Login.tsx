import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { CodeBracketIcon, UserIcon, LockClosedIcon } from '@heroicons/react/24/outline';
import { toast } from 'react-hot-toast';
import Button from '../UI/Button';
import { useAuthStore, useThemeStore, useAppStore } from '../../store';
import { motion } from 'framer-motion';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const { login } = useAuthStore();
  const { isDark, toggleTheme } = useThemeStore();
  const { language, setLanguage } = useAppStore();
  const [credentials, setCredentials] = useState({
    username: '',
    password: '',
  });
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!credentials.username || !credentials.password) {
      toast.error('Please enter username and password');
      return;
    }

    setIsLoading(true);

    // Mock authentication - in production, this would call the API
    setTimeout(() => {
      // Accept any credentials for demo
      const mockUser = {
        id: 1,
        username: credentials.username,
        name: credentials.username.charAt(0).toUpperCase() + credentials.username.slice(1),
        role: 'admin',
      };
      
      // Mock token
      localStorage.setItem('authToken', 'mock-jwt-token');
      
      login(mockUser);
      toast.success('Login successful!');
      navigate('/');
      setIsLoading(false);
    }, 1000);
  };

  const handleLanguageToggle = () => {
    const newLang = language === 'ru' ? 'en' : 'ru';
    setLanguage(newLang);
    i18n.changeLanguage(newLang);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800 px-4">
      {/* Theme and language toggles */}
      <div className="absolute top-4 right-4 flex gap-2">
        <button
          onClick={handleLanguageToggle}
          className="px-4 py-2 rounded-lg bg-white dark:bg-gray-800 shadow-md hover:shadow-lg transition-shadow"
        >
          {language === 'ru' ? '🇬🇧 EN' : '🇷🇺 RU'}
        </button>
        <button
          onClick={toggleTheme}
          className="px-4 py-2 rounded-lg bg-white dark:bg-gray-800 shadow-md hover:shadow-lg transition-shadow"
        >
          {isDark ? '☀️' : '🌙'}
        </button>
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="w-full max-w-md"
      >
        <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8">
          {/* Logo and title */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary-100 dark:bg-primary-900/30 mb-4">
              <CodeBracketIcon className="h-8 w-8 text-primary-600 dark:text-primary-400" />
            </div>
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
              PR Reviewer
            </h1>
            <p className="mt-2 text-gray-600 dark:text-gray-400">
              {t('nav.login')} to your account
            </p>
          </div>

          {/* Login form */}
          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Username
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <UserIcon className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="text"
                  value={credentials.username}
                  onChange={(e) => setCredentials({ ...credentials, username: e.target.value })}
                  className="w-full pl-10 pr-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
                  placeholder="Enter username"
                  autoFocus
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Password
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <LockClosedIcon className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="password"
                  value={credentials.password}
                  onChange={(e) => setCredentials({ ...credentials, password: e.target.value })}
                  className="w-full pl-10 pr-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
                  placeholder="Enter password"
                />
              </div>
            </div>

            <div className="flex items-center justify-between">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
                />
                <span className="ml-2 text-sm text-gray-600 dark:text-gray-400">
                  Remember me
                </span>
              </label>
              <a href="#" className="text-sm text-primary-600 dark:text-primary-400 hover:underline">
                Forgot password?
              </a>
            </div>

            <Button
              type="submit"
              variant="primary"
              className="w-full"
              loading={isLoading}
            >
              {t('nav.login')}
            </Button>
          </form>

          {/* Demo note */}
          <div className="mt-6 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
            <p className="text-sm text-blue-700 dark:text-blue-400">
              <strong>Demo mode:</strong> Use any username and password to login
            </p>
          </div>
        </div>
      </motion.div>
    </div>
  );
};

export default Login;
