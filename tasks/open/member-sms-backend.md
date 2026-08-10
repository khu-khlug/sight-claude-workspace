---
type: backend
---

# 운영진 문자 발송 기능 구현

## 작업 개요

`Sight-Legacy`의 `/manage/sms` 기능을 `sight-spring-backend`에 재구현한다. 운영진이 회원 또는 직접 입력한 전화번호를 수신자로 지정하여 동아리 운영 목적의 문자 메시지를 발송할 수 있도록 한다.

구현 기준은 [문자 발송 정책](../../policy/문자_발송/POLICY.md)으로 한다. 특히 발송 유형은 운영진이 선택하지 않고, 수신자별 치환이 완료된 메시지의 바이트 수에 따라 SMS 또는 LMS로 서버가 결정한다.

작업이 완료되면 백엔드는 운영진이 사용할 회원 조회, 공식 발신번호 관리 및 문자 발송 리소스 API를 제공해야 한다. 비운영진은 해당 API에 접근할 수 없어야 한다.

## 비즈니스 동작

- 문자 발송을 시작할 수 있는 주체는 운영진이다.
- 회원 수신자는 사이트에서 현재 회원으로 관리되고 있는 회원 중 교류 회원이 아닌 회원으로 제한한다.
- 회원 수신자의 전화번호가 없으면 해당 회원에게 발송하지 않는다.
- 회원 전화번호도 숫자 이외의 문자를 제거한 뒤 사용하며, 제거 결과가 비어 있으면 해당 회원의 결과를 `SKIPPED`로 처리한다.
- 회원이 아닌 수신자는 숫자 이외의 문자를 제거한 전화번호를 직접 지정할 수 있다.
- 직접 지정 전화번호는 `additionalPhoneNumbers` 배열로 전달한다. 배열의 각 문자열에 쉼표가 포함되어 있으면 쉼표로 먼저 분리한 뒤 각 항목을 별도 수신자로 처리한다.
- 직접 지정 전화번호는 쉼표로 분리한 각 항목에서 숫자 이외의 문자를 제거하며, 제거 결과가 비어 있는 항목은 발송 대상에서 제외한다.
- 메시지 내용이 공백만으로 이루어져 있으면 발송을 거부한다.
- 회원 수신자에 대해서는 메시지의 `{realname}`을 회원 이름으로 치환한다.
- 직접 지정 수신자에 대해서는 `{realname}`을 숫자만 남긴 정규화된 수신 전화번호로 치환한다.
- 치환이 완료된 수신자별 메시지의 바이트 수가 90바이트 이하이면 SMS로 발송한다.
- 치환이 완료된 수신자별 메시지의 바이트 수가 90바이트를 초과하면 LMS로 자동 전환한다.
- 바이트 수는 ASCII 문자를 1바이트, 그 밖의 문자를 2바이트로 계산한다.
- 운영진은 SMS 또는 LMS를 수동으로 선택하여 자동 전환을 우회할 수 없다.
- LMS의 제목은 `쿠러그, 경희대학교 중앙 IT 동아리`로 한다.
- 모든 발송에는 동아리 공식 전화번호를 발신번호로 사용한다.
- 회원과 직접 지정 수신자는 하나의 발송 요청에서 함께 지정할 수 있다.
- 동일한 회원 ID 또는 정규화된 전화번호가 여러 번 지정되면 첫 번째 지정만 발송한다. 회원 수신자와 직접 지정 수신자의 전화번호가 같으면 회원 수신자를 우선한다.
- 중복으로 제외된 입력은 발송 결과에 포함하지 않는다. 발송 결과는 중복 제거 후 회원 수신자 요청 순서, 직접 지정 수신자 요청 순서로 반환한다.
- 수신자별 발송은 독립적으로 처리하며, 한 수신자의 실패가 다른 수신자의 발송을 중단시키지 않는다.
- 수신자가 없거나 모든 수신자가 발송 대상에서 제외되면 외부 문자 발송을 수행하지 않는다.

## HTTP API 계약

### 변경 사항

#### 기존 회원 조회 API 유지

- 대상 리소스는 회원이며, 회원의 생명주기와 조회 조건은 기존 회원 도메인과 API가 소유한다.
- 회원 수신자 조회에는 기존 `GET /manager/users`를 사용한다.
- 운영진 인증과 회원 목록의 공통 기본 조건은 기존 API 계약을 유지한다.
- `name`, `limit`, `offset` query parameter로 이름 검색과 페이지네이션을 수행한다.
- 기존 응답의 `users[].id`, `users[].profile.name`, `users[].profile.college`, `users[].admission`, `users[].status`, `users[].profile.phone`을 문자 발송 수신자 선택에 사용할 수 있다.
- `limit`은 기존 API의 허용 범위인 1 이상 50 이하를 사용한다.
- 기존 회원 조회 API에 문자 발송 전용 응답 형식이나 query parameter를 추가하지 않는다.

#### `GET /manager/sender-phone`

- 대상 리소스는 동아리 공식 발신번호 설정이며, 백엔드가 단일 현재 값을 관리한다.
- 현재 동아리 공식 발신번호를 반환한다.
- 운영진만 요청할 수 있다.
- `SMS_SENDER_PHONE` 값이 비어 있으면 `404 Not Found`를 반환한다.

| response field | type | required | 의미 |
| --- | --- | --- | --- |
| `phone` | string | yes | 현재 동아리 공식 발신번호 |

#### `PUT /manager/sender-phone`

- 운영진만 요청할 수 있다.
- request body의 `phone`을 동아리 공식 발신번호로 저장한다.
- 저장 시 숫자 이외의 문자를 제거한다.
- 숫자 이외의 문자를 제거한 결과가 비어 있으면 `400 Bad Request`로 거부한다.
- 별도의 길이 또는 국내 전화번호 prefix 검증은 수행하지 않는다. SOLAPI에 사전 등록된 발신번호인지 여부도 저장 시 확인하지 않는다.
- 성공 시 `204 No Content`를 반환한다.
- 변경된 발신번호는 `PUT` 성공 이후 시작하는 발신번호 조회와 문자 발송부터 사용한다. 여러 application instance가 실행 중이어도 이전 값이 반환되거나 발송에 사용되어서는 안 된다.

| request field | type   | required | nullable | 의미          |
| ------------- | ------ | -------- | -------- | ----------- |
| `phone`       | string | yes      | no       | 동아리 공식 발신번호 |

#### `POST /manager/sms-messages`

- 대상 리소스는 문자 메시지 발송 요청이며, 발송 요청 자체를 별도로 저장하지 않고 외부 문자 발송 서비스에 전달한다.
- 운영진만 요청할 수 있다.
- request body의 `memberIds`와 `additionalPhoneNumbers`를 합쳐 수신자 목록을 만든다.
- `memberIds`에는 현재 회원으로 관리되고 있고 교류 회원이 아닌 회원의 ID만 허용한다. 조건을 만족하지 않는 회원 ID가 하나라도 포함되면 전체 요청을 `400 Bad Request`로 거부한다.
- `additionalPhoneNumbers`의 각 문자열은 쉼표로 먼저 분리한다. 분리된 각 항목에서 숫자 이외의 문자를 제거하며, 제거 결과가 비어 있는 항목은 발송 대상에서 제외한다.
- `message`가 공백만으로 이루어져 있으면 `400 Bad Request`를 반환한다.
- 유효한 회원 ID가 없고 정규화 후 남은 직접 지정 전화번호도 없으면 `400 Bad Request`를 반환한다.
- 유효한 회원의 전화번호가 없거나 정규화 결과가 비어 있으면 해당 회원을 `SKIPPED`로 결과에 포함한다. 이 회원만 지정된 요청은 `400 Bad Request`가 아니라 `422 Unprocessable Entity`를 반환하며 외부 발송은 수행하지 않는다.
- 외부 발송 대상이 하나 이상인데 `SMS_SENDER_PHONE` 값, `SOLAPI_API_KEY` 또는 `SOLAPI_API_SECRET`이 비어 있으면 외부 발송을 수행하지 않고 `500 Internal Server Error`를 반환한다.
- 서버가 수신자별 최종 메시지를 만든 뒤 바이트 수에 따라 `SMS` 또는 `LMS`를 결정한다.
- 외부 문자 발송 서비스의 처리 결과를 수신자별로 반환한다.
- 응답 상태는 다음과 같이 구분한다.
  - `200 OK`: 모든 외부 발송 대상이 SOLAPI에 정상 접수되고 `SKIPPED` 결과가 없음
  - `422 Unprocessable Entity`: SOLAPI가 하나 이상의 수신자를 거부했거나 `SKIPPED` 결과가 있으며 `results`에 수신자별 결과를 포함함
  - `400 Bad Request`: 메시지가 공백만으로 이루어졌거나, 허용되지 않는 회원 ID가 포함되었거나, 유효한 회원 ID와 정규화 후 남은 직접 지정 전화번호가 모두 없음
  - `401 Unauthorized`: 인증되지 않은 요청
  - `403 Forbidden`: 운영진이 아닌 요청
  - `500 Internal Server Error`: 설정이 누락되었거나 SOLAPI 요청 자체가 timeout, 네트워크 오류, 인증 오류 또는 비정상 HTTP 응답으로 실패하여 수신자별 접수 결과를 얻을 수 없음
- 성공 또는 부분 성공 응답은 `results` 배열을 포함한다.
- `500 Internal Server Error`는 공통 오류 응답을 사용하며 `results`를 반환하지 않는다.
- SOLAPI가 외부 발송 대상을 모두 거부한 경우에도 수신자별 결과를 얻었다면 `422 Unprocessable Entity`와 `results`를 반환한다.

| request field            | type            | required | nullable | 의미                     |
| ------------------------ | --------------- | -------- | -------- | ---------------------- |
| `memberIds`              | array of number | yes      | no       | 회원 수신자 ID 목록. 빈 배열 허용  |
| `additionalPhoneNumbers` | array of string | yes      | no       | 직접 지정 전화번호 목록. 빈 배열과 원소별 쉼표 구분 허용 |
| `message`                | string          | yes      | no       | 발송할 메시지 원문             |

| response field       | type           | required | 의미                        |
| -------------------- | -------------- | -------- | ------------------------- |
| `results`            | array          | yes      | 수신자별 발송 결과                |
| `results[].memberId` | number or null | yes      | 회원 수신자인 경우 회원 ID          |
| `results[].phone`    | string or null | yes      | 정규화된 수신 전화번호. 전화번호가 없는 회원은 `null` |
| `results[].type`     | string or null | yes      | `SMS` 또는 `LMS`. `SKIPPED` 결과는 `null` |
| `results[].status`   | string         | yes      | `SENT`, `FAILED` 또는 `SKIPPED` |
| `results[].message`  | string or null | yes      | `FAILED` 또는 `SKIPPED` 사유를 나타내는 한국어 메시지 |

- `SENT`는 SOLAPI가 발송 요청을 정상 접수했다는 의미이며 수신자의 최종 수신 완료를 의미하지 않는다.
- 직접 지정 전화번호는 숫자 이외의 문자를 제거한 결과가 비어 있지 않으면 외부 발송 대상으로 인정하며 별도의 길이 또는 국내 전화번호 prefix 검증을 수행하지 않는다. SOLAPI가 전화번호를 거부하면 `FAILED`로 변환한다.
- `results`는 중복 제거 후 회원 수신자를 `memberIds`의 최초 등장 순서로 먼저 반환하고, 직접 지정 수신자를 `additionalPhoneNumbers`와 원소별 쉼표 분리 결과의 최초 등장 순서로 반환한다.
- 요청 재시도 시 동일한 문자 메시지가 중복 발송될 수 있으므로, 별도의 멱등성 보장은 하지 않는다.
- 기존 회원 조회 API를 변경하지 않으며 문자 메시지 발송 리소스를 새로 추가한다.

## 데이터베이스 계약

- 회원의 기존 `phone`, 회원 상태 및 교류 회원 분류를 사용한다.
- 회원 테이블의 기존 데이터 구조는 변경하지 않는다.
- 동아리 공식 발신번호는 기존 `system_config` 설정 저장소의 `SMS_SENDER_PHONE` key에 저장한다.
- 발신번호 변경 성공 이후 시작하는 요청은 여러 application instance에서도 변경된 값을 조회하고 사용해야 하므로 instance-local stale cache에서 이전 값을 반환하지 않는다.
- 문자 발송 이력 테이블과 이력 backfill은 이번 작업에 포함하지 않는다.

## 외부 시스템 계약

- 외부 문자 발송 서비스는 SOLAPI를 사용한다.
- SOLAPI [메시지 발송 API](https://solapi.com/developers/api/messages)의 `POST https://api.solapi.com/messages/v4/send-many/detail`을 사용한다.
- 백엔드는 외부 발송 대상 전체를 `messages` 배열에 담아 한 번의 SOLAPI 요청으로 전달한다.
- 각 `messages[]`에는 `from`, `to`, `text`, `type`, `autoTypeDetect`, `customFields`를 전달한다.
- `type`은 서버가 계산한 `SMS` 또는 `LMS`를 명시하고 `autoTypeDetect`는 `false`로 전달하여 SOLAPI의 자동 유형 판별이 서버의 결정을 변경하지 못하게 한다.
- LMS인 `messages[]`에는 `subject=쿠러그, 경희대학교 중앙 IT 동아리`를 추가한다. SMS에는 `subject`를 전달하지 않는다.
- `customFields`에는 요청 내 수신자를 식별할 수 있는 비개인정보 순번만 전달하며 회원 ID, 이름, 전화번호 또는 메시지 내용을 추가하지 않는다.
- SOLAPI 인증 정보는 `SOLAPI_API_KEY`와 `SOLAPI_API_SECRET` 환경 변수로 주입하며 소스 코드와 응답에 저장하지 않는다.
- SOLAPI 요청은 [API Key 인증 계약](https://solapi.com/developers/api/authentication-api-key)에 따라 `Content-Type: application/json`과 `Authorization: HMAC-SHA256 apiKey=..., date=..., salt=..., signature=...` header를 사용한다.
- 인증 signature는 UTC ISO 8601 형식의 `date`와 요청마다 새로 생성한 `salt`를 이어 붙인 값을 API secret으로 HMAC-SHA256 처리하여 만든다.
- SOLAPI 연결 및 응답 timeout은 각각 10초로 설정하고 자동 재시도하지 않는다.
- SOLAPI 요청 자체의 timeout, 네트워크 오류, 인증 오류 또는 비정상 HTTP 응답은 요청 단위 실패로 처리하고 `500 Internal Server Error`로 변환한다.
- 자동 재시도는 중복 문자 발송을 일으킬 수 있으므로 기본으로 수행하지 않는다.
- SOLAPI가 정상 HTTP 응답 안에서 일부 수신자를 거부한 경우 정상 접수된 수신자의 발송은 취소하지 않고 부분 성공으로 반환한다.
- SOLAPI 응답의 `messageList`와 `failedMessageList`를 `customFields`의 비개인정보 순번으로 요청 수신자와 연결하여 각각 `SENT` 또는 `FAILED`로 변환한다.
- 외부 발송 대상이 없으면 SOLAPI를 호출하지 않는다.

## 보안 및 개인정보

- 모든 문자 발송 API는 인증된 운영진만 호출할 수 있어야 한다.
- 회원 전화번호는 운영진 문자 발송 화면과 발송 처리에 필요한 범위에서만 조회·전달한다.
- 직접 입력 전화번호와 메시지 내용은 URL, query parameter 및 일반 애플리케이션 로그에 포함하지 않는다.
- SOLAPI 인증 정보와 `Authorization` header는 응답, 로그 및 예외 메시지에 노출하지 않는다.
- 문자 메시지에 포함된 회원 이름과 전화번호는 발송 목적 외에 사용하지 않는다.
- 백엔드는 이번 작업에서 문자 메시지와 전화번호의 별도 영구 보존을 추가하지 않는다. SOLAPI에 전달된 데이터에는 SOLAPI의 데이터 보관 정책이 적용된다.

## 비가역적 부작용 및 운영 영향

- 문자 발송은 외부 서비스 비용을 발생시키며, 실제 수신자에게 되돌릴 수 없는 알림을 전달한다.
- 회원 수와 직접 지정 전화번호 수에 비례하여 SOLAPI 요청 payload와 발송 비용이 증가한다. 외부 발송 대상은 한 번의 SOLAPI 요청으로 전달한다.
- 서버는 최종 메시지와 문자 유형을 직접 계산하며 요청에 포함된 문자 유형을 신뢰하지 않는다.
- SOLAPI 요청 자체가 실패하면 수신자별 접수 여부를 확정할 수 없으므로 `500 Internal Server Error`를 반환하고 자동 재전송하지 않는다.
- SOLAPI가 정상 응답 안에서 일부 수신자를 거부하면 정상 접수된 수신자를 취소하지 않고 `422 Unprocessable Entity`와 수신자별 결과를 반환한다.
- 동일 요청의 재시도는 중복 발송을 일으킬 수 있으므로 자동 재전송하지 않는다.
- 발송 이력 저장, 잔액 조회, 예약 발송 및 대량 발송 rate limit은 이번 작업에 포함하지 않는다.

## 호환성 및 배포

- 새 API를 추가 방식으로 배포한다.
- 기존 API consumer에 영향을 주지 않도록 새 API를 추가 방식으로 제공한다.
- 외부 문자 발송 인증 정보와 공식 발신번호가 운영 환경에 설정된 뒤 발송 기능을 활성화한다.
- 레거시 구현의 API key와 secret은 새 백엔드에서 재사용하지 않는다. 운영 전 새 SOLAPI 인증 정보를 발급하고 레거시 인증 정보는 폐기한다.
- 구버전 레거시 `/manage/sms`와 새 API를 동시에 사용하면 중복 발송이 발생할 수 있으므로 운영 전환 시점에 발송 경로를 하나로 정한다.
- 이번 작업은 기존 레거시 데이터의 변환을 요구하지 않는다.
- API consumer가 새 API를 사용하기 전에 새 API를 배포한다.
- rollback 시 새 API의 접근을 차단하고 기존 레거시 발송 경로를 사용한다. 기존 회원 데이터와 `system_config` 구조는 변경하지 않으므로 데이터 rollback은 필요하지 않다.

## 검증

- 운영진만 문자 발송 API를 호출할 수 있고 비운영진은 `403 Forbidden`을 받는지 확인한다.
- 회원 목록에 현재 회원으로 관리되는 회원과 교류 회원 제외 조건이 정확히 적용되는지 확인한다.
- 회원 전화번호가 없는 회원에게는 외부 발송이 수행되지 않는지 확인한다.
- 직접 지정 전화번호의 하이픈·공백·괄호가 제거되고, 숫자가 하나도 없는 항목만 제외되며, 모든 항목이 제외되면 `400 Bad Request`가 반환되는지 확인한다.
- 숫자가 남지만 SOLAPI가 유효하지 않은 번호로 거부한 직접 지정 수신자는 `422 Unprocessable Entity`의 `FAILED` 결과로 반환되는지 확인한다.
- 메시지가 공백만으로 이루어졌을 때 `400 Bad Request`가 반환되는지 확인한다.
- ASCII 메시지와 한글 메시지의 90바이트 경계에서 SMS/LMS가 정확히 선택되는지 확인한다.
- `{realname}` 치환 후 90바이트 이하인 수신자는 SMS, 초과한 수신자는 LMS로 각각 처리되는지 확인한다.
- LMS 제목과 공식 발신번호가 외부 서비스 요청에 포함되는지 확인한다.
- SOLAPI 요청이 `/messages/v4/send-many/detail`에 JSON으로 전달되고 HMAC-SHA256 인증 header가 적용되는지 확인한다.
- 서버가 계산한 `type`, `autoTypeDetect=false` 및 비개인정보 수신자 순번이 수신자별 SOLAPI 요청 데이터에 포함되는지 확인한다.
- 회원 수신자와 직접 지정 수신자가 한 요청에서 함께 처리되는지 확인한다.
- 직접 지정 전화번호 배열의 한 원소에 쉼표로 여러 번호를 전달해도 각 번호가 별도 수신자로 정규화되는지 확인한다.
- 중복 수신자는 결과에서 제외되고 회원 수신자와 직접 지정 수신자가 계약에 정의된 순서로 반환되는지 확인한다.
- SOLAPI의 정상 응답 안에 일부 또는 전체 수신자 거부 결과가 있으면 다른 수신자의 정상 접수를 유지하고 `422 Unprocessable Entity`와 수신자별 결과를 반환하는지 확인한다.
- SOLAPI 요청의 timeout·네트워크 오류·인증 오류·비정상 HTTP 응답이 민감정보 없이 한국어 공통 오류와 `500 Internal Server Error`로 반환되는지 확인한다.
- `SOLAPI_API_KEY`, `SOLAPI_API_SECRET` 또는 발신번호가 비어 있으면 외부 발송을 수행하지 않고 `500 Internal Server Error`를 반환하는지 확인한다.
- 발신번호 변경 성공 이후 다른 application instance의 조회와 문자 발송에도 이전 cache 값이 사용되지 않는지 확인한다.
- 같은 요청을 재시도할 때 멱등성을 보장하지 않는 동작을 확인한다.
- 공식 발신번호 조회·변경과 운영진 권한 검사를 확인한다.
- 백엔드 테스트를 실행한다.

## 비목표

- 일반 회원의 문자 발송 기능
- 예약 발송, 반복 발송 및 자동 발송 일정
- 문자 발송 이력 조회와 통계
- SOLAPI 잔액 조회 및 충전
- 광고성 문자 동의·수신 거부 관리 체계
- SMS/LMS 이외의 MMS, 알림톡 및 푸시 알림 연동
- 레거시 문자 발송 데이터의 마이그레이션
