import './globals.css';
import PwaRegistration from '../components/pwa-registration';

export const metadata = {
  title: 'Layer OS',
  description: 'Layer OS public shell with protected founder admin surface.',
  applicationName: 'Layer OS',
  manifest: '/manifest.webmanifest',
  appleWebApp: {
    capable: true,
    title: 'Layer OS',
    statusBarStyle: 'default',
  },
  formatDetection: {
    telephone: false,
  },
  icons: {
    icon: [
      { url: '/assets/media/brand/symbol.png' },
      { url: '/assets/media/brand/icon-192.png', sizes: '192x192', type: 'image/png' },
      { url: '/assets/media/brand/icon-512.png', sizes: '512x512', type: 'image/png' },
    ],
    apple: [
      { url: '/assets/media/brand/icon-192.png', sizes: '192x192', type: 'image/png' },
    ],
  },
  other: {
    'mobile-web-app-capable': 'yes',
  },
};

export const viewport = {
  width: 'device-width',
  initialScale: 1,
  viewportFit: 'cover',
  themeColor: '#f4f4f4',
};

export default function RootLayout({ children }) {
  return (
    <html lang="ko">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="" />
        <link rel="preconnect" href="https://cdn.jsdelivr.net" />
        <link
          rel="stylesheet"
          href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@300;400&display=swap"
        />
        <link
          rel="stylesheet"
          href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/variable/pretendardvariable.min.css"
        />
      </head>
      <body className="page-home home-extract">
        <PwaRegistration />
        {children}
      </body>
    </html>
  );
}
