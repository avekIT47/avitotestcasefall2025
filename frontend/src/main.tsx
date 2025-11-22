import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './index.css';

// App version - increment this when you need to force cache clear
const APP_VERSION = '2.0.0';

// Clear Service Worker cache on version change
if ('serviceWorker' in navigator) {
  const currentVersion = localStorage.getItem('app-version');
  if (currentVersion !== APP_VERSION) {
    console.log(`[Update] Clearing cache - ${currentVersion} -> ${APP_VERSION}`);
    
    // Unregister all service workers
    navigator.serviceWorker.getRegistrations().then(registrations => {
      for (const registration of registrations) {
        registration.unregister();
      }
    });
    
    // Clear all caches
    if ('caches' in window) {
      caches.keys().then(names => {
        for (const name of names) {
          caches.delete(name);
        }
      });
    }
    
    // Update version
    localStorage.setItem('app-version', APP_VERSION);
    
    // Reload to get fresh content
    setTimeout(() => window.location.reload(), 100);
  }
}

// Emergency fix: Clean corrupted localStorage
try {
  const appStorage = localStorage.getItem('app-storage');
  if (appStorage) {
    const data = JSON.parse(appStorage);
    if (data?.state?.selectedRows) {
      // If selectedRows exists and is a plain object (corrupted Set)
      const sr = data.state.selectedRows;
      if (sr && typeof sr === 'object' && !Array.isArray(sr) && sr.constructor === Object) {
        console.warn('[Fix] Removing corrupted selectedRows from localStorage');
        delete data.state.selectedRows;
        localStorage.setItem('app-storage', JSON.stringify(data));
      }
    }
  }
} catch (e) {
  console.error('[Fix] Error cleaning localStorage:', e);
  // If there's any error, just remove the whole thing
  localStorage.removeItem('app-storage');
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);