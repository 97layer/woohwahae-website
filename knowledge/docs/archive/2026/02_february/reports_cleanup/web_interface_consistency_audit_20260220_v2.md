# WOOHWAHAE 웹사이트 인터페이스 일관성 감사 보고서 v2.0
**날짜**: 2026-02-20
**범위**: website/ 전체 HTML 페이지 (25개) + CSS (2909줄)
**기준**: [knowledge/docs/WOOHWAHAE_BRAND_MANUAL.md](knowledge/docs/WOOHWAHAE_BRAND_MANUAL.md)

---

## 🔴 CRITICAL — 즉시 수정 필요

### 1. 네비게이션 불일치 (Header + Footer 동시)

**문제**: Header와 Footer 모두 3가지 구조 혼재

| 구분 | 패턴 A | 패턴 B | 패턴 C |
|------|--------|--------|--------|
| **Header** | Archive/Service/Project/About/Contact (10개) | Archive/Atelier/Shop/Contact (6개) | Archive/Service/About/Contact (1개, shop.html) |
| **Footer** | 동일 (대다수) | Archive/Atelier/Shop/Contact/About (순서 다름, photography.html) | Project 누락 (shop.html) |

**photography.html Footer 특이점**: About이 마지막 위치

**해결 방안**:
1. 표준 네비게이션 확정: `Archive / Service / Project / About / Contact` 권장
2. Header + Footer 동시 일괄 수정

**영향**: 사용자 경험 일관성 파괴 — 페이지마다 메뉴 구조가 달라짐

---

### 2. CSS 버전 불일치 (캐싱 문제)

| 버전 | 파일 수 | 페이지 예시 |
|------|---------|------------|
| `?v=typo_v29` | 6개 | index, about, service, contact, project, archive/index |
| `?v=typo_v18` | 8개 | archive/issue-00 ~ issue-008 |
| 버전 없음 | 11개 | photography, atelier, shop, privacy, terms, 404 |

**문제**: CSS 업데이트 후 일부 페이지는 브라우저 캐시로 구버전 계속 사용

---

### 3. favicon 확장자 불일치

| 타입 | 파일 수 | 페이지 |
|------|---------|--------|
| `symbol.jpg` | 5개 | index, about, service, contact, project |
| `symbol.png` | 14개 | archive 전체, photography, shop |
| `symbol.svg` | 3개 | atelier, _archive_backup 일부 |
| 없음 | 3개 | 404, privacy, terms |

**브랜드 가이드 미명시** — 표준 favicon 타입 없음

---

### 4. Footer 브랜드 설명 문구 불일치

| 문구 | 파일 수 | 페이지 |
|------|---------|--------|
| "Archive for Slow Life" | 21개 | 표준 (브랜드 매뉴얼 명시) |
| "Slow Life Atelier" | 3개 | photography, _archive_backup/playlist, _archive_backup/project |

**이슈**: 브랜드 매뉴얼의 태그라인("Archive for Slow Life")과 불일치

---

### 5. Footer 카피라이트 4가지 패턴

| 패턴 | 파일 수 | 예시 |
|------|---------|------|
| `© 2026 WOOHWAHAE · Archive for Slow Life` | 18개 | 표준 |
| `© 2026 WOOHWAHAE · Based in Ulsan` | 3개 | photography, _archive_backup 일부 |
| `© 2026 ... · 개인정보처리방침 · 이용약관` (링크 포함) | 1개 | index.html |
| 빈 태그 | 3개 | privacy, terms, _archive_backup/index_v3 |

---

### 6. 개인정보처리방침/이용약관 링크 누락 (법적 이슈)

**문제**: Footer에 법적 문서 링크가 4개 페이지에만 존재

| 링크 유무 | 파일 수 |
|-----------|---------|
| 있음 | 4개 (index, privacy, terms, _archive_backup/index_v3) |
| 없음 | 21개 (나머지 전체) |

**법적 리스크**: 웹사이트 운영 시 모든 페이지에서 개인정보처리방침 접근 가능해야 함

**해결 방안**: Footer 표준 템플릿에 링크 필수 포함

---

### 7. section-label 대소문자 브랜드 원칙 위반

**브랜드 매뉴얼 원칙**: "섹션 레이블 모두 uppercase DM Mono" (명시됨)

**현실**:

| 스타일 | 예시 | 페이지 |
|--------|------|--------|
| ALL UPPERCASE (정상) | `ARCHIVE FOR SLOW LIFE` | index.html 1개소만 |
| Sentence case (위반) | `Archive`, `Service`, `About` | 대다수 페이지 |
| Mixed case (위반) | `Object`, `Legal` | shop, privacy, terms |

**CSS 정의**:
```css
.section-label {
  font-family: var(--font-mono);
  text-transform: uppercase;  /* CSS로 강제하지만 HTML에서 이미 소문자 */
}
```

**이슈**: HTML에 소문자 입력 → CSS uppercase로 변환 → 일관성 없음
**해결**: HTML에 직접 UPPERCASE 입력 권장

---

## 🟡 MEDIUM — 개선 권장

### 8. Naver Booking 링크 URL 3가지 버전

| URL 패턴 | 파일 수 | 차이점 |
|----------|---------|--------|
| `?lng=...&lat=...&placePath=/stylist` (full) | 7개 | 전체 파라미터 |
| `?lng=...&lat=...` (placePath 없음) | 1개 | service.html |
| `place/1017153611` (쿼리 없음) | 2개 | photography, _archive_backup/photography |

**영향**: URL 추적 분석 불일치 + 랜딩 페이지 다를 가능성

---

### 9. 불필요한 인라인 스타일 (CSS 토큰 시스템 우회)

**발견 위치**: 여러 페이지 산재

```html
<!-- about.html -->
<span style="font-style:italic;color:var(--text-faint)">羽化</span>

<!-- service.html -->
<p style="font-size:0.8rem;color:var(--text-faint);margin-bottom:var(--space-sm);">

<!-- index.html -->
<div class="mag-row fade-in" style="border-top:1px solid var(--line);">
```

**문제**: 브랜드 매뉴얼에 명시된 CSS 토큰 시스템 우회

**해결 방안**: Utility class 생성
```css
.text-italic-faint { font-style: italic; color: var(--text-faint); }
.cta-caption { font-size: 0.8rem; color: var(--text-faint); margin-bottom: var(--space-sm); }
.divider-top { border-top: 1px solid var(--line); }
```

---

### 10. nav-logo 내부 요소 불일치

| 패턴 | 파일 수 | 페이지 |
|------|---------|--------|
| 심볼 이미지만 | 대다수 | index, about, service, project, archive 전체 |
| 심볼 + `<span class="nav-brand-name">` | 5개 | atelier, shop, privacy, terms, _archive_backup/index_v3 |

**영향**: 네비게이션 높이/레이아웃 변동 가능

---

### 11. lang-toggle 버튼 부재

| 버튼 유무 | 파일 수 | 페이지 |
|-----------|---------|--------|
| 있음 | 11개 | index, about, service, contact, project, archive/index, issue-001~003, issue-00 |
| 없음 | 14개 | photography, atelier, shop, 404, privacy, terms, issue-004~008 |

**이슈**: 다국어 지원 기능 페이지마다 다름

---

### 12. JS 파일 로딩 방식 불일치

| 패턴 | 파일 수 |
|------|---------|
| `defer` 사용 | 5개 (index, about, service, contact, project) |
| `defer` 없음 | 20개 (나머지 전체) |

**영향**: 페이지 로딩 성능 차이

---

## 🟢 MINOR — 장기 개선

### 13. 페이지별 theme-color 불일치

| 색상 | 파일 수 | 페이지 |
|------|---------|--------|
| `#FAFAF7` (표준 --bg) | 대다수 | 표준 페이지 |
| `#1A1A1A` (어두운 회색) | 1개 | photography.html |

**확인 필요**: photography 페이지 의도적 dark theme인지

---

### 14. 404 페이지 Footer 간소화

**문제**: `404.html`은 footer에 Connect 섹션 (Instagram/Booking/Email) 누락

```html
<!-- 404.html -->
<footer>
  <div class="footer-grid">
    <!-- brand + nav만 있음, Connect 섹션 없음 -->
  </div>
</footer>
```

**확인 필요**: 의도적 간소화인지

---

### 15. 중복 페이지 존재 (_archive_backup/)

**문제**: 구버전 페이지 4개 보관 중

```
_archive_backup/
├── index_v3_atelier.html
├── photography.html
├── playlist.html
└── project.html
```

**영향**: 혼동 가능성 + 배포 용량 증가

---

### 16. Instagram 링크 일관성 (재확인 필요)

**확인 필요**: 모든 페이지가 `https://instagram.com/woohwahae` (www 없음) 통일되어 있는지

---

### 17. 중복 CSS 정의 발견

**style.css 내 중복 코드**:
- `reading-progress-bar` 정의 2회 (line 2493~2498, 2522~2527)
- `article-meta-read-time` 정의 2회 (line 2500~2503, 2529~2532)
- Shop Header 정의 2회 (line 2506~2562)

**영향**: CSS 파일 크기 증가 + 유지보수 혼란

---

## 📊 통계 요약 (최종)

| 항목 | 상태 |
|------|------|
| 전체 HTML 파일 | 25개 (+ _archive_backup 4개) |
| CSS 파일 | 1개 (style.css, 2909줄) |
| JS 파일 | 7개 |
| **Header 네비게이션 구조** | 3가지 혼재 ❌ |
| **Footer 네비게이션 구조** | 3가지 혼재 ❌ |
| **Footer 브랜드 문구** | 2가지 혼재 ❌ |
| **Footer 카피라이트** | 4가지 혼재 ❌ |
| **section-label 대소문자** | 브랜드 원칙 위반 ❌ |
| CSS 버전 | 3가지 혼재 ❌ |
| favicon 타입 | 4가지 혼재 ❌ |
| 언어 토글 | 11/25 페이지만 ⚠️ |
| defer 속성 | 5/25 페이지만 ⚠️ |
| **개인정보처리방침 링크** | 4/25 페이지만 🔴 (법적 이슈) |
| **Naver Booking URL** | 3가지 혼재 ⚠️ |
| **인라인 스타일** | 다수 페이지 산재 ⚠️ |
| **CSS 중복 정의** | 3개소 ⚠️ |

---

## ✅ 수정 우선순위 (재정렬)

### Phase 1 — CRITICAL (즉시)

1. **개인정보처리방침 링크** — 모든 페이지 Footer에 추가 (법적 요구사항)
2. **네비게이션 통일** — Header + Footer 동시 표준화
3. **CSS 버전 통일** — `?v=typo_v30` 일괄 적용 (v29→v30 bump)
4. **section-label 대문자** — 브랜드 원칙 준수

### Phase 2 — HIGH (1주 내)

5. **favicon 통일** — `symbol.png` 표준화 + 브랜드 매뉴얼 명시
6. **Footer 문구 통일** — "Archive for Slow Life" 표준 적용
7. **Naver Booking URL** — 전체 쿼리 파라미터 버전 통일

### Phase 3 — MEDIUM (2주 내)

8. **인라인 스타일 제거** — Utility class 전환
9. **언어 토글 결정** — 제공 여부 확정 후 일괄 추가/삭제
10. **JS defer 통일** — analytics.js 전체 `defer` 적용

### Phase 4 — LOW (장기)

11. **CSS 중복 제거** — style.css 중복 정의 정리
12. **_archive_backup 정리** — Git 히스토리로 이동

---

## 🛠️ 자동화 스크립트

```bash
#!/bin/bash
# 웹사이트 일관성 자동 수정 스크립트

cd /Users/97layer/97layerOS/website

# 1. CSS 버전 통일 (v30)
find . -name "*.html" -type f -exec sed -i '' 's|style\.css?v=typo_v[0-9]*|style.css?v=typo_v30|g' {} \;
find . -name "*.html" -type f -exec sed -i '' 's|style\.css"|style.css?v=typo_v30"|g' {} \;

# 2. favicon 통일 (symbol.png)
find . -name "*.html" -type f -exec sed -i '' 's|symbol\.jpg|symbol.png|g' {} \;
find . -name "*.html" -type f -exec sed -i '' 's|symbol\.svg|symbol.png|g' {} \;
find . -name "*.html" -type f -exec sed -i '' 's|type="image/jpeg"|type="image/png"|g' {} \;
find . -name "*.html" -type f -exec sed -i '' 's|type="image/svg+xml"|type="image/png"|g' {} \;

# 3. Footer 브랜드 문구 통일
find . -name "*.html" -type f -exec sed -i '' 's|Slow Life Atelier|Archive for Slow Life|g' {} \;

# 4. analytics.js defer 추가
find . -name "*.html" -type f -exec sed -i '' 's|<script src="assets/js/analytics.js"></script>|<script src="assets/js/analytics.js" defer></script>|g' {} \;
find . -name "*.html" -type f -exec sed -i '' 's|<script src="../../assets/js/analytics.js"></script>|<script src="../../assets/js/analytics.js" defer></script>|g' {} \;
find . -name "*.html" -type f -exec sed -i '' 's|<script src="../assets/js/analytics.js"></script>|<script src="../assets/js/analytics.js" defer></script>|g' {} \;

echo "✅ 자동 수정 완료"
echo "⚠️ 수동 확인 필요:"
echo "  - 네비게이션 구조 통일"
echo "  - section-label 대문자화"
echo "  - 개인정보처리방침 링크 추가"
```

---

## 📋 체크리스트 (순호님 확인 필요)

### 즉시 결정 필요
- [ ] **네비게이션 표준** 확정: `Archive / Service / Project / About / Contact` vs `Archive / Atelier / Shop / About / Contact`
- [ ] **section-label** 전부 UPPERCASE로 HTML 수정 승인
- [ ] **개인정보처리방침 링크** 전체 페이지 Footer 추가 승인

### 단기 결정 필요
- [ ] **favicon 표준** 확정: `symbol.png` 권장
- [ ] **언어 토글** 제공 여부 확정
- [ ] **nav-brand-name** 텍스트 포함 여부 확정

### 확인 사항
- [ ] `photography.html`의 dark theme 의도 확인
- [ ] `404.html` Footer 간소화 의도 확인
- [ ] `_archive_backup/` 폴더 정리 방침

---

**보고서 작성**: Claude (LAYER OS)
**경로**: [knowledge/reports/web_interface_consistency_audit_20260220_v2.md](knowledge/reports/web_interface_consistency_audit_20260220_v2.md)
**버전**: 2.0 (심층 분석 — 총 17개 이슈)
