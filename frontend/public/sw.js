const CACHE_NAME = 'pr-reviewer-v2';
const urlsToCache = [
  '/',
  '/index.html',
  '/manifest.json',
];

// Install service worker
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        console.log('Opened cache');
        return cache.addAll(urlsToCache);
      })
  );
});

// Cache and return requests
self.addEventListener('fetch', (event) => {
  // Don't intercept API requests (localhost:8080 or any external API)
  if (event.request.url.includes('localhost:8080') || 
      event.request.url.includes('/api/') ||
      event.request.url.startsWith('http://localhost:8080') ||
      event.request.url.startsWith('https://')) {
    return; // Let browser handle it normally
  }

  // Don't cache JS and CSS files to avoid stale code
  if (event.request.url.endsWith('.js') || 
      event.request.url.endsWith('.css') ||
      event.request.url.includes('/assets/')) {
    return; // Let browser handle it normally (no caching)
  }

  event.respondWith(
    caches.match(event.request)
      .then((response) => {
        // Cache hit - return response
        if (response) {
          return response;
        }

        return fetch(event.request).then(
          (response) => {
            // Check if we received a valid response
            if (!response || response.status !== 200 || response.type !== 'basic') {
              return response;
            }

            // Clone the response
            const responseToCache = response.clone();

            caches.open(CACHE_NAME)
              .then((cache) => {
                cache.put(event.request, responseToCache);
              });

            return response;
          }
        );
      })
  );
});

// Update service worker
self.addEventListener('activate', (event) => {
  const cacheWhitelist = [CACHE_NAME];

  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheWhitelist.indexOf(cacheName) === -1) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});
