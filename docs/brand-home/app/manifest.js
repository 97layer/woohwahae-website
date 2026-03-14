export default function manifest() {
  return {
    id: '/admin/login',
    name: 'Layer OS',
    short_name: 'Layer OS',
    description: 'Single Layer OS PWA for the public shell and protected founder admin cockpit.',
    start_url: '/admin/login',
    scope: '/',
    display: 'standalone',
    orientation: 'portrait',
    background_color: '#f4f4f4',
    theme_color: '#f4f4f4',
    categories: ['business', 'productivity', 'developer'],
    icons: [
      {
        src: '/assets/media/brand/icon-192.png',
        sizes: '192x192',
        type: 'image/png',
      },
      {
        src: '/assets/media/brand/icon-512.png',
        sizes: '512x512',
        type: 'image/png',
      },
      {
        src: '/assets/media/brand/symbol.png',
        sizes: '512x512',
        type: 'image/png',
        purpose: 'any',
      },
    ],
    shortcuts: [
      {
        name: '운영판',
        short_name: '운영판',
        url: '/admin',
        icons: [{ src: '/assets/media/brand/icon-192.png', sizes: '192x192' }],
      },
      {
        name: '리뷰룸',
        short_name: '리뷰룸',
        url: '/admin/review-room',
        icons: [{ src: '/assets/media/brand/icon-192.png', sizes: '192x192' }],
      },
      {
        name: '홈',
        short_name: '홈',
        url: '/',
        icons: [{ src: '/assets/media/brand/icon-192.png', sizes: '192x192' }],
      },
    ],
  };
}
