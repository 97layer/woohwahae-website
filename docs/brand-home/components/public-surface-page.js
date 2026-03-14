export default function PublicSurfacePage({ page }) {
  return (
    <>
      <section className="home-section surface-index" aria-labelledby={`${page.slug}-title`}>
        <div className="home-shell__inner">
          <div className="section-grid">
            <article className="hero-copy surface-index__intro">
              <p className="section-eyebrow">{page.eyebrow}</p>
              <h1 className="section-title surface-index__title" id={`${page.slug}-title`}>
                {page.title.split('\n').map((line) => (
                  <span className="hero-title__line" key={`${page.eyebrow}-${line}`}>
                    {line}
                  </span>
                ))}
              </h1>
              <p className="section-body">{page.body}</p>
            </article>

            <aside className="proof-panel surface-index__panel">
              <p className="hero-panel__eyebrow">{page.panel.eyebrow}</p>
              <div className="proof-notes">
                <div className="hero-panel__stack-head">
                  <strong>{page.panel.title}</strong>
                </div>
                <p className="section-body">{page.panel.body}</p>
              </div>
              <div className="axis-list">
                {page.panel.links.map((link) => (
                  <a
                    key={`${page.panel.eyebrow}-${link.label}-${link.href}`}
                    className={`axis-card surface-index__link${link.active ? ' is-active' : ''}`}
                    href={link.href}
                    aria-current={link.active ? 'page' : undefined}
                  >
                    <div className="axis-card__head">
                      <strong>{link.label}</strong>
                    </div>
                    <p>{link.desc}</p>
                    <div className="axis-card__bar" aria-hidden="true">
                      <span style={{ width: link.active ? '100%' : '42%' }} />
                    </div>
                  </a>
                ))}
              </div>
            </aside>
          </div>
        </div>
      </section>

      <section className="home-section" aria-labelledby={`${page.slug}-grid-title`}>
        <div className="home-shell__inner">
          <p className="section-eyebrow">{page.section.eyebrow}</p>
          <h2 className="section-title" id={`${page.slug}-grid-title`}>{page.section.title}</h2>
          <div className="signal-grid">
            {page.section.cards.map((card) => (
              <article className="signal-card" key={`${page.slug}-${card.eyebrow}-${card.title}`}>
                <p className="section-eyebrow">{card.eyebrow}</p>
                <h3 className="signal-card__title">{card.title}</h3>
                <p>{card.body}</p>
              </article>
            ))}
          </div>
        </div>
      </section>
    </>
  );
}
