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

## 역할 범위

- **O**: 컴포넌트, hooks, API 함수 작성
- **X**: 코드 리뷰 (별도 skill 사용)
