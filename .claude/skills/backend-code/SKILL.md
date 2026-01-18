---
name: backend-code
description: 백엔드 코드를 작성합니다. Controller, Service, Repository, Entity 코드 작성이 필요할 때 사용하세요.
---

# Backend 코드 작성

Spring Boot + Kotlin 기반 백엔드 코드를 작성합니다. CLAUDE.md 컨벤션을 따릅니다.

## Instructions

### 1. 계층 구조 준수
```
controllers → service → domain
     ↓
 repository
```

### 2. 각 계층의 책임
- **Controller**: DTO validation, 응답 DTO 생성만. 비즈니스 로직 금지
- **Service**: 모든 비즈니스 로직 담당
- **Domain**: 순수 함수, side-effect 없음
- **Repository**: 데이터 접근만

### 3. 네이밍 규칙
- DTO: `List/Get/Create/Update/Delete{리소스}Request/Response`
- Service 메서드: `list/get/create/update/delete{리소스}`
- 에러 메시지: 한국어로 작성

### 4. API 경로 규칙
- 클래스 레벨 `@RequestMapping` 사용 금지
- 각 메서드에 전체 경로 직접 지정

## 예시

```kotlin
// Controller
@GetMapping("/users/{id}")
fun getUser(@PathVariable id: Long): GetUserResponse {
    val user = userService.getUserById(id)
    return GetUserResponse.from(user)
}

// Service
fun getUserById(id: Long): User {
    return userRepository.findById(id)
        .orElseThrow { NotFoundException("사용자를 찾을 수 없습니다") }
}
```

## 역할 범위

- **O**: Controller, Service, Repository, Entity, DTO 코드 작성
- **X**: 테스트 코드 작성, 코드 리뷰 (별도 skill 사용)
