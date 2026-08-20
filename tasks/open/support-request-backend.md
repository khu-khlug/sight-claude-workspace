---
type: backend
---

# 지원 신청 Backend

## 작업 개요

회원이 지원 신청을 등록하고 모든 회원이 신청과 운영진 댓글을 확인할 수 있는 지원 신청 리소스를 제공한다. 기존 레거시 게시물 모델과 분리하여 지원 신청, 댓글, 접근 권한 및 알림을 Backend가 소유한다. 완료 시 `sight-frontend`가 지원 신청의 등록, 목록, 상세 조회, 수정, 삭제 및 운영진 댓글 작성에 사용할 HTTP API와 Discord·사이트 내 알림 동작을 제공한다.

## 비즈니스 동작

- `USER` 또는 `MANAGER` 역할의 회원은 제목, 내용 및 카테고리로 지원 신청을 등록할 수 있다. 카테고리는 `SERVER_SPACE`, `SUBDOMAIN`, `HARDWARE`, `BOOK`, `OTHER` 중 하나다.
- 지원 신청의 신청자는 첫 댓글이 아직 없을 때만 제목, 내용 및 카테고리를 변경하거나 지원 신청을 삭제할 수 있다. 첫 댓글이 하나라도 생성된 지원 신청은 수정하거나 삭제할 수 없다.
- `USER` 또는 `MANAGER` 역할의 모든 회원은 모든 지원 신청과 댓글을 열람할 수 있다.
- 운영진만 지원 신청에 댓글을 작성할 수 있다. 댓글의 생성은 승인, 반려 또는 완료 상태를 만들지 않는다.
- 지원 신청이 생성되면 모든 운영진에게 사이트 내 알림을 만들고 Discord 운영 알림 채널에 등록 사실을 전송한다.
- 댓글이 생성되면 모든 운영진에게 사이트 내 알림을 만든다. 댓글 작성자가 신청자가 아니면 신청자에게도 사이트 내 알림을 만든다. 운영진이 작성한 댓글이면 신청자가 Discord 연동 상태일 때만 Discord DM을 전송한다.
- 사이트 내 알림 또는 Discord 전송이 실패해도 생성된 지원 신청이나 댓글은 취소하지 않는다. Discord 전송은 자동 재시도하지 않고 실패를 기록한다.
- `Idempotency-Key`가 같은 생성 요청은 최초 성공 결과를 반환하며 알림을 다시 만들거나 Discord 메시지를 다시 전송하지 않는다. 같은 키로 다른 요청 본문을 보내면 `409 Conflict`를 반환한다.

## HTTP API 계약

### 공통 representation

- 모든 식별자는 ULID 문자열이다. 시각 field는 ISO 8601 UTC 문자열이다.
- `SupportRequest`는 `id`, `category`, `title`, `content`, `requester`, `hasComments`, `createdAt`, `updatedAt`를 가진다.
- `requester`와 댓글의 `author`는 `id`(number), `name`(string), `profile.name`(string)만 제공한다.
- `SupportRequestDetail`은 `SupportRequest`와 생성 시각 오름차순 `comments` 배열을 포함한다.
- `SupportRequestComment`는 `id`, `content`, `author`, `createdAt`를 가진다.

### 지원 신청 collection

- `POST /support-requests`는 `USER` 또는 `MANAGER` 인증이 필요하다. `Idempotency-Key` header는 1~64자의 UUID 문자열이며 필수다.
- request body는 `category`(필수 enum), `title`(필수, 공백만 허용하지 않는 최대 255자 string), `content`(필수, 공백만 허용하지 않는 string)다.
- 새 지원 신청을 생성하면 `201 Created`와 `SupportRequest`를 반환하고 `Location: /support-requests/{id}`를 포함한다. 유효하지 않은 field는 `400 Bad Request`를 반환한다.
- 같은 인증 사용자와 같은 `Idempotency-Key`의 재요청은 본문이 같을 때 최초의 `201 Created` representation을 반환한다. 본문이 다르면 `409 Conflict`를 반환한다.
- `GET /support-requests`는 `USER` 또는 `MANAGER` 인증이 필요하다. query parameter는 `offset`(기본 0, 0 이상), `limit`(기본 20, 1~100), `category`(선택 enum)다.
- 모든 인증 회원 응답은 모든 지원 신청을 `createdAt` 내림차순으로 반환한다. 응답 body는 `count`(number)와 `supportRequests`(`SupportRequest[]`)다.

### 지원 신청 item

- `GET /support-requests/{supportRequestId}`는 `USER` 또는 `MANAGER` 인증 회원에게 `200 OK`와 `SupportRequestDetail`을 반환한다. 존재하지 않는 리소스는 `404 Not Found`를 반환한다.
- `PUT /support-requests/{supportRequestId}`는 신청자만 요청할 수 있다. request body와 validation은 생성 요청과 같고, 성공 시 `200 OK`와 갱신된 `SupportRequest`를 반환한다.
- 지원 신청이 존재하지 않거나 요청자가 신청자가 아니면 `404 Not Found`를 반환한다. 첫 댓글이 생성된 경우에는 `409 Conflict`를 반환한다.
- `DELETE /support-requests/{supportRequestId}`는 신청자만 요청할 수 있으며 성공 시 `204 No Content`를 반환한다. 존재하지 않거나 요청자가 신청자가 아니면 `404 Not Found`, 첫 댓글이 생성된 경우에는 `409 Conflict`를 반환한다.

### 지원 신청 댓글 collection

- `POST /support-requests/{supportRequestId}/comments`는 `MANAGER` 인증이 필요하다. `Idempotency-Key` header와 공백만 허용하지 않는 `content` string을 가진 request body가 필수다.
- 성공 시 `201 Created`와 `SupportRequestComment`, `Location: /support-requests/{supportRequestId}/comments/{id}`를 반환한다.
- 존재하지 않는 지원 신청은 `404 Not Found`, 운영진이 아닌 회원의 요청은 `403 Forbidden`, 잘못된 요청은 `400 Bad Request`를 반환한다. 같은 idempotency key의 처리 규칙은 지원 신청 생성과 같다.
- 이번 작업은 댓글의 수정과 삭제 API를 제공하지 않는다.

### 유지할 계약

- 기존 `GET /notifications`, `POST /notifications/read` 사이트 내 알림 API의 representation과 인증 계약은 변경하지 않는다.
- 기존 Discord 연동 정보의 조회와 연결 해제 계약은 변경하지 않는다.

## 데이터베이스 계약

- `support_request` table을 추가한다. `id CHAR(26)` primary key, `requester_id BIGINT NOT NULL`, `category VARCHAR(20) NOT NULL`, `title VARCHAR(255) NOT NULL`, `content LONGTEXT NOT NULL`, `created_at DATETIME NOT NULL`, `updated_at DATETIME NOT NULL`을 가진다.
- `category`는 `SERVER_SPACE`, `SUBDOMAIN`, `HARDWARE`, `BOOK`, `OTHER`만 허용한다. `requester_id, created_at DESC, id DESC`와 `category, created_at DESC, id DESC` index를 둔다.
- `support_request_comment` table을 추가한다. `id CHAR(26)` primary key, `support_request_id CHAR(26) NOT NULL`, `author_id BIGINT NOT NULL`, `content LONGTEXT NOT NULL`, `created_at DATETIME NOT NULL`을 가진다. `support_request_id, created_at ASC, id ASC` index를 둔다.
- 댓글은 지원 신청 삭제 여부의 판정 근거이므로 `support_request_comment.support_request_id`는 `support_request.id`를 참조하고 삭제를 제한한다. 지원 신청은 댓글이 없는 경우에만 삭제되므로 cascade delete를 사용하지 않는다.
- `support_request_idempotency` table을 추가한다. `requester_id BIGINT NOT NULL`, `operation VARCHAR(32) NOT NULL`, `idempotency_key VARCHAR(64) NOT NULL`, `request_hash CHAR(64) NOT NULL`, `resource_id CHAR(26) NOT NULL`, `created_at DATETIME NOT NULL`을 저장하고 `(requester_id, operation, idempotency_key)` unique constraint를 둔다. `operation`은 `CREATE_REQUEST`, `CREATE_COMMENT`만 허용한다.
- 기존 `khlug_board`, `khlug_document`, `khlug_comment`의 지원 신청 데이터는 이 작업에서 migration 또는 backfill하지 않는다. 새 리소스는 별도 table만 사용한다.
- 새 table 생성은 additive migration이며 구버전 application과 함께 실행 가능하다. rollback은 새 API 사용 중지 후 새 table을 제거하는 방식으로만 허용한다.

## 외부 시스템 계약

- Discord Bot API를 사용해 `discord.channels.support-request-alert` 설정값(`DISCORD_SUPPORT_REQUEST_ALERT_CHANNEL_ID`)의 text channel에 지원 신청 등록 사실을 전송한다. payload에는 지원 신청 ID, 카테고리, 제목과 신청자 표시 이름을 포함하고 본문 전체는 포함하지 않는다.
- 운영진 댓글의 Discord DM은 `discord_integration`에 저장된 신청자의 Discord user ID로 전송한다. payload에는 지원 신청 ID, 제목 및 댓글 작성자 표시 이름을 포함하고 댓글 본문 전체는 포함하지 않는다.
- Discord Bot token은 기존 `DISCORD_BOT_TOKEN`을 사용한다. 운영 알림 채널 설정이 비어 있거나 채널/DM 전송이 실패하면 해당 전송을 건너뛰거나 실패로 기록하고 HTTP 응답과 DB transaction을 실패시키지 않는다.
- Discord API는 timeout을 두고 자동 재시도하지 않는다. idempotent 생성 요청의 재응답은 Discord API를 다시 호출하지 않는다.

## 보안 및 개인정보

- 모든 API는 인증이 필요하다. 목록과 상세 조회는 모든 인증 회원에게 허용하고, 댓글 작성은 운영진으로, 수정과 삭제는 신청자로 서버에서 강제한다.
- 지원 신청 제목과 내용, 댓글은 동아리 운영 관련 정보이므로 인증되지 않은 사람이나 Discord 운영 알림 채널 이외의 외부 대상에게 반환하거나 전달하지 않는다.
- Discord 운영 알림과 DM에는 요청·댓글 본문을 넣지 않는다. Discord user ID, 요청 내용 및 댓글 본문을 application log, metric, tracing attribute에 기록하지 않는다.
- Discord DM은 연동된 신청자에게만 보내며, 연동되지 않은 회원을 위해 Discord ID를 추정하거나 새 연동을 만들지 않는다.

## 비가역적 부작용 및 운영 영향

- 지원 신청 삭제는 첫 댓글 전 신청자만 수행할 수 있는 영구 삭제다. 삭제 뒤에는 목록과 상세 조회에서 반환하지 않으며 복구 기능은 제공하지 않는다.
- 지원 신청·댓글 생성은 사이트 내 알림 생성과 Discord 외부 메시지 전송을 유발한다. Discord 전송은 되돌릴 수 없으므로 생성이 중복 처리되지 않도록 idempotency key를 사용한다.
- Discord 전송 실패는 error log로 지원 신청 ID와 전송 대상 종류만 남긴다. 실패한 Discord 메시지의 수동 재전송 및 자동 재시도는 제공하지 않는다.

## 호환성 및 배포

- 새 API와 새 table은 기존 API를 변경하지 않는다. `sight-frontend`의 지원 신청 화면은 Backend 배포와 `DISCORD_SUPPORT_REQUEST_ALERT_CHANNEL_ID` 설정 후 공개한다.
- 레거시 `/support`와 `khlug_document` 기반 지원 신청은 이 작업에서 계속 동작한다. 두 시스템의 신청 데이터는 서로 조회하거나 수정하지 않는다.
- Discord 운영 알림 채널 설정 또는 Bot 권한이 준비되지 않은 환경에서도 API는 동작하지만 Discord 알림만 생략 또는 실패 기록된다.

## 검증

- 인증 회원과 운영진이 각 카테고리로 지원 신청을 생성하면 `201` 응답, 신청자 보존, 운영진 사이트 내 알림, Discord 운영 채널 전송이 발생하는지 controller·service integration test로 확인한다.
- 일반 회원과 운영진의 목록에 모든 신청이 생성 시각 내림차순으로 반환되고, 일반 회원도 다른 회원의 상세와 댓글을 조회할 수 있는지 확인한다.
- 첫 댓글 전 신청자는 수정·삭제할 수 있고, 첫 댓글 생성과 동시에 수정·삭제가 `409`가 되는지 동시 요청을 포함한 service test로 확인한다.
- 일반 회원의 댓글 요청이 `403`으로 거절되고, 운영진 댓글 생성 시 신청자 사이트 내 알림·연동된 경우의 DM 전송을 확인한다.
- 미연동 신청자, Discord DM 차단, 운영 알림 채널 누락, Discord API timeout·실패에서 지원 신청·댓글이 성공적으로 남고 HTTP 성공 응답이 유지되는지 확인한다.
- 빈 값, 허용되지 않은 카테고리, 잘못된 pagination, 중복 idempotency key 및 같은 key의 상이한 body에 각각 `400` 또는 `409`가 반환되는지 확인한다.
- 새 table migration을 빈 DB와 기존 레거시 데이터가 있는 DB에 적용해 기존 table을 변경하지 않는지 확인한다.

## 비목표

- 지원 신청의 승인, 반려, 완료 상태 및 그에 따른 예산·자산·도서·서버 자원 할당은 제공하지 않는다.
- 댓글의 수정·삭제, Discord 실패 메시지 재전송, 레거시 지원 신청 데이터 이전은 제공하지 않는다.
