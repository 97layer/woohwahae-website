const CACHE_NAME = 'layer-os-shell-v1';
const CORE_SHELL = [
  '/',
  '/admin/login',
  '/manifest.webmanifest',
  '/assets/media/brand/icon-192.png',
  '/assets/media/brand/icon-512.png',
  '/assets/media/brand/symbol.png',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(CORE_SHELL)).then(() => self.skipWaiting()),
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys.map((key) => {
          if (key !== CACHE_NAME) {
            return caches.delete(key);
          }
          return Promise.resolve(false);
        }),
      ),
    ).then(() => self.clients.claim()),
  );
});

function isStaticAsset(request, url) {
  return (
    request.destination === 'style' ||
    request.destination === 'script' ||
    request.destination === 'image' ||
    request.destination === 'font' ||
    url.pathname.startsWith('/_next/static/')
  );
}

async function staleWhileRevalidate(request) {
  const cache = await caches.open(CACHE_NAME);
  const cached = await cache.match(request);
  const networkPromise = fetch(request)
    .then((response) => {
      if (response && response.ok) {
        cache.put(request, response.clone());
      }
      return response;
    })
    .catch(() => cached);
  return cached || networkPromise;
}

async function navigationResponse(request, url) {
  try {
    const response = await fetch(request);
    if (response && response.ok && !url.pathname.startsWith('/admin')) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, response.clone());
    }
    return response;
  } catch {
    const cache = await caches.open(CACHE_NAME);
    if (url.pathname.startsWith('/admin')) {
      return (await cache.match('/admin/login')) || (await cache.match('/'));
    }
    return (await cache.match(request)) || (await cache.match('/'));
  }
}

self.addEventListener('fetch', (event) => {
  const { request } = event;
  if (request.method !== 'GET') {
    return;
  }

  const url = new URL(request.url);
  if (url.origin !== self.location.origin) {
    return;
  }

  if (url.pathname.startsWith('/api/') || url.pathname.includes('webpack-hmr')) {
    return;
  }

  if (request.mode === 'navigate') {
    event.respondWith(navigationResponse(request, url));
    return;
  }

  if (isStaticAsset(request, url)) {
    event.respondWith(staleWhileRevalidate(request));
  }
});
