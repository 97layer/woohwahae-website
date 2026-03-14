'use client';

import { useEffect, useState } from 'react';

function isStandalone() {
  return window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone === true;
}

export default function PwaInstallButton() {
  const [installEvent, setInstallEvent] = useState(null);
  const [standalone, setStandalone] = useState(false);

  useEffect(() => {
    const sync = () => setStandalone(isStandalone());
    sync();

    const media = window.matchMedia('(display-mode: standalone)');
    const onBeforeInstallPrompt = (event) => {
      event.preventDefault();
      setInstallEvent(event);
    };
    const onInstalled = () => {
      setStandalone(true);
      setInstallEvent(null);
    };

    if (typeof media.addEventListener === 'function') {
      media.addEventListener('change', sync);
    } else if (typeof media.addListener === 'function') {
      media.addListener(sync);
    }
    window.addEventListener('beforeinstallprompt', onBeforeInstallPrompt);
    window.addEventListener('appinstalled', onInstalled);

    return () => {
      if (typeof media.removeEventListener === 'function') {
        media.removeEventListener('change', sync);
      } else if (typeof media.removeListener === 'function') {
        media.removeListener(sync);
      }
      window.removeEventListener('beforeinstallprompt', onBeforeInstallPrompt);
      window.removeEventListener('appinstalled', onInstalled);
    };
  }, []);

  if (standalone) {
    return <span className="status-pill status-pill--good">PWA</span>;
  }

  if (!installEvent) {
    return <span className="status-pill status-pill--muted">web app</span>;
  }

  async function installApp() {
    await installEvent.prompt();
    const choice = await installEvent.userChoice.catch(() => null);
    if (choice?.outcome === 'accepted') {
      setStandalone(true);
      setInstallEvent(null);
    }
  }

  return (
    <button type="button" className="status-pill pwa-install-button" onClick={installApp}>
      앱 설치
    </button>
  );
}
