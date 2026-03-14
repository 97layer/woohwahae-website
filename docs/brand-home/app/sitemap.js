const publishedAt = '2026-03-13';

export default function sitemap() {
  return [
    { url: 'https://woohwahae.kr/', lastModified: publishedAt, changeFrequency: 'weekly', priority: 1.0 },
    { url: 'https://woohwahae.kr/about', lastModified: publishedAt, changeFrequency: 'weekly', priority: 0.9 },
    { url: 'https://woohwahae.kr/about/woosunho', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.7 },
    { url: 'https://woohwahae.kr/archive', lastModified: publishedAt, changeFrequency: 'weekly', priority: 0.9 },
    { url: 'https://woohwahae.kr/archive/log', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.7 },
    { url: 'https://woohwahae.kr/archive/curation', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.7 },
    { url: 'https://woohwahae.kr/works', lastModified: publishedAt, changeFrequency: 'weekly', priority: 0.8 },
    { url: 'https://woohwahae.kr/works/atelier', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.7 },
    { url: 'https://woohwahae.kr/works/offering', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.7 },
    { url: 'https://woohwahae.kr/works/project', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.7 },
    { url: 'https://woohwahae.kr/field-lab', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.5 },
    { url: 'https://woohwahae.kr/field-lab/neural', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.4 },
    { url: 'https://woohwahae.kr/privacy', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.4 },
    { url: 'https://woohwahae.kr/terms', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.4 },
    { url: 'https://woohwahae.kr/payment-success', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.3 },
    { url: 'https://woohwahae.kr/payment-fail', lastModified: publishedAt, changeFrequency: 'monthly', priority: 0.3 },
  ];
}
