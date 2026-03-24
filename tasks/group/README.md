# 그룹 시스템 문서 인덱스

이 문서들은 쿠러그 사이트 레거시 시스템의 **그룹 기능**을 분석하여, 새 시스템으로 재구현하기 위한 참고 자료로 정리한 것입니다.

---

## 문서 구조

```
docs/
├── index.md                        ← 현재 문서
├── GROUP_BUSINESS_RULES.md         ← 도메인/비즈니스 규칙 (핵심)
├── GROUP_IMPLEMENTATION_TASKS.md   ← 재구현 태스크 계획
├── GROUP_DATA_SCHEMA.md            ← 레거시 DB 스키마
└── GROUP_SYSTEM.md                 ← 공통 인프라 구현 상세
```

---

## 각 문서의 역할

### [GROUP_BUSINESS_RULES.md](./GROUP_BUSINESS_RULES.md)

**새 시스템 설계의 기준이 되는 핵심 문서.** 기술적 구현을 배제하고 "무엇을 해야 하는가"에 집중합니다.

담고 있는 내용:

- 그룹 분류 체계 (6개 카테고리, 4단계 공개 범위)
- 그룹 생명주기 (상태 전이 규칙, 50일 자동 중단)
- 멤버 관리 규칙 (참여/탈퇴/위임 조건)
- 칸반 보드 기능 (리스트 > 카드 > 기록의 CRUD 및 부가 기능)
- 포트폴리오, 세미나, 즐겨찾기
- ExPoint 포인트 체계 및 등급
- 특수 그룹 규칙 (고객센터, 활용 실습, 9477)
- 외부 연동 (Discord, Slack, 아이디어클라우드)

예시:

> **운영 카테고리가 아닌 그룹에서 그룹장은 탈퇴할 수 없음** — 먼저 다른 멤버에게 그룹장을 위임해야 함

> 매 시간 검사하여, 마지막 활동 이후 **50일**이 경과한 진행 중 그룹을 자동 중단 처리. 단, **그룹 정보 수정만으로는 활동으로 인정되지 않음**

> 기록은 생성 시 **기본적으로 포트폴리오에 포함**됨

---

### [GROUP_IMPLEMENTATION_TASKS.md](./GROUP_IMPLEMENTATION_TASKS.md)

**재구현을 위한 Phase별 태스크 계획.** 의존 관계와 진행 순서를 정의합니다.

담고 있는 내용:

- Phase 1~7의 태스크 목록 및 의존 관계
- 기존 시스템과의 주요 차이점 (HTML → Markdown 전환)
- Phase별 주요 설계 결정 사항
- 새 시스템 고유의 구현 주의사항

예시:

> Phase 1 (DB 스키마) → Phase 2 (공통 인프라) → Phase 3 (도메인 모델) → Phase 4~7 (병렬 가능)

> **전역 ID 채번 방식**: 기존 시스템은 `sequence` 테이블 단일 auto-increment를 모든 엔티티가 공유. 새 시스템에서 동일 방식을 유지할지 결정 필요

---

### [GROUP_DATA_SCHEMA.md](./GROUP_DATA_SCHEMA.md)

**레거시 DB 스키마.** 데이터 마이그레이션 및 새 스키마 설계 시 참고합니다.

담고 있는 내용:

- 그룹 핵심 테이블 10개의 컬럼 정의 (타입, 기본값, 제약조건)
- 공통 의존 테이블 (sequence, members, files, notification 등)
- 테이블 간 관계도
- 데이터 구조의 특이사항

예시:

> 카드가 아카이브되면 `list` 컬럼 값이 원래 리스트 ID에서 **해당 그룹의 `group.id`**로 변경됨

> `group_member` 테이블에 PK 없음. 고객센터 그룹(`id=15265`)은 이 테이블에 아무 레코드도 없지만, 비즈니스 로직에서 모든 회원이 멤버로 취급됨

---

### [GROUP_SYSTEM.md](./GROUP_SYSTEM.md)

**공통 인프라의 구현 상세.** 동일한 동작을 새 시스템에서 재구현할 때 참고합니다.

담고 있는 내용:

- 전역 ID 채번 방식 (`sequence` 테이블)
- 알림 시스템 (API v2 연동, 카테고리 코드, 수신자 유형, 고객센터 익명 처리)
- ExPoint 시스템 (변동 로직, 전체 변동 목록)
- HTML XSS 필터 (새 시스템에서는 Markdown sanitize로 대체)
- 자동 중단 스케줄러 (50일 규칙, 자동 복귀)
- 외부 API 연동 (Discord, Slack 상세 스펙)

예시:

> `$to = 'manager'`: 운영진 전체에게 발송. `secret=true`이면 발신자를 익명으로 처리

> `$to = 'broadcast'`: state 0(휴학), 1(재학)인 모든 회원에게 익명으로 발송

---

## 문서 간 참조 관계

```
GROUP_BUSINESS_RULES ←── 설계 기준으로 참조
    ↑                        │
    │                        ↓
    │               GROUP_IMPLEMENTATION_TASKS
    │                        │
    │           ┌────────────┼────────────┐
    │           ↓            ↓            ↓
    │   GROUP_DATA_SCHEMA  GROUP_SYSTEM  (새 API 설계)
    │      (DB 참고)     (인프라 참고)
    │           │            │
    └───────────┴────────────┘
         데이터 마이그레이션
```

- **새 기능을 설계할 때** → `GROUP_BUSINESS_RULES.md`에서 규칙 확인
- **구현 순서를 정할 때** → `GROUP_IMPLEMENTATION_TASKS.md`에서 Phase 확인
- **DB 스키마를 설계할 때** → `GROUP_DATA_SCHEMA.md`에서 기존 구조 참고
- **인프라(알림/포인트/스케줄러)를 구현할 때** → `GROUP_SYSTEM.md`에서 동작 상세 확인
