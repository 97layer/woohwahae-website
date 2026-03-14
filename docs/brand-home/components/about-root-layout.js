import { aboutRootContent } from '../content/about-root-content';

export default function AboutRootLayout() {
  return (
    <section className="about-root" aria-labelledby="about-root-title">
      <div className="home-shell__inner about-root__inner">
        <header className="about-root__hero">
          <div className="about-root__hero-copy">
            <p className="section-eyebrow">About / Root</p>
            <h1 className="about-root__title" id="about-root-title">Root</h1>
            <p className="about-root__preamble">{aboutRootContent.preamble}</p>
          </div>

          <nav className="about-root__toc" aria-label="About root contents">
            {aboutRootContent.toc.map((item) => (
              <a className="about-root__toc-link" href={`#${item.id}`} key={item.id}>
                {item.label}
              </a>
            ))}
          </nav>
        </header>

        <details className="about-root__preface">
          <summary className="about-root__preface-toggle">{aboutRootContent.preface.title}</summary>
          <div className="about-root__preface-body">
            {aboutRootContent.preface.paragraphs.map((paragraph) => (
              <p key={paragraph}>{paragraph}</p>
            ))}
          </div>
        </details>

        <div className="about-root__sections">
          {aboutRootContent.sections.map((section) => (
            <section className="about-root__section" id={section.id} key={section.id}>
              <div className="about-root__section-head">
                <span className="about-root__section-index">{section.index}</span>
                <span className="about-root__section-sub">{section.sub}</span>
              </div>

              <div className="about-root__section-body">
                {section.blocks.map((block, blockIndex) => (
                  <article className="about-root__block" key={`${section.id}-${blockIndex}`}>
                    {block.title ? <p className="about-root__glyph">{block.title}</p> : null}
                    {block.lead ? <p className="about-root__lead">{block.lead}</p> : null}

                    {block.paragraphs?.map((paragraph) => (
                      <p className="about-root__paragraph" key={`${section.id}-${blockIndex}-${paragraph}`}>
                        {paragraph}
                      </p>
                    ))}

                    {block.declaration ? (
                      <p className="about-root__declaration">{block.declaration}</p>
                    ) : null}
                  </article>
                ))}
              </div>
            </section>
          ))}
        </div>

        <section className="about-root__appendix" aria-labelledby="about-root-appendix-title">
          <div className="about-root__section-head">
            <span className="about-root__section-index" id="about-root-appendix-title">부록</span>
          </div>

          <div className="about-root__section-body">
            {aboutRootContent.appendix.map((item) => (
              item.href ? (
                <a className="about-root__appendix-link" href={item.href} key={item.title}>
                  <span>{item.title}</span>
                  <span>→</span>
                </a>
              ) : (
                <details className="about-root__appendix-block" key={item.title}>
                  <summary className="about-root__appendix-toggle">{item.title}</summary>
                  <div className="about-root__appendix-body">
                    {item.items?.map((entry) => (
                      <article className="about-root__appendix-entry" key={entry.term}>
                        <p className="about-root__appendix-term">{entry.term}</p>
                        <p className="about-root__appendix-text">{entry.body}</p>
                      </article>
                    ))}
                  </div>
                </details>
              )
            ))}
          </div>
        </section>
      </div>
    </section>
  );
}
