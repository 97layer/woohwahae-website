const frameRows = [
  {
    serial: '001',
    date: '2026.03',
    title: '브랜드 뒤편의 문장',
    summary: '보여지는 화면보다 뒤에서 오래 남는 문장과 판단의 축을 적어 두는 자리입니다.',
    body: [
      '로그는 결과를 설명하는 페이지보다 조금 더 안쪽에 있습니다. 화면에 보이는 선택이 왜 그렇게 정리되었는지, 무엇을 덜어냈는지, 어떤 문장이 끝까지 남았는지를 기록하는 자리입니다.',
      '그래서 레이아웃도 시각 장치보다 읽기 순서를 먼저 세웁니다. 날짜와 제목, 한 줄 요약, 그리고 조금 더 긴 문단. 그 정도면 이후 실제 글이 들어와도 질서가 흔들리지 않습니다.',
    ],
    note: '브랜드 비하인드 / 구조 메모',
    selected: true,
  },
  {
    serial: '002',
    date: '2026.03',
    title: '정렬 전 메모',
    summary: '짧게 남겨야 할 판단과 보류한 결정을 눌러 적는 두 번째 행입니다.',
    body: [
      '짧은 메모는 장문의 축약이 아니라 그날의 리듬을 가장 직접적으로 남기는 단위입니다.',
      '아직 확정되지 않은 문장도 이 행 안에서는 충분히 머무를 수 있습니다.',
    ],
    note: '짧은 기록 / 보류 메모',
  },
  {
    serial: '003',
    date: '2026.02',
    title: '긴 호흡을 위한 자리',
    summary: '본문이 길어져도 제목과 도입, 본문 흐름이 먼저 읽히도록 잡은 행입니다.',
    body: [
      '로그가 늘어나면 결국 긴 글과 짧은 글이 함께 쌓입니다. 그래서 처음부터 긴 호흡을 견딜 수 있는 폭과 줄 간격을 먼저 고정해 둡니다.',
      '텍스트가 길어도 화면이 버티는 것이 아니라, 읽는 사람이 끝까지 따라갈 수 있어야 합니다.',
    ],
    note: '장문 기록',
  },
  {
    serial: '004',
    date: '2026.02',
    title: '작은 관찰의 자리',
    summary: '방향을 바꾸는 작은 변화와 관찰을 한 줄씩 남겨 두는 행입니다.',
    body: [
      '크게 선언하지 않아도 나중에 다시 보게 되는 문장이 있습니다. 작은 관찰은 그 자체로 구조를 바꾸지는 않지만, 방향을 돌려보게 만드는 경우가 있습니다.',
      '그래서 이 행은 짧지만 지워지지 않는 기록을 위해 남겨 둡니다.',
    ],
    note: '관찰 로그',
  },
  {
    serial: '005',
    date: '2026.02',
    title: '초안의 기준선',
    summary: '콘텐츠가 모두 정리되기 전에도 텍스트 위계만 먼저 확인하는 자리입니다.',
    body: [
      '레이아웃을 먼저 세우는 이유는 내용이 들어올 자리를 확실히 고정하기 위해서입니다.',
      '제목과 메타, 도입과 본문 사이의 간격이 먼저 정리되면 이후 실제 로그를 넣을 때도 결이 일정하게 유지됩니다.',
    ],
    note: '초안 정렬',
  },
  {
    serial: '006',
    date: '2026.01',
    title: '반복의 누적',
    summary: '반복적인 작업 메모가 차분한 간격으로 누적되도록 만든 마지막 행입니다.',
    body: [
      '과정은 대개 비슷한 날들의 반복처럼 보이지만, 그 안에서 남겨야 할 차이는 분명히 존재합니다.',
      '이 마지막 행은 같은 톤의 기록이 쌓일 때도 판면이 무너지지 않는지 확인하는 기준점입니다.',
    ],
    note: '과정 축적',
  },
];

export default function ArchiveLogLayout() {
  const selectedFrame = frameRows.find((row) => row.selected) ?? frameRows[0];
  const groupedFrames = frameRows.reduce((groups, row) => {
    const year = row.date.split('.')[0];
    const currentGroup = groups[groups.length - 1];

    if (!currentGroup || currentGroup.year !== year) {
      groups.push({ year, rows: [row] });
      return groups;
    }

    currentGroup.rows.push(row);
    return groups;
  }, []);
  const rowCountLabel = String(frameRows.length).padStart(2, '0');

  return (
    <section className="archive-log" aria-labelledby="archive-log-title">
      <div className="home-shell__inner archive-log__inner">
        <div className="archive-log__surface">
          <header className="archive-log__intro">
            <div className="archive-log__title-block">
              <p className="section-eyebrow">Archive / Log</p>
              <h1 className="archive-log__title" id="archive-log-title">Log</h1>
            </div>

            <div className="archive-log__copy-block">
              <p className="archive-log__copy">브랜드 비하인드 / 작업 메모 / 과정 기록</p>

              <nav className="archive-log__subnav" aria-label="Archive sections">
                <a className="archive-log__subnav-link is-active" href="/archive/log" aria-current="page">Log</a>
                <a className="archive-log__subnav-link" href="/archive/curation">Curation</a>
              </nav>
            </div>
          </header>

          <section className="archive-log__controlband" aria-label="Log controls">
            <span className="archive-log__register-label" aria-hidden="true">기록부</span>

            <div className="archive-log__register-note" aria-hidden="true">
              <span>최신순</span>
              <span>텍스트 위주</span>
            </div>

            <div className="archive-log__metrics" aria-hidden="true">
              <span>{rowCountLabel}건</span>
            </div>
          </section>

          <div className="archive-log__stage">
            <section className="archive-log__ledger" aria-label="Log ledger">
              {groupedFrames.map((group) => (
                <section className="archive-log__year-group" aria-labelledby={`archive-log-year-${group.year}`} key={group.year}>
                  <div className="archive-log__year-head">
                    <h2 className="archive-log__year-label" id={`archive-log-year-${group.year}`}>{group.year}</h2>
                    <span className="archive-log__year-count">{String(group.rows.length).padStart(2, '0')}</span>
                  </div>

                  <ol className="archive-log__ledger-list">
                    {group.rows.map((row) => (
                      <li className={`archive-log__ledger-item${row.selected ? ' is-selected' : ''}`} key={row.serial}>
                        <article className="archive-log__entry">
                          <div className="archive-log__ledger-meta">
                            <span className="archive-log__ledger-serial">{row.serial}</span>
                            <span>{row.date}</span>
                          </div>

                          <div className="archive-log__ledger-body">
                            <h3 className="archive-log__ledger-title">{row.title}</h3>
                            <p className="archive-log__ledger-summary">{row.summary}</p>

                            <div className="archive-log__ledger-foot">
                              <span className="archive-log__ledger-caption">{row.note}</span>
                            </div>
                          </div>
                        </article>
                      </li>
                    ))}
                  </ol>
                </section>
              ))}
            </section>

            <aside className="archive-log__spotlight" aria-label="Log reading pane">
              <p className="section-eyebrow">본문 보기</p>
              <div className="archive-log__spotlight-head">
                <h2 className="archive-log__spotlight-title">{selectedFrame.title}</h2>
                <span className="archive-log__spotlight-index">{selectedFrame.serial}</span>
              </div>

              <div className="archive-log__spotlight-meta">
                <div className="archive-log__spotlight-metric">
                  <span className="archive-log__spotlight-label">분류</span>
                  <span className="archive-log__spotlight-value">Log</span>
                </div>
                <div className="archive-log__spotlight-metric">
                  <span className="archive-log__spotlight-label">번호</span>
                  <span className="archive-log__spotlight-value">{selectedFrame.serial}</span>
                </div>
                <div className="archive-log__spotlight-metric">
                  <span className="archive-log__spotlight-label">날짜</span>
                  <span className="archive-log__spotlight-value">{selectedFrame.date}</span>
                </div>
              </div>

              <p className="archive-log__spotlight-summary">{selectedFrame.summary}</p>

              <div className="archive-log__spotlight-body">
                {selectedFrame.body.map((paragraph) => (
                  <p className="archive-log__spotlight-paragraph" key={`${selectedFrame.serial}-${paragraph}`}>
                    {paragraph}
                  </p>
                ))}
              </div>

              <p className="archive-log__spotlight-note">{selectedFrame.note}</p>
            </aside>
          </div>
        </div>
      </div>
    </section>
  );
}
