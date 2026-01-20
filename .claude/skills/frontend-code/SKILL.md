---
name: frontend-code
description: 프론트엔드 코드를 작성합니다. React 컴포넌트, hooks, API 연동 코드 작성이 필요할 때 사용하세요.
---

# Frontend 코드 작성

React + TypeScript 기반 프론트엔드 코드를 작성합니다.

## Instructions

### 1. 로직/UI 분리 (필수)

- **로직**: `hooks/`, `api/` 디렉토리
- **UI**: `components/`, `pages/` 디렉토리
- 컴포넌트에 비즈니스 로직 직접 작성 금지

### 2. 서버 상태 관리

- TanStack Query 사용
- `useQuery`, `useMutation` 활용

### 3. UI 컴포넌트

- Chakra UI 사용
- 기존 컴포넌트 스타일 참고

### 4. 디렉토리 구조

```
src/
├── api/          # API 호출 함수
├── hooks/        # Custom hooks (로직)
├── components/   # 재사용 UI 컴포넌트
├── features/     # 기능별 컨테이너
└── pages/        # 라우트 페이지
```

## 예시

```typescript
// api/member/getMember.ts
export const getMember = (id: string) =>
  axios.get<GetMemberResponse>(`/members/${id}`);

// hooks/useMember.ts
export const useMember = (id: string) =>
  useQuery({ queryKey: ['member', id], queryFn: () => getMember(id) });

// components/MemberCard.tsx (UI만)
export const MemberCard = ({ member }: { member: Member }) => (
  <Card><Text>{member.name}</Text></Card>
);
```

### 5. UI/UX 구현 원칙

- 가장 중요한 가치는 "사용성"과 "일관성"입니다.
- 모든 UI는 일관성이 있어야 합니다. 필요 시 원칙 문서를 만들고, 지속적으로 유지 관리(추가/수정/삭제)해야 합니다.
  - 생성한 문서는 `sight-claude-workspace/knowledge/frontend-impl` 디렉토리 하위에 생성합니다.
  - 생성한 문서는 상세한 코드 대신 원칙을 정의해야 하며, 불필요한 형용사나 부사는 제외하고 중요한 내용만 포함해야 합니다.
- 아름다운 UI보다는 사용하기 편한 UI를 추구합니다.
- 화려한 애니메이션보다는 필요 시 적당한 애니메이션으로 자연스러운 UI 전환을 구현합니다.
- 기본적으로 모든 화면은 최대 가로 길이 1024px로 구성되어야 하며, 연관된 UI는 Container 안에 묶어 유저가 쉽게 기능을 구분할 수 있도록 합니다. Container는 그림자보다는 회색 테두리를 선호합니다.
- Container는 기본적으로 세로 방향으로 나열되나, Container 내 내용이 많지 않아 1024px가 긴 경우에는 가로 최대 2개까지만 Container를 넣을 수 있습니다. 이때 width 비율은 상황에 맞게 적절히 조절합니다.
- 다른 UI와 구분되어 확실하게 떠 있어야 하는 UI를 제외하고는 최대한 그림자 사용을 지양합니다.
- 모든 UI는 반응형 대응이 되어 있어야 합니다. 즉, 데스크톱과 모바일 화면에 대한 대응을 적절하게 구현해야 합니다.
- 60-30-10 rule을 따라, 색을 적절히 활용합니다.
  - 메인 브랜드 컬러는 `#00a0e9`입니다.
- 로딩이 들어가는 경우 상황에 맞게 적절히 로딩 관련 처리를 구현해야 합니다.
  - 최초 데이터 로딩이나 버튼 클릭으로 발생한 API 호출의 응답 대기 등 여러 상황이 있습니다. 적절히 해당 상황의 UI에 맞게

## 역할 범위

- **O**: 컴포넌트, hooks, API 함수 작성
- **X**: 코드 리뷰 (별도 skill 사용)
