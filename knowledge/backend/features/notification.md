# Notification 기능

사용자에게 알림을 발송하고 관리하는 기능입니다.

## 개요

- **저장소**: MySQL (`notification` 테이블)
- **ID 생성**: ULID 기반 문자열
- **카테고리**: SYSTEM(시스템 공지), GROUP(그룹 관련)

## 파일 위치

```
src/main/kotlin/com/sight/domain/notification/
├── Notification.kt           # Entity
└── NotificationCategory.kt   # 카테고리 Enum

src/main/kotlin/com/sight/repository/
└── NotificationRepository.kt # Repository

src/main/kotlin/com/sight/service/
└── NotificationService.kt    # Service
```

## 알림 생성 방법

### Service 주입

```kotlin
@Service
class SomeService(
    private val notificationService: NotificationService,
) {
    // ...
}
```

### 알림 생성

```kotlin
notificationService.createNotification(
    userId = 12345L,
    category = NotificationCategory.GROUP,
    title = "그룹 가입 승인",
    content = "OO 그룹 가입이 승인되었습니다.",
    url = "/groups/123",  // optional, 클릭 시 이동할 URL
)
```

## Entity 스펙

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | String | ULID 기반 PK |
| `userId` | Long | 알림 수신자 ID |
| `category` | NotificationCategory | SYSTEM, GROUP |
| `title` | String | 제목 (최대 200자) |
| `content` | String | 내용 (최대 1000자) |
| `url` | String? | 클릭 시 이동 URL (최대 500자) |
| `readAt` | LocalDateTime? | 읽은 시각 (null이면 미읽음) |
| `createdAt` | LocalDateTime | 생성 시각 |

## API 엔드포인트

| 메서드 | 경로 | 권한 | 설명 |
|--------|------|------|------|
| GET | `/notifications` | USER, MANAGER | 알림 목록 조회 (offset/limit 페이지네이션) |
| POST | `/notifications/read` | USER, MANAGER | 알림 일괄 읽음 처리 |
| POST | `/internal/notifications` | SYSTEM | 알림 생성 (Internal API) |

## 사용 예시

### 그룹 가입 승인 시 알림

```kotlin
notificationService.createNotification(
    userId = member.userId,
    category = NotificationCategory.GROUP,
    title = "그룹 가입 승인",
    content = "${group.name} 그룹 가입이 승인되었습니다.",
    url = "/groups/${group.id}",
)
```

### 시스템 공지 발송

```kotlin
notificationService.createNotification(
    userId = targetUserId,
    category = NotificationCategory.SYSTEM,
    title = "시스템 점검 안내",
    content = "2024년 1월 1일 02:00~04:00 시스템 점검이 예정되어 있습니다.",
)
```
