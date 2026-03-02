# 프론트엔드 개발 원칙

---

## 1. API 구조 패턴

### 디렉토리 구조

**접근 권한에 따른 분류:**
- `src/api/manage/` - 운영진만 사용하는 API
- `src/api/public/` - 운영진과 일반 회원 모두 사용하는 API

**파일 분류 기준:**
- REST API 리소스 기준으로 파일 분류 (페이지 기준 X)
- 예: `notification.ts`, `group.ts`, `schedule.ts`, `board.ts`
- 각 파일에 해당 리소스의 DTO 타입과 API 함수를 함께 정의

### 파일 구성

**리소스별 단일 파일 (예: `notification.ts`, `group.ts`)**
- 해당 리소스의 DTO 타입 정의 (접미사 `Dto` 사용: `NotificationDto`, `ListNotificationsResponseDto`)
- DTO 타입들을 내보내는 combined type 정의 (예: `NotificationPublicApiDto`)
- enum 타입은 대문자 snake case 값 사용
- API 함수들을 객체로 묶어서 내보냄 (예: `NotificationPublicApi`, `GroupManageApi`)
- 함수는 async로 정의하고 타입을 명시적으로 반환
- 주석으로 함수 역할과 파라미터 설명 추가
- Mock 구현 시 실제 API 호출 코드를 주석으로 남기고, mock 데이터는 같은 파일에 정의

### 네이밍 규칙

- Entity DTO: `{Entity}Dto` (예: `NotificationDto`, `ScheduleDto`)
- Response DTO: `{Action}{Entity}ResponseDto` (예: `ListNotificationsResponseDto`)
- Request DTO: `{Action}{Entity}RequestDto` (예: `CreateGroupRequestDto`)
- API 함수: camelCase 동사로 시작 (예: `listNotifications`, `createGroup`)
- API 객체: `{Resource}{Access}Api` (예: `NotificationPublicApi`, `GroupManageApi`)

### Mock 구현

- 실제 API 구현 전에는 mock 데이터 반환
- 300ms delay 추가로 로딩 상태 테스트 가능 (`VITE_MOCK_DELAY_MS` 환경변수로 조절)
- 실제 API 호출 코드는 주석으로 보존
- mock 데이터는 해당 API 파일 내에 정의 (별도 파일 불필요)

### 클라이언트 사용

- `src/api/client/index.ts` (v1) 또는 `src/api/client/v2.ts` (v2) 사용
- withCredentials는 true로 설정

---

## 2. 레이아웃 구현

### MainLayout

메인 애플리케이션 레이아웃으로, NavigationBar와 콘텐츠 영역을 포함합니다.

**구조:**
- NavigationBar: 상단 고정 네비게이션
- 콘텐츠 영역: max-width 1024px, 중앙 정렬
- NavigationBar 높이만큼 상단 padding 추가 (데스크톱: 88px, 모바일: 72px)

**스타일:**
- 배경색: `#f8f9fa` (밝은 회색)
- 콘텐츠 좌우 padding: 16px
- 하단 padding: 데스크톱 40px, 모바일 32px

### 페이지 구성 패턴

- 각 섹션 간 간격: 24px (Box 컴포넌트의 `gap` 속성)
- 섹션 제목: `fontSize="xl"`, `fontWeight="bold"`, `marginBottom="16px"`

---

## 3. 네비게이션 바 구현

### 구조

- 상단 고정: `position: fixed`
- blur 효과: `backdrop-filter: blur(10px)` + 반투명 배경(`rgba`)
- 최대 너비 1024px, 중앙 정렬

### 반응형

- 데스크톱(768px 이상): 가로 메뉴 + 서브 메뉴는 호버 시 드롭다운
- 모바일(768px 미만): 햄버거 메뉴 + 토글 방식 사이드 메뉴

### 알림 기능

- 알림 아이콘에 읽지 않은 개수를 빨간 배지로 표시
- 클릭 시 드롭다운으로 알림 목록 표시
- TanStack Query로 30초마다 자동 갱신(`refetchInterval: 30000`)
- 로딩 상태는 Spinner, 빈 상태는 안내 메시지

### 스타일

- 메뉴 항목: 호버 시 하단 테두리로 시각적 피드백
- 서브 메뉴: 흰색 반투명 배경 + blur 효과
- 알림 드롭다운: 흰색 배경 + 그림자
- 읽지 않은 알림: 연한 파란색 배경 표시

### 상태 관리

- 메뉴/알림 토글 상태는 컴포넌트 로컬 상태 사용
- 하나가 열리면 다른 하나는 자동으로 닫힘
- 알림 데이터는 TanStack Query로 서버 상태 관리

---

## 4. 리스트 및 카드 표시

### 데이터 로딩 상태

- TanStack Query의 `isLoading` 상태 활용
- 로딩 중: 중앙 정렬된 Spinner 표시 (`color="var(--main-color)"`)
- 로딩 컨테이너: `padding: 40px`

### 빈 데이터 상태

- 데이터가 없을 때 명확한 안내 메시지 표시
- 중앙 정렬, 회색 텍스트(`color="gray.500"`)
- "~이/가 없습니다" 형식의 명확한 메시지

### 카드 레이아웃

**그리드 구성:**
- 반응형 그리드: `display="grid"` + `gridTemplateColumns`
- 모바일: 1열, 태블릿: 2열, 데스크톱: 3열
- 카드 간격: `gap="16px"`

**카드 스타일:**
- 테두리 우선, 그림자 최소화 (`borderWidth="1px"`, `borderColor="gray.200"`)
- 호버 효과: `borderColor: "var(--main-color)"`, `transform: "translateY(-2px)"`
- 부드러운 전환: `transition="all 0.2s"`
- 내부 패딩: `padding="16px"`

### 리스트 아이템

**기본 구조:**
- 가로 배치: `display="flex"`, `alignItems="center"`
- 호버 효과: 배경색 변경 (`backgroundColor="gray.50"`)
- 클릭 가능: `cursor="pointer"`

**시간 표시:**
- 상대적 시간: dayjs의 `fromNow()` 활용
- dayjs 한국어 설정:
  ```typescript
  import relativeTime from "dayjs/plugin/relativeTime";
  import "dayjs/locale/ko";
  dayjs.extend(relativeTime);
  dayjs.locale("ko");
  ```

**배지 사용:**
- 타입/카테고리 구분에 Chakra UI Badge 사용
- 색상: 개인(`blue`), 동아리(`green`), 그룹(`purple`)

### 텍스트 처리

**말줄임:**
- `overflow="hidden"` + `textOverflow="ellipsis"` + `whiteSpace="nowrap"`
- 컨테이너에 `minWidth="0"` 설정 필요

**폰트 크기:**
- 제목: `fontSize="xl"` (섹션) 또는 `fontSize="lg"` (카드)
- 본문: `fontSize="md"`
- 부가 정보: `fontSize="sm"` 또는 `fontSize="xs"`

### Container 사용

- 각 섹션은 `Container` 컴포넌트로 감싸기
- 제목은 `fontWeight="bold"` + `marginBottom="16px"`

---

## 5. 관리자 목록+상세 UI 패턴

### 레이아웃

- 목록: `<table>` 마크업 사용 (탐색 효율)
- 상세: 우측 Drawer (Chakra UI v3 `placement="end"`)
- 목록에서 행 클릭 → Drawer 열림

### Chakra UI v3 Drawer 사용

```tsx
<Drawer.Root placement="end" size="sm" open={isOpen} onOpenChange={({ open }) => !open && onClose()}>
  <Portal>
    <Drawer.Backdrop />
    <Drawer.Positioner>
      <Drawer.Content>
        <Drawer.Header>...</Drawer.Header>
        <Drawer.Body>...</Drawer.Body>
        <Drawer.Footer>...</Drawer.Footer>
      </Drawer.Content>
    </Drawer.Positioner>
  </Portal>
</Drawer.Root>
```

- 모바일에서 Drawer가 꽉 차면 backdrop 클릭 불가 → Header에 X 버튼 명시 필수
- `<Drawer.CloseTrigger asChild>` + 커스텀 버튼으로 구현
- Drawer 닫힐 때 선택된 항목(`selectedUser`)을 즉시 null로 초기화하지 않음 → 닫힘 애니메이션 중 내용 깜빡임 방지

### 상태 관리

```typescript
const [selectedUser, setSelectedUser] = useState<User | null>(null);
const [isDrawerOpen, setIsDrawerOpen] = useState(false);

const handleSelectUser = (user: User) => {
  setSelectedUser(user);
  setIsDrawerOpen(true);
};

const handleCloseDrawer = () => {
  setIsDrawerOpen(false);
  // selectedUser는 즉시 null화하지 않음
};
```

---

## 6. 반응형 표 (Table → Card 전환)

CSS만으로 `<table>` 마크업을 모바일에서 카드형으로 전환하는 패턴.

### 핵심 CSS

```css
/* MemberTable/style.module.css */
@media (max-width: 768px) {
  .table thead { display: none; }
  .table td { border-bottom: none; padding: 2px 0; }
}

/* MemberTableRow/style.module.css */
@media (max-width: 768px) {
  .row {
    display: block;
    position: relative;
    padding: 10px 32px 10px 8px; /* 우측 공간은 아이콘 버튼용 */
    border-bottom: 1px solid #eee;
  }
  .row:last-child { border-bottom: none; }
  .row td { display: block; padding: 2px 0; }

  /* 인라인으로 묶을 셀 */
  .row td:nth-child(2),
  .row td:nth-child(3) { display: inline; }

  /* 우측 상단 고정 버튼 */
  .row td:last-child { position: absolute; top: 10px; right: 0; }
}
```

### 주의사항

- `data-label` + `::before { content: attr(data-label) }` 로 라벨 표시 가능하나 불필요 시 제거
- 모바일에서 hover 효과 제거: `.row:hover { background-color: transparent; }`
- 아이콘 버튼(ChevronRight 등)은 desktop에서 `display: none`, mobile에서 `display: block`

---

## 7. CSS 관련 주의사항

### CSS Modules 간 cascade 충돌

CSS Modules는 같은 specificity일 때 **stylesheet 로드 순서**로 우선순위가 결정됨. 부모 컴포넌트의 CSS가 자식 컴포넌트의 CSS보다 나중에 로드되면 덮어쓸 수 있음.

- 예: `MemberTable/style.module.css`의 `.table td { padding: 12px }`가 `MemberTableRow/style.module.css`의 `.row td { padding: 0 }`을 이길 수 있음
- **해결책**: 동일한 요소에 대한 스타일은 가능하면 한 파일에서 관리. 미디어 쿼리 override는 원래 스타일이 있는 파일에 작성

### `border-collapse: separate`에서 `tr` border 불가

CSS 스펙상 `border-collapse: separate` 상태에서는 `tr`에 `border`를 지정해도 렌더링되지 않음. `td` / `th`에만 적용해야 함. `tr`에 border를 쓰려면 `border-collapse: collapse`로 변경 필요.

---

## 8. 공유 코드 레벨 결정 기준

### 컴포넌트 vs 공유 CSS

- 자체 로직/상태가 없고 단순 스타일 래퍼 수준 → **공유 CSS 파일** (`badge.module.css` 등)
- 렌더링 로직이 반복되는 경우 → **컴포넌트** 분리

### 컴포넌트 배치 레벨

- **페이지 레벨** (`features/xxx/ComponentName/`): 현재 해당 feature 내에서만 사용되거나, 도메인 특화 데이터 구조에 묶인 경우
- **전역 공용** (`components/ComponentName/`): 여러 feature에서 실제로 재사용되는 경우

**원칙**: 현재 사용처 기반으로 결정. 실제 재사용 근거 없이 공용으로 올리지 않음 (YAGNI). 다른 feature에서 사용이 생기는 시점에 승격.

### 공유 상수/타입 파일

feature 내 여러 컴포넌트가 동일한 타입/상수를 사용할 때 feature 루트에 `types.ts` 또는 `constants.ts`로 추출:

```
MemberListContainer/
  types.ts         ← SearchType, StudentStatusLabel
  badge.module.css ← 공유 배지 스타일
  TagList/         ← 페이지 레벨 공유 컴포넌트
```
