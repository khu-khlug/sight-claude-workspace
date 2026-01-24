# 백엔드 개발 원칙

## 1. 계층 아키텍처

### 의존성 규칙

**허용:**
- Controller → Service
- Service → Repository
- Service → Domain
- Repository → Domain/Entity

**금지:**
- Controller → Repository 직접 호출
- Controller → Domain 직접 호출
- Domain → 다른 계층 의존

### 계층별 책임

| 계층 | 책임 | 금지 |
|------|------|------|
| Controller | DTO validation, 응답 DTO 생성 | 비즈니스 로직 |
| Service | 모든 비즈니스 로직 | - |
| Domain | 순수 함수, 상태 없음 | side-effect |
| Repository | 데이터 접근 | 비즈니스 로직 |

---

## 2. 네이밍 규칙

### DTO 네이밍

```
{동작}{리소스}Request
{동작}{리소스}Response
```

**동작 접두사:** `List`, `Get`, `Create`, `Update`, `Delete`

**예시:** `ListUsersResponse`, `GetUserResponse`, `CreateUserRequest`

### Service 메서드 네이밍

```kotlin
fun list{리소스}(): List<T>
fun get{리소스}(id): T
fun create{리소스}(request): T
fun update{리소스}(id, request): T
fun delete{리소스}(id)
```

### 에러 메시지

- 한국어로 작성
- 사용자가 이해할 수 있는 문장으로

```kotlin
// Good
throw NotFoundException("사용자를 찾을 수 없습니다")

// Bad
throw NotFoundException("User not found")
```

---

## 3. API 경로 규칙

### 클래스 레벨 @RequestMapping 금지

각 메서드에 전체 경로를 직접 지정합니다.

```kotlin
// Good
@RestController
class UserController {
    @GetMapping("/api/users/{id}")
    fun getUser(@PathVariable id: Long): GetUserResponse
}

// Bad
@RestController
@RequestMapping("/api/users")  // 금지
class UserController {
    @GetMapping("/{id}")
    fun getUser(@PathVariable id: Long): GetUserResponse
}
```

**이유:** 전체 경로 파악 용이, IDE 검색 용이, 경로 충돌 방지

---

## 4. 트랜잭션 규칙

### 기본 규칙

- 조회 메서드: `@Transactional(readOnly = true)`
- 변경 메서드: `@Transactional`
- Controller에서 트랜잭션 사용 금지

### 트랜잭션 중복 방지

내부적으로 다른 트랜잭션 메서드를 호출하는 래퍼 메서드에서는 트랜잭션 어노테이션을 생략합니다.

```kotlin
@Transactional(readOnly = true)
fun getValue(key: ConfigKey): String { /* DB 조회 */ }

// 래퍼 메서드에서는 트랜잭션 불필요
fun getValueAsBoolean(key: ConfigKey): Boolean {
    return getValue(key).toBoolean()
}
```

---

## 5. Entity 원칙

### data class 사용 시 불변성 유지

**금지:** `@UpdateTimestamp` 어노테이션 (JPA 자동 변경이 불변성 위반)

**권장:** `copy()` 사용 시 `updatedAt`을 명시적으로 설정

```kotlin
// 잘못된 예
data class SomeEntity(
    @UpdateTimestamp  // X
    val updatedAt: LocalDateTime = LocalDateTime.now(),
)

// 올바른 예
val updated = existing.copy(
    someField = newValue,
    updatedAt = LocalDateTime.now()  // 명시적 갱신
)
```

**참고:** `@CreationTimestamp`는 사용 가능 (생성 시 한 번만 설정)

---

## 6. 예외 처리 규칙

### 타입 변환 시 명확한 에러 메시지

```kotlin
// 잘못된 예
fun getValueAsInt(key: ConfigKey): Int {
    return getValue(key).toInt()  // X - 원인 파악 어려움
}

// 올바른 예
fun getValueAsInt(key: ConfigKey): Int {
    val value = getValue(key)
    return try {
        value.toInt()
    } catch (e: NumberFormatException) {
        throw InvalidConfigValueException(
            "설정 값 '${key.name}'을 정수로 변환할 수 없습니다: $value"
        )
    }
}
```

### 커스텀 예외 정의

도메인별로 의미 있는 예외 클래스를 정의합니다.

---

## 7. 캐싱 규칙

### 동시성 이슈 방지

**권장 순서:**
1. 캐시 무효화 먼저 수행
2. DB 저장
3. (선택) 새 값을 캐시에 저장

```kotlin
@Transactional
fun setValue(key: ConfigKey, value: String): SystemConfig {
    cache.invalidate(key)  // 1. 캐시 무효화 먼저
    val saved = repository.save(entity)  // 2. DB 저장
    cache.put(key, value)  // 3. 캐시 저장
    return saved
}
```

### 직접 구현 캐시 (권장)

외부 라이브러리(Caffeine 등) 대신 ConcurrentHashMap으로 직접 구현합니다.

```kotlin
data class CacheEntry<V>(val value: V, val expiresAt: Instant)

private val cache = ConcurrentHashMap<Key, CacheEntry<Value>>()
private val ttlSeconds = 60L

fun get(key: Key): Value? {
    val entry = cache[key] ?: return null
    if (Instant.now().isAfter(entry.expiresAt)) {
        cache.remove(key)
        return null
    }
    return entry.value
}

fun put(key: Key, value: Value) {
    cache[key] = CacheEntry(value, Instant.now().plusSeconds(ttlSeconds))
}
```

### TTL 가이드

- 설정 값 등 자주 변경되지 않는 데이터: 1분 ~ 5분
- 자주 조회되는 마스터 데이터: 5분 ~ 30분
- 실시간성이 중요한 데이터: 캐싱 지양 또는 짧은 TTL
