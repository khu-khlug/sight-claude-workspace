# 프론트엔드 개발 원칙

---

## 1. API 구조 패턴

### 디렉토리 구조

- 도메인별로 디렉토리 구성: `src/api/{domain}/`
- 각 도메인은 `types.ts`, `index.ts`, 필요 시 `mock.ts` 포함

### 파일 구성

**types.ts**
- 모든 API 요청/응답 타입 정의
- DTO 타입들을 내보내는 combined type 정의 (예: `MainApiDto`)
- enum 타입은 대문자 snake case 값 사용

**index.ts**
- API 함수들을 객체로 묶어서 내보냄 (예: `MainApi`, `UserPublicApi`)
- 함수는 async로 정의하고 타입을 명시적으로 반환
- 주석으로 함수 역할과 파라미터 설명 추가
- Mock 구현 시 실제 API 호출 코드를 주석으로 남김

**mock.ts (선택)**
- Mock 데이터를 별도 파일로 분리
- 실제 API 응답 형태와 동일한 구조로 작성

### 네이밍 규칙

- Response DTO: `{Action}{Entity}Response` (예: `ListNotificationsResponse`)
- Request DTO: `{Action}{Entity}Request` (예: `CreateGroupRequest`)
- API 함수: camelCase 동사로 시작 (예: `getNotifications`, `createGroup`)
- API 객체: `{Domain}Api` (예: `MainApi`, `GroupManageApi`)

### Mock 구현

- 실제 API 구현 전에는 mock 데이터 반환
- 300ms delay 추가로 로딩 상태 테스트 가능
- 실제 API 호출 코드는 주석으로 보존

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
