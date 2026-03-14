export default function PublicUtilityPage({
  eyebrow,
  title,
  body,
  note = '',
  cards = [],
  actions = [],
}) {
  return (
    <>
      <section className="home-section surface-index" aria-labelledby={`${eyebrow}-title`}>
        <div className="home-shell__inner">
          <div className="section-grid">
            <article className="hero-copy surface-index__intro">
              <p className="section-eyebrow">{eyebrow}</p>
              <h1 className="section-title surface-index__title" id={`${eyebrow}-title`}>
                {title.split('\n').map((line) => (
                  <span className="hero-title__line" key={`${eyebrow}-${line}`}>
                    {line}
                  </span>
                ))}
              </h1>
              <p className="section-body">{body}</p>
              {note ? <p className="proof-note">{note}</p> : null}
            </article>

            <aside className="proof-panel surface-index__panel">
              <p className="hero-panel__eyebrow">Route</p>
              <div className="proof-notes">
                <div className="hero-panel__stack-head">
                  <strong>Current shell continuity</strong>
                </div>
                <p className="section-body">
                  Donor website route continuity now stays inside the current WOOHWAHAE shell.
                </p>
              </div>
              {actions.length > 0 ? (
                <div className="axis-list">
                  {actions.map((action) => (
                    <a key={`${eyebrow}-${action.href}`} className="axis-card surface-index__link" href={action.href}>
                      <div className="axis-card__head">
                        <strong>{action.label}</strong>
                      </div>
                      <p>{action.desc}</p>
                      <div className="axis-card__bar" aria-hidden="true">
                        <span style={{ width: '100%' }} />
                      </div>
                    </a>
                  ))}
                </div>
              ) : null}
            </aside>
          </div>
        </div>
      </section>

      {cards.length > 0 ? (
        <section className="home-section" aria-labelledby={`${eyebrow}-grid-title`}>
          <div className="home-shell__inner">
            <p className="section-eyebrow">{eyebrow}</p>
            <h2 className="section-title" id={`${eyebrow}-grid-title`}>What matters here</h2>
            <div className="signal-grid">
              {cards.map((card) => (
                <article className="signal-card" key={`${eyebrow}-${card.title}`}>
                  <p className="section-eyebrow">{card.eyebrow}</p>
                  <h3 className="signal-card__title">{card.title}</h3>
                  <p>{card.body}</p>
                </article>
              ))}
            </div>
          </div>
        </section>
      ) : null}
    </>
  );
}
