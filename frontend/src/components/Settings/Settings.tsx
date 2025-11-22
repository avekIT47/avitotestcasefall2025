import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  BellIcon,
  GlobeAltIcon,
  MoonIcon,
  ShieldCheckIcon,
  CpuChipIcon,
  DocumentArrowDownIcon,
} from '@heroicons/react/24/outline';
import { toast } from 'react-hot-toast';
import Button from '../UI/Button';
import { useThemeStore, useAppStore } from '../../store';
import { requestNotificationPermission } from '../../utils';
import { motion } from 'framer-motion';

const Settings: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { isDark, toggleTheme } = useThemeStore();
  const { language, setLanguage } = useAppStore();
  const [notifications, setNotifications] = useState({
    enabled: false,
    prCreated: true,
    prMerged: true,
    reviewerAssigned: true,
  });
  const [apiUrl, setApiUrl] = useState(import.meta.env.VITE_API_URL || 'http://localhost:8080');

  const handleEnableNotifications = async () => {
    const granted = await requestNotificationPermission();
    if (granted) {
      setNotifications({ ...notifications, enabled: true });
      toast.success('Notifications enabled');
    } else {
      toast.error('Notifications permission denied');
    }
  };

  const handleLanguageChange = (newLang: 'ru' | 'en') => {
    setLanguage(newLang);
    i18n.changeLanguage(newLang);
    toast.success(`Language changed to ${newLang === 'ru' ? 'Russian' : 'English'}`);
  };

  const handleExportSettings = () => {
    const settings = {
      theme: isDark ? 'dark' : 'light',
      language,
      notifications,
      apiUrl,
    };
    const blob = new Blob([JSON.stringify(settings, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'pr-reviewer-settings.json';
    a.click();
    URL.revokeObjectURL(url);
    toast.success('Settings exported');
  };

  const handleImportSettings = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      try {
        const settings = JSON.parse(event.target?.result as string);
        
        // Apply settings
        if (settings.theme === 'dark' && !isDark) toggleTheme();
        if (settings.theme === 'light' && isDark) toggleTheme();
        if (settings.language) handleLanguageChange(settings.language);
        if (settings.notifications) setNotifications(settings.notifications);
        if (settings.apiUrl) setApiUrl(settings.apiUrl);
        
        toast.success('Settings imported successfully');
      } catch (error) {
        toast.error('Failed to import settings');
      }
    };
    reader.readAsText(file);
  };

  const settingSections = [
    {
      title: 'Appearance',
      icon: MoonIcon,
      settings: [
        {
          label: 'Theme',
          description: 'Choose between light and dark mode',
          control: (
            <div className="flex gap-2">
              <Button
                size="sm"
                variant={!isDark ? 'primary' : 'ghost'}
                onClick={() => isDark && toggleTheme()}
              >
                Light
              </Button>
              <Button
                size="sm"
                variant={isDark ? 'primary' : 'ghost'}
                onClick={() => !isDark && toggleTheme()}
              >
                Dark
              </Button>
            </div>
          ),
        },
      ],
    },
    {
      title: 'Language',
      icon: GlobeAltIcon,
      settings: [
        {
          label: 'Interface Language',
          description: 'Choose your preferred language',
          control: (
            <select
              value={language}
              onChange={(e) => handleLanguageChange(e.target.value as 'ru' | 'en')}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="ru">Русский</option>
              <option value="en">English</option>
            </select>
          ),
        },
      ],
    },
    {
      title: 'Notifications',
      icon: BellIcon,
      settings: [
        {
          label: 'Enable Notifications',
          description: 'Receive browser notifications for important events',
          control: (
            <Button
              size="sm"
              variant={notifications.enabled ? 'primary' : 'secondary'}
              onClick={handleEnableNotifications}
            >
              {notifications.enabled ? 'Enabled' : 'Enable'}
            </Button>
          ),
        },
        {
          label: 'PR Created',
          description: 'Notify when a new PR is created',
          control: (
            <input
              type="checkbox"
              checked={notifications.prCreated}
              onChange={(e) => setNotifications({ ...notifications, prCreated: e.target.checked })}
              disabled={!notifications.enabled}
              className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
            />
          ),
        },
        {
          label: 'PR Merged',
          description: 'Notify when a PR is merged',
          control: (
            <input
              type="checkbox"
              checked={notifications.prMerged}
              onChange={(e) => setNotifications({ ...notifications, prMerged: e.target.checked })}
              disabled={!notifications.enabled}
              className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
            />
          ),
        },
        {
          label: 'Reviewer Assigned',
          description: 'Notify when you are assigned as a reviewer',
          control: (
            <input
              type="checkbox"
              checked={notifications.reviewerAssigned}
              onChange={(e) => setNotifications({ ...notifications, reviewerAssigned: e.target.checked })}
              disabled={!notifications.enabled}
              className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
            />
          ),
        },
      ],
    },
    {
      title: 'API Configuration',
      icon: CpuChipIcon,
      settings: [
        {
          label: 'API URL',
          description: 'Backend server URL',
          control: (
            <input
              type="text"
              value={apiUrl}
              onChange={(e) => setApiUrl(e.target.value)}
              className="w-64 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          ),
        },
      ],
    },
    {
      title: 'Data Management',
      icon: DocumentArrowDownIcon,
      settings: [
        {
          label: 'Export Settings',
          description: 'Download your settings as a JSON file',
          control: (
            <Button size="sm" variant="secondary" onClick={handleExportSettings}>
              Export
            </Button>
          ),
        },
        {
          label: 'Import Settings',
          description: 'Import settings from a JSON file',
          control: (
            <>
              <input
                type="file"
                accept=".json"
                onChange={handleImportSettings}
                className="hidden"
                id="import-settings"
              />
              <label htmlFor="import-settings">
                <Button size="sm" variant="secondary" as="span">
                  Import
                </Button>
              </label>
            </>
          ),
        },
      ],
    },
    {
      title: 'Security',
      icon: ShieldCheckIcon,
      settings: [
        {
          label: 'Session Timeout',
          description: 'Automatically logout after inactivity',
          control: (
            <select className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500">
              <option value="15">15 minutes</option>
              <option value="30">30 minutes</option>
              <option value="60">1 hour</option>
              <option value="never">Never</option>
            </select>
          ),
        },
      ],
    },
  ];

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1,
      },
    },
  };

  const itemVariants = {
    hidden: { opacity: 0, x: -20 },
    visible: {
      opacity: 1,
      x: 0,
      transition: { duration: 0.3 },
    },
  };

  return (
    <div className="space-y-6 max-w-4xl">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          {t('nav.settings')}
        </h1>
        <p className="mt-1 text-gray-600 dark:text-gray-400">
          Manage your application preferences
        </p>
      </div>

      {/* Settings Sections */}
      <motion.div
        className="space-y-6"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        {settingSections.map((section) => (
          <motion.div
            key={section.title}
            variants={itemVariants}
            className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700"
          >
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <div className="flex items-center gap-3">
                <section.icon className="h-5 w-5 text-gray-500 dark:text-gray-400" />
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                  {section.title}
                </h2>
              </div>
            </div>
            <div className="p-6 space-y-6">
              {section.settings.map((setting, index) => (
                <div key={index} className="flex items-center justify-between">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-gray-900 dark:text-white">
                      {setting.label}
                    </p>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {setting.description}
                    </p>
                  </div>
                  <div className="ml-4">{setting.control}</div>
                </div>
              ))}
            </div>
          </motion.div>
        ))}
      </motion.div>

      {/* About Section */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
        className="bg-gradient-to-r from-primary-50 to-primary-100 dark:from-primary-900/20 dark:to-primary-800/20 rounded-lg p-6"
      >
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          About PR Reviewer
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Version 1.0.0 • Built with React, TypeScript, and Tailwind CSS
        </p>
        <div className="flex gap-2">
          <Button size="sm" variant="ghost" as="a" href="https://github.com" target="_blank">
            GitHub
          </Button>
          <Button size="sm" variant="ghost" as="a" href="/docs" target="_blank">
            Documentation
          </Button>
        </div>
      </motion.div>
    </div>
  );
};

export default Settings;
