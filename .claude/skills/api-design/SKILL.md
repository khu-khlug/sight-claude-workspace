---
name: api-design
description: REST API 엔드포인트를 설계합니다. API 스펙 정의, 엔드포인트 설계, 요청/응답 DTO 설계가 필요할 때 사용하세요.
---

# API 설계

REST API 엔드포인트 스펙을 정의합니다. 설계만 하고 코드는 작성하지 않습니다.

## Instructions

### 1. 기존 컨벤션 확인
프로젝트의 DTO 네이밍 규칙 준수:
- 목록 조회: `List{리소스}Response`, `List{리소스}sResponse`
- 단건 조회: `Get{리소스}Response`
- 생성: `Create{리소스}Request/Response`
- 수정: `Update{리소스}Request/Response`
- 삭제: `Delete{리소스}Request/Response`

### 2. 엔드포인트 설계
각 API에 대해 정의:
- HTTP 메서드 (GET/POST/PUT/DELETE)
- 경로
- 권한 (USER/MANAGER)
- 요청 파라미터/본문
- 응답 본문
- 상태 코드

### 3. 출력 형식

```markdown
### [HTTP METHOD] [경로]
- **권한**: [필요한 권한]
- **설명**: [API 설명]
- **요청 파라미터**: (있는 경우)
  - [파라미터명] ([타입]): [설명]
- **요청 본문**: (있는 경우)
  ```typescript
  { 필드명: 타입 }
  ```
- **응답**:
  ```typescript
  { 필드명: 타입 }
  ```
- **상태 코드**: [HTTP 상태 코드]
```

## 역할 범위

- **O**: API 스펙 설계, DTO 구조 정의, 경로/메서드 결정
- **X**: 코드 작성 (별도 skill 사용)
