# SystemConfig 기능

시스템 설정/플래그 값을 MySQL에 저장하고, 1분 TTL 메모리 캐싱으로 효율적으로 조회하는 내부 유틸리티 기능입니다.

## 개요

- **저장소**: MySQL (`system_config` 테이블)
- **캐싱**: ConcurrentHashMap 기반 직접 구현, 1분 TTL
- **Key 관리**: `ConfigKey` Enum으로 정의
- **Value 타입**: String만 (사용처에서 파싱)
- **기본값**: Enum에 `defaultValue` 정의, DB에 없으면 기본값 반환

## 파일 위치

```
src/main/kotlin/com/sight/core/config/
├── ConfigKey.kt              # 설정 키 Enum
├── SystemConfig.kt           # Entity
├── SystemConfigRepository.kt # Repository
└── SystemConfigRegistry.kt   # Registry (캐싱 포함)

src/main/kotlin/com/sight/core/exception/
└── InvalidConfigValueException.kt  # 파싱 예외
```

## 사용 방법

### 의존성 주입

```kotlin
@Service
class SomeService(
    private val systemConfigRegistry: SystemConfigRegistry,
) {
    // ...
}
```

### 값 조회

```kotlin
// String으로 조회
val value = systemConfigRegistry.getValue(ConfigKey.SOME_KEY)

// 타입 변환 헬퍼
val boolValue = systemConfigRegistry.getValueAsBoolean(ConfigKey.MAINTENANCE_MODE)
val intValue = systemConfigRegistry.getValueAsInt(ConfigKey.MAX_GROUP_MEMBERS)
val longValue = systemConfigRegistry.getValueAsLong(ConfigKey.SESSION_TIMEOUT_MINUTES)
```

### 값 저장

```kotlin
systemConfigRegistry.setValue(ConfigKey.SOME_KEY, "new_value")
```

### 캐시 무효화

```kotlin
// 특정 키만
systemConfigRegistry.refreshCache(ConfigKey.SOME_KEY)

// 전체
systemConfigRegistry.refreshAllCache()
```

## 새 설정 키 추가 방법

`ConfigKey.kt`에 새 항목 추가:

```kotlin
enum class ConfigKey(
    val defaultValue: String,
    val description: String,
) {
    // 기존 키들...

    NEW_KEY(
        defaultValue = "default_value",
        description = "새 설정 키 설명",
    ),
}
```

## 조회 흐름

```
getValue(key) 호출
    ↓
캐시에서 조회
    ↓ (miss)
DB에서 조회
    ↓ (없음)
Enum의 defaultValue 반환 (DB 저장 안 함)
    ↓ (있음)
캐시에 저장 후 반환
```

## 현재 정의된 설정 키

| Key | 기본값 | 설명 |
|-----|--------|------|
| `KHLUG_ACCOUNT_NUMBER` | `""` | 동아리 계좌 번호 |
