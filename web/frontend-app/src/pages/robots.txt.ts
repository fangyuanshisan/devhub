export function GET({ site }: { site: URL }) {
  return new Response(`User-agent: *
Allow: /

Sitemap: ${new URL('/sitemap-index.xml', site).toString()}
`, {
    headers: {
      'Content-Type': 'text/plain; charset=utf-8',
    },
  });
}
