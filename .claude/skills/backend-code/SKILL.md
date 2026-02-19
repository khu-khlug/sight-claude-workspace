---
name: backend-code
description: 백엔드 코드를 작성합니다. Controller, Service, Repository, Entity 코드 작성이 필요할 때 사용하세요.
---

# Backend 코드 작성

Spring Boot + Kotlin 기반 백엔드 코드를 작성합니다.

## 원칙 참조

코드 작성 전 `knowledge/backend/principles.md`를 참조합니다.

주요 섹션:
- 계층 아키텍처
- 네이밍 규칙
- API 경로 규칙
- 트랜잭션 규칙
- Entity 원칙
- 예외 처리 규칙
- 캐싱 규칙

## 워크플로우

1. 요구사항 확인
2. `knowledge/backend/principles.md` 참조
3. 원칙에 따라 코드 작성
4. ktlint 포맷팅 적용
5. 수정으로 인해 더 이상 사용되지 않는 코드 제거

## 예시

```kotlin
// Controller
@GetMapping("/api/users/{id}")
fun getUser(@PathVariable id: Long): GetUserResponse {
    val user = userService.getUser(id)
    return GetUserResponse.from(user)
}

// Service
@Transactional(readOnly = true)
fun getUser(id: Long): User {
    return userRepository.findById(id)
        .orElseThrow { NotFoundException("사용자를 찾을 수 없습니다") }
}
```

## 역할 범위

- **O**: Controller, Service, Repository, Entity, DTO 코드 작성
- **X**: 테스트 코드 작성 (backend-test 사용), 코드 리뷰 (backend-review 사용)
