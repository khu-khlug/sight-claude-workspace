# Backend Task 문서 작성 Standard

## 목적

Backend Task 문서는 작업으로 인해 달라지는 계약, 비즈니스 동작, 비가역적 결정을 사람이 구현 전에 검토할 수 있도록 설명한다.

사람은 다음 내용을 검토한다.

- HTTP API 계약
- database 계약
- 외부 시스템 계약
- 사용자에게 관찰되는 비즈니스 동작
- 보안, 개인정보 및 비가역적 부작용
- 호환성 및 배포 전략

반복되는 구현 문제는 `rules/`로, 기계적으로 판별할 수 있는 규칙은 lint로 관리한다. Task에는 내부 class, method, 파일 구성 등 외부 계약에 포함되지 않는 구현 세부 사항의 재량 범위를 나열하지 않는다.

## 저장 위치와 상태

- Backend Task 문서는 문서 시작의 YAML frontmatter에 `type: backend`를 단일 값으로 선언한다.
- 완료되지 않은 Task 문서는 `tasks/open/`에 둔다.
- 완료된 Task 문서는 `tasks/completed/`로 옮긴다.
- `tasks/_schema/backend.yaml`은 Task 문서의 필수·선택 입력과 섹션 순서를 정의한다.

## 작성 원칙

### 도메인 용어로 동작을 설명한다

비즈니스 동작은 저장소와 이해관계자가 공유하는 유비쿼터스 언어로 작성한다.

- 동작의 주체와 대상
- 허용 및 거부 조건
- 상태 전이
- 저장하거나 전달하는 데이터
- 외부에서 관찰되는 성공과 실패

일반적인 기술 용어, 비즈니스 용어, 저장소에서 영어로 사용하는 용어는 그대로 사용한다. 각 문장만으로 의미와 다음 동작을 예측할 수 있게 작성한다.

HTTP field, database schema, configuration key, event field 등 계약에 포함되는 identifier는 정확한 이름을 작성한다.

### 변경 사항과 유지할 계약을 작성한다

사람이 변경 범위와 호환성 조건을 구분해 검토할 수 있도록 작성한다.

모든 필수 영역에 변경 사항을 작성하거나 `변경 없음`을 명시한다. 작업 중 달라질 가능성이 있거나 호환성을 보장해야 하는 기존 계약은 유지할 계약으로 작성한다.

```markdown
## HTTP API 계약

### 변경 사항

- 성공 응답에 검토 시각을 나타내는 `reviewedAt` field를 추가한다.
- `reviewedAt`은 ISO 8601 형식의 문자열이다.

### 유지할 계약

- `POST /manager/applications/{applicationId}/approve`를 유지한다.
- 운영진만 요청할 수 있다.
- 가입 신청이 존재하지 않을 때의 `404 Not Found` 응답을 유지한다.
- 이미 처리된 가입 신청에는 `409 Conflict`를 반환한다.
```

### 확정된 내용만 작성한다

Backend Task는 작업에 영향을 주는 요구사항, 정책, 계약과 설계 결정이 모두 확정된 후에 작성한다. 사람의 판단이 필요하거나 두 가지 이상의 해석이 가능한 내용이 남아 있으면 Task에 질문, 선택지, 권장안으로 기록하지 않는다. Task 작성을 중단하고 사용자에게 질문한 뒤, 답변으로 확정된 내용을 반영한다.

### REST 리소스를 기준으로 HTTP API를 설계한다

백엔드는 자신이 소유하고 생명주기와 규칙을 관리하는 리소스를 중심으로 REST API를 제공한다.

- URI는 특정 화면, page, component 또는 클라이언트 flow가 아니라 리소스와 collection을 표현한다.
- HTTP method, status code, header와 representation은 리소스에 대한 조회, 생성, 변경, 삭제의 표준 의미를 따른다.
- 필터링, 정렬, pagination과 연관 리소스는 리소스 모델과 다양한 consumer의 재사용을 기준으로 설계한다.
- 특정 화면에 한번에 표시하기 위해 서로 다른 리소스를 하나의 화면 전용 response로 합치거나 화면 이름으로 endpoint를 만들지 않는다.
- 프론트엔드는 자신의 사용자 경험 설계에 필요한 리소스 API를 각각 선택하여 호출하고, 결과를 프론트엔드에서 구성한다.

### Frontend와 Backend의 개념을 분리한다

Backend Task는 백엔드가 소유한 도메인 리소스, 비즈니스 규칙과 외부 계약을 기준으로 작성한다. 화면 구성, component 상태, 화면 전용 표시 모델과 프론트엔드 flow를 Backend의 개념이나 API 설계 근거로 사용하지 않는다.

Frontend와 Backend는 각자의 설계 요구사항에 따라 독립적으로 구현한다. 두 저장소가 공유하는 의존점은 확정된 HTTP API 계약으로 한정하며, 한쪽의 내부 모델이나 화면 구성을 다른 쪽의 개념으로 도입하지 않는다.

## 입력 작성 지침

Task 문서는 `tasks/_schema/backend.yaml`에 정의된 섹션을 같은 이름과 순서로 작성한다. 필수 여부는 YAML schema를 기준으로 한다.

### 1. 작업 개요

- 해결하려는 문제 또는 제공하려는 기능
- 작업의 대상 사용자 또는 시스템
- 완료 여부를 판단할 수 있는 결과

### 2. 비즈니스 동작

도메인 용어로 다음 내용을 작성한다.

- 동작의 시작 조건
- 주요 상태와 상태 전이
- 허용 및 거부 조건
- 성공 결과
- 실패 결과
- 중복 요청 또는 재시도 동작

변경이 없다면 `변경 없음`을 작성한다.

### 3. HTTP API 계약

API마다 다음 내용을 작성한다.

- Backend가 소유하는 대상 리소스와 생명주기
- HTTP method와 path
- 인증 및 인가 조건
- path, query, header, cookie parameter
- request body
- response status와 body
- error status와 외부에서 구분 가능한 error
- field 이름, type, 필수 여부, nullable 여부 및 의미
- pagination, filtering, sorting
- 멱등성 및 중복 요청 처리
- 기존 API와의 호환성
- 특정 화면이 아닌 다른 consumer에서도 같은 리소스 계약으로 사용할 수 있는지

해당하지 않거나 변경이 없다면 `변경 없음`을 작성한다.

### 4. 데이터베이스 계약

- 추가, 변경, 삭제하는 table과 column
- type, 길이 또는 precision
- nullable 여부와 default
- primary key, foreign key, unique 및 check constraint
- index
- 제한된 저장값과 각 값의 의미
- 연관관계와 삭제 정책
- 기존 데이터에 적용할 migration 또는 backfill
- 대상 데이터 규모와 lock 또는 성능 위험
- 구버전과 신버전 application의 동시 실행 호환성
- rollback 가능 여부와 방법

해당하지 않거나 변경이 없다면 `변경 없음`을 작성한다.

### 5. 외부 시스템 계약

외부 연동마다 다음 내용을 작성한다.

- 대상 시스템과 연동 목적
- request, response, event 또는 message payload
- 인증 방식과 필요한 권한
- timeout, retry 및 rate limit 처리
- 중복 전달과 재처리 동작
- 부분 실패 동작
- 기존 consumer 또는 provider와의 호환성

해당하지 않거나 변경이 없다면 `변경 없음`을 작성한다.

### 6. 보안 및 개인정보

- 인증 및 인가 범위
- 수집, 저장, 반환 또는 전달하는 개인정보와 민감정보
- log, metric, tracing에 포함되는 민감정보
- 데이터 보존 및 삭제 조건
- 권한 상승 또는 정보 노출 가능성

해당하지 않거나 변경이 없다면 `변경 없음`을 작성한다.

### 7. 비가역적 부작용 및 운영 영향

- 데이터 삭제 또는 대량 변경
- 외부 알림과 message 발송
- 비용을 발생시키는 외부 호출
- batch와 scheduler의 실행 조건
- 예상 호출량 또는 처리량 변화
- 장애 시 중단, 재개 및 복구 방법
- 필요한 log, metric 또는 alert

해당하지 않거나 변경이 없다면 `변경 없음`을 작성한다.

### 8. 호환성 및 배포

- 하위 호환 여부
- 함께 변경할 consumer
- 단계적 배포 또는 feature flag
- migration, backfill, application 배포 순서
- 구버전과 신버전이 함께 실행되는 동안의 동작
- rollback 조건과 제한

고려할 내용이 없다면 `변경 없음`을 작성한다.

### 9. 검증

- 주요 성공 흐름
- 주요 실패 흐름
- 권한
- 경계값 및 상태 전이
- 중복 요청과 재시도
- migration 및 기존 데이터
- 외부 시스템 실패

각 항목의 기대 결과와 검증 방법을 작성한다.

### 10. 비목표

이번 작업에서 다루지 않는 인접 영역을 작성한다.

비목표가 없다면 `없음`을 작성한다.

## 사람의 검토가 필요한 변경

- HTTP API 계약의 추가, 변경 또는 삭제
- 인증 또는 인가 조건 변경
- database schema, constraint, index 또는 저장값 의미 변경
- migration 또는 backfill
- 외부 시스템 계약 변경
- 사용자에게 관찰되는 비즈니스 흐름 또는 규칙 변경
- 개인정보 또는 민감정보 처리 변경
- 데이터 삭제, 대량 변경, 외부 발송 등 비가역적 부작용
- 호환되지 않는 변경

위 내용에 대한 요구사항이나 결정이 모호하면 Task를 작성하지 않고 사용자에게 질문하여 먼저 확정한다.

## Task 문서, schema, `rules/` 및 lint의 역할

- Task 문서는 이번 작업의 계약, 비즈니스 의미와 승인 경계를 기록한다.
- YAML schema는 type별 필수·선택 입력과 기계적으로 검증할 형식을 정의한다.
- `rules/`는 여러 작업에 반복 적용할 구현 가드레일을 정의한다.
- lint는 YAML schema에 따라 필수 내용의 존재와 문서 형식을 deterministic하게 검사한다.
- 사람은 확정된 계약, 비즈니스 의미와 설계가 의도에 맞게 기록되었는지 검토한다.

현재 lint는 `tasks/open/`과 `tasks/completed/` 아래의 모든 Markdown 문서에 대해 다음 내용을 검사한다. 그 밖의 `tasks/` 하위 디렉터리는 검사하지 않는다.

- YAML frontmatter의 단일 `type`에 대응하는 `tasks/_schema/{type}.yaml`과 `tasks/_standard/{type}.md`가 존재한다.
- schema에 정의된 모든 필수 입력이 존재한다.
- 선택 입력은 생략할 수 있지만 작성하면 schema의 type과 순서를 따른다.
- 작성된 schema section에 공백을 제외한 내용이 한 글자 이상 존재한다.

추가 `##` 섹션은 허용한다.
