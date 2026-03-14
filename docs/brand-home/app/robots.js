export default function robots() {
  return {
    rules: [
      {
        userAgent: '*',
        allow: '/',
        disallow: ['/admin/', '/api/admin/', '/backend/', '/.claude/'],
      },
    ],
    sitemap: 'https://woohwahae.kr/sitemap.xml',
  };
}
