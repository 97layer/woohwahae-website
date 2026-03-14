'use client';

import { useEffect } from 'react';

function syncDisplayMode() {
  const standalone = window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone === true;
  document.documentElement.classList.toggle('app-standalone', standalone);
  document.documentElement.classList.toggle('app-browser', !standalone);
}

export default function PwaRegistration() {
  useEffect(() => {
    syncDisplayMode();

    const media = window.matchMedia('(display-mode: standalone)');
    const onModeChange = () => syncDisplayMode();
    if (typeof media.addEventListener === 'function') {
      media.addEventListener('change', onModeChange);
    } else if (typeof media.addListener === 'function') {
      media.addListener(onModeChange);
    }

    async function registerWorker() {
      if (!('serviceWorker' in navigator)) {
        return;
      }
      try {
        const registration = await navigator.serviceWorker.register('/sw.js', { scope: '/' });
        if (typeof registration.update === 'function') {
          registration.update().catch(() => {});
        }
      } catch {
        // Keep the shell usable even when service worker registration fails.
      }
    }

    registerWorker();

    return () => {
      if (typeof media.removeEventListener === 'function') {
        media.removeEventListener('change', onModeChange);
      } else if (typeof media.removeListener === 'function') {
        media.removeListener(onModeChange);
      }
    };
  }, []);

  return null;
}
