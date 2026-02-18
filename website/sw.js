/* =============================================
   WOOHWAHAE — Service Worker
   Cache-First + Network Fallback
   ============================================= */

const CACHE_VERSION = 'woohwahae-v4';
const STATIC_ASSETS = [
  '/',
  '/index.html',
  '/about.html',
  '/atelier.html',
  '/shop.html',
  '/playlist.html',
  '/project.html',
  '/photography.html',
  '/archive/',
  '/assets/css/style.css',
  '/assets/js/main.js',
  '/assets/js/analytics.js',
  '/assets/img/symbol.jpg',
  '/assets/img/icon-192.png',
  '/manifest.webmanifest'
];

// ─── Install: Pre-cache static assets ───
self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_VERSION)
      .then(cache => cache.addAll(STATIC_ASSETS))
      .then(() => self.skipWaiting())
  );
});

// ─── Activate: Clean old caches ───
self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys()
      .then(keys => Promise.all(
        keys.filter(k => k !== CACHE_VERSION).map(k => caches.delete(k))
      ))
      .then(() => self.clients.claim())
  );
});

// ─── Fetch: Strategy per request type ───
self.addEventListener('fetch', event => {
  const { request } = event;
  const url = new URL(request.url);

  // Skip non-GET and external requests
  if (request.method !== 'GET' || url.origin !== location.origin) return;

  // Network-First for dynamic data (JSON)
  if (url.pathname.endsWith('.json')) {
    event.respondWith(networkFirst(request));
    return;
  }

  // Cache-First for everything else
  event.respondWith(cacheFirst(request));
});

// ─── Strategies ───
async function cacheFirst(request) {
  const cached = await caches.match(request);
  if (cached) return cached;

  try {
    const response = await fetch(request);
    if (response.ok) {
      const cache = await caches.open(CACHE_VERSION);
      cache.put(request, response.clone());
    }
    return response;
  } catch {
    // Offline fallback
    return caches.match('/') || new Response('Offline', { status: 503 });
  }
}

async function networkFirst(request) {
  try {
    const response = await fetch(request);
    if (response.ok) {
      const cache = await caches.open(CACHE_VERSION);
      cache.put(request, response.clone());
    }
    return response;
  } catch {
    const cached = await caches.match(request);
    return cached || new Response('[]', {
      headers: { 'Content-Type': 'application/json' }
    });
  }
}
