/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  async redirects() {
    return [
      {
        source: '/privacy.html',
        destination: '/privacy',
        permanent: true,
      },
      {
        source: '/terms.html',
        destination: '/terms',
        permanent: true,
      },
      {
        source: '/payment-success.html',
        destination: '/payment-success',
        permanent: true,
      },
      {
        source: '/payment-fail.html',
        destination: '/payment-fail',
        permanent: true,
      },
      {
        source: '/works/product',
        destination: '/works/offering',
        permanent: true,
      },
      {
        source: '/works/product/:path*',
        destination: '/works/offering',
        permanent: true,
      },
    ];
  },
};

export default nextConfig;
