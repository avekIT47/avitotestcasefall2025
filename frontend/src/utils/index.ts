import { format, parseISO, formatDistance } from 'date-fns';
import { ru, enUS } from 'date-fns/locale';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: string | Date | null | undefined, locale: 'ru' | 'en' = 'ru'): string {
  if (!date) {
    return '-';
  }
  
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date;
    
    // Проверка на валидность даты
    if (isNaN(dateObj.getTime())) {
      return '-';
    }
    
    return format(dateObj, 'dd.MM.yyyy HH:mm', {
      locale: locale === 'ru' ? ru : enUS,
    });
  } catch (error) {
    console.error('Error formatting date:', error);
    return '-';
  }
}

export function formatRelativeTime(date: string | Date | null | undefined, locale: 'ru' | 'en' = 'ru'): string {
  if (!date) {
    return '-';
  }
  
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date;
    
    // Проверка на валидность даты
    if (isNaN(dateObj.getTime())) {
      return '-';
    }
    
    return formatDistance(dateObj, new Date(), {
      addSuffix: true,
      locale: locale === 'ru' ? ru : enUS,
    });
  } catch (error) {
    console.error('Error formatting relative time:', error);
    return '-';
  }
}

export function downloadFile(content: string, filename: string, type: string = 'text/csv') {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

export function exportToCSV(data: any[], headers: string[], filename: string = 'export.csv') {
  const csvHeaders = headers.join(',');
  const csvRows = data.map(row => 
    headers.map(header => {
      const keys = header.split('.');
      let value = row;
      for (const key of keys) {
        value = value?.[key];
      }
      return typeof value === 'string' && value.includes(',') 
        ? `"${value}"` 
        : value ?? '';
    }).join(',')
  );
  
  const csv = [csvHeaders, ...csvRows].join('\n');
  downloadFile(csv, filename);
}

export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout>;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

export function getStatusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'open':
      return 'text-blue-600 bg-blue-100 dark:text-blue-400 dark:bg-blue-900/30';
    case 'merged':
      return 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-900/30';
    case 'closed':
      return 'text-red-600 bg-red-100 dark:text-red-400 dark:bg-red-900/30';
    case 'active':
      return 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-900/30';
    case 'inactive':
      return 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-900/30';
    default:
      return 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-900/30';
  }
}

export function truncateString(str: string, maxLength: number = 50): string {
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength - 3) + '...';
}

// PWA helpers
export async function registerServiceWorker() {
  // Temporarily disabled to clear cache
  console.log('Service Worker registration is temporarily disabled');
  
  // Force unregister all existing service workers
  if ('serviceWorker' in navigator) {
    try {
      const registrations = await navigator.serviceWorker.getRegistrations();
      for (const registration of registrations) {
        await registration.unregister();
        console.log('Service Worker unregistered:', registration);
      }
    } catch (error) {
      console.error('Service Worker unregistration failed:', error);
    }
  }
  
  // Clear all caches
  if ('caches' in window) {
    try {
      const cacheNames = await caches.keys();
      for (const name of cacheNames) {
        await caches.delete(name);
        console.log('Cache deleted:', name);
      }
    } catch (error) {
      console.error('Cache deletion failed:', error);
    }
  }
}

export function checkOnlineStatus(): boolean {
  return navigator.onLine;
}

// Notification helpers
export async function requestNotificationPermission(): Promise<boolean> {
  if (!('Notification' in window)) {
    return false;
  }
  
  if (Notification.permission === 'granted') {
    return true;
  }
  
  if (Notification.permission !== 'denied') {
    const permission = await Notification.requestPermission();
    return permission === 'granted';
  }
  
  return false;
}

export function showNotification(title: string, options?: NotificationOptions) {
  if (Notification.permission === 'granted') {
    new Notification(title, options);
  }
}
