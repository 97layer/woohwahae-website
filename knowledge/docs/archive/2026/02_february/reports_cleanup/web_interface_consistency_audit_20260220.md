# WOOHWAHAE 웹사이트 인터페이스 일관성 감사 보고서
**날짜**: 2026-02-20
**범위**: website/ 전체 HTML 페이지 (25개)
**기준**: [knowledge/docs/WOOHWAHAE_BRAND_MANUAL.md](knowledge/docs/WOOHWAHAE_BRAND_MANUAL.md)

---

## 🔴 CRITICAL — 즉시 수정 필요

### 1. 네비게이션 불일치 (최우선)

**문제**: 3가지 네비게이션 구조가 혼재

| 페이지 그룹 | 네비게이션 항목 | 파일 수 |
|------------|----------------|---------|
| **현재 표준** (index/about/service/contact/project) | Archive / Service / Project / About / Contact | 10개 |
| **구 버전** (photography/atelier/shop/404/privacy/terms) | Archive / Atelier / Shop / Contact | 6개 |
| **혼합** (shop.html의 nav) | Archive / Service / About / Contact | 1개 |

**영향**: 사용자가 페이지 이동 시 네비게이션 구조가 바뀌어 혼란.

**해결 방안**:
```
표준 네비게이션 확정 필요:
Option A (현재 주력): Archive / Service / Project / About / Contact
Option B (구버전 살림): Archive / Atelier / Shop / About / Contact
```

**해당 파일**:
- 구버전 nav: `photography.html`, `atelier.html`, `shop.html`, `404.html`, `privacy.html`, `terms.html`
- 표준 nav: `index.html`, `about.html`, `service.html`, `contact.html`, `project.html`, `archive/index.html` + 하위 issue 전체

---

### 2. CSS 버전 불일치 (캐싱 이슈)

**문제**: 3가지 CSS 버전 혼재

| 버전 | 파일 수 | 페이지 예시 |
|------|---------|------------|
| `style.css?v=typo_v29` | 6개 | index, about, service, contact, project, archive/index |
| `style.css?v=typo_v18` | 8개 | archive/issue-00 ~ issue-008 |
| `style.css` (버전 없음) | 11개 | photography, atelier, shop, privacy, terms, 404 |

**영향**: 스타일 업데이트 후 일부 페이지는 캐시된 구버전 CSS 사용.

**해결 방안**:
1. 모든 페이지를 `?v=typo_v29` 또는 최신 버전으로 통일
2. 자동화: 배포 스크립트에서 CSS 버전 일괄 치환

---

### 3. favicon 확장자 불일치

**문제**: 4가지 favicon 타입 혼재

| 타입 | 파일 수 | 페이지 예시 |
|------|---------|------------|
| `symbol.jpg` | 5개 | index, about, service, contact, project |
| `symbol.png` | 14개 | archive/index, issue-001~008, photography, shop |
| `symbol.svg` | 3개 | atelier, _archive_backup 일부 |
| favicon 없음 | 3개 | 404, privacy, terms |

**영향**: 브랜드 시각 아이덴티티 일관성 저하.

**해결 방안**:
- 브랜드 매뉴얼에 favicon 표준 명시 필요
- 추천: `symbol.png` 통일 (PNG가 가장 범용적)

---

## 🟡 MEDIUM — 개선 권장

### 4. nav-logo 내부 요소 불일치

**문제**: 일부 페이지만 `<span class="nav-brand-name">Woohwahae</span>` 존재

| 패턴 | 파일 수 | 페이지 |
|------|---------|--------|
| 심볼 이미지만 | 대다수 | index, about, service, project, archive 전체 |
| 심볼 + 텍스트 | 5개 | atelier, shop, privacy, terms, _archive_backup/index_v3 |

**영향**: 네비게이션 로고 높이/레이아웃 변동 가능.

**해결 방안**:
- 브랜드 가이드 결정: "로고 이미지만 vs 이미지+텍스트"
- 결정 후 전체 통일

---

### 5. lang-toggle 버튼 부재

**문제**: EN/KR 언어 토글 버튼이 일부 페이지에만 존재

| 버튼 유무 | 파일 수 | 페이지 |
|-----------|---------|--------|
| 있음 | 11개 | index, about, service, contact, project, archive/index, issue-001~003, issue-00 |
| 없음 | 14개 | photography, atelier, shop, 404, privacy, terms, issue-004~008 |

**코드**:
```html
<button class="nav-lang-toggle" id="lang-toggle" aria-label="Toggle Language">EN</button>
```

**영향**: 다국어 지원 기능이 페이지마다 다름.

**해결 방안**:
- 다국어 제공 확정 시: 모든 페이지에 추가
- 미제공 확정 시: 모든 페이지에서 제거

---

### 6. JS 파일 로딩 방식 불일치

**문제**: `defer` 속성 사용 여부 혼재

| 패턴 | 파일 수 | 예시 |
|------|---------|------|
| `defer` 사용 | 5개 | index, about, service, contact, project |
| `defer` 없음 | 20개 | 나머지 전체 |

**코드**:
```html
<!-- 현재 표준 (defer) -->
<script src="assets/js/analytics.js" defer></script>

<!-- 구버전 (defer 없음) -->
<script src="assets/js/analytics.js"></script>
```

**영향**: 페이지 로딩 성능 차이.

**해결 방안**:
- 브랜드 매뉴얼 "기술 스택 & 운영" 섹션에 `defer` 사용 원칙 명시
- 전체 페이지에 `defer` 적용 권장

---

## 🟢 MINOR — 장기 개선

### 7. 페이지별 theme-color 불일치

| 색상 | 파일 수 | 페이지 |
|------|---------|--------|
| `#FAFAF7` (표준 --bg) | 대다수 | index, about, service, 대부분 |
| `#1A1A1A` (어두운 회색) | 1개 | photography.html |

**영향**: 모바일 브라우저 상단바 색상이 photography 페이지만 다름.

**해결 방안**:
- photography 페이지가 의도적으로 dark theme인지 확인
- 아니라면 `#FAFAF7`로 통일

---

### 8. 중복 페이지 존재 (구조 정리 필요)

**문제**: `_archive_backup/` 폴더에 구버전 페이지 보관 중

```
_archive_backup/
├── index_v3_atelier.html
├── photography.html
├── playlist.html
└── project.html
```

**영향**: 혼동 가능성 + 배포 용량 증가.

**해결 방안**:
- 백업 파일을 Git 히스토리에만 보관하거나
- 외부 아카이브 폴더로 이동 (website 밖)

---

## 📊 통계 요약

| 항목 | 상태 |
|------|------|
| 전체 HTML 파일 | 25개 (+ _archive_backup 4개) |
| CSS 파일 | 1개 (style.css) |
| JS 파일 | 7개 |
| 네비게이션 구조 | 3가지 혼재 ❌ |
| CSS 버전 | 3가지 혼재 ❌ |
| favicon 타입 | 4가지 혼재 ❌ |
| 언어 토글 | 11/25 페이지만 ⚠️ |
| defer 속성 | 5/25 페이지만 ⚠️ |

---

## ✅ 수정 우선순위

### Phase 1 (즉시)
1. **네비게이션 통일** — 표준 nav 구조 확정 + 전체 적용
2. **CSS 버전 통일** — `?v=typo_v29` 일괄 적용

### Phase 2 (1주 내)
3. **favicon 통일** — `symbol.png` 표준화
4. **언어 토글 결정** — 제공 여부 확정 후 일괄 추가/삭제

### Phase 3 (장기)
5. **JS defer 통일** — analytics.js에 defer 일괄 적용
6. **구버전 파일 정리** — _archive_backup 폴더 이동

---

## 🛠️ 자동화 제안

```bash
# CSS 버전 일괄 치환 스크립트
find website -name "*.html" -type f -exec sed -i '' 's|style\.css?v=typo_v[0-9]*|style.css?v=typo_v29|g' {} \;
find website -name "*.html" -type f -exec sed -i '' 's|style\.css"|style.css?v=typo_v29"|g' {} \;

# favicon 일괄 통일 (symbol.png)
find website -name "*.html" -type f -exec sed -i '' 's|symbol\.jpg|symbol.png|g' {} \;
find website -name "*.html" -type f -exec sed -i '' 's|symbol\.svg|symbol.png|g' {} \;
find website -name "*.html" -type f -exec sed -i '' 's|type="image/jpeg"|type="image/png"|g' {} \;
find website -name "*.html" -type f -exec sed -i '' 's|type="image/svg+xml"|type="image/png"|g' {} \;
```

---

## 📋 체크리스트 (순호님 확인 필요)

- [ ] **네비게이션 표준** 결정: `Archive / Service / Project / About / Contact` vs `Archive / Atelier / Shop / Contact`
- [ ] **favicon 표준** 결정: `symbol.png` vs `symbol.jpg` vs `symbol.svg`
- [ ] **언어 토글** 제공 여부 확정
- [ ] **nav-brand-name** 텍스트 포함 여부 확정
- [ ] `photography.html`의 dark theme 의도 확인
- [ ] `_archive_backup/` 폴더 정리 방침

---

**보고서 작성**: Claude (LAYER OS)
**경로**: [knowledge/reports/web_interface_consistency_audit_20260220.md](knowledge/reports/web_interface_consistency_audit_20260220.md)
