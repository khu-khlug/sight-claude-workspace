# sight-workspace

Sight 서비스를 구성하는 여러 저장소를 하나의 작업 공간에서 함께 개발하기 위한 통합 워크스페이스입니다.

프론트엔드와 백엔드처럼 서로 연관된 변경을 AI 에이전트와 한 번에 설계하고 구현할 수 있도록, 실제 코드 저장소들을 이 디렉터리 아래에 배치해 사용합니다. 각 코드 저장소는 독립적인 Git 저장소와 개발·배포 흐름을 유지합니다.

## 목적

- 프론트엔드와 백엔드를 포함한 여러 저장소의 통합 개발 환경을 제공합니다.
- 저장소 간 요구사항과 작업 진행 상황을 하나의 Task 체계로 관리합니다.
- AI 에이전트가 따라야 할 공통 규칙, 정책 및 작업 표준을 관리합니다.
- 개별 저장소의 규칙을 존중하면서 저장소 간 변경을 일관되게 조율합니다.

## 구조

작업할 때는 `sight-workspace` 저장소를 먼저 clone한 뒤, 프론트엔드와 백엔드 등 관련 코드 저장소를 `sight-workspace` 디렉터리 안에 각각 clone합니다. 이렇게 구성하면 하나의 작업 공간에서 여러 저장소의 코드를 함께 살펴보고 변경할 수 있으며, 결과적으로 다음과 같은 디렉터리 구조가 됩니다.

```text
sight-workspace/
├── sight-spring-backend/  # 백엔드 Git 저장소
├── sight-frontend/        # 프론트엔드 Git 저장소
├── tasks/                 # 통합 Task 문서
├── tools/                 # 워크스페이스 관리 도구
└── README.md
```

각 하위 코드 저장소는 독립적인 Git 저장소이며 `sight-workspace` 저장소에는 포함하지 않습니다. `sight-workspace`는 사이트 코드 자체보다 여러 저장소에 걸친 개발 컨텍스트와 공통 운영 기준을 관리합니다.

## 작업 원칙

- 사이트 코드의 커밋과 Pull Request는 해당 코드 저장소에서 관리합니다.
- 저장소 전체에 적용되는 Task, 규칙, 정책 및 표준은 이 워크스페이스에서 관리합니다.
- 각 코드 저장소에서 작업할 때는 해당 저장소의 고유한 지침도 함께 준수합니다.
- 여러 저장소를 변경하는 작업은 하나의 통합 Task를 기준으로 추적하되, 구현과 검증은 저장소별로 수행합니다.

## Task 기반 작업 진행

Task 문서는 구현 전에 `tasks/open/`에 작성하고 검토하는 것을 원칙으로 합니다. 현재 Task Standard는 백엔드 저장소에서 사용하던 정책을 그대로 이전한 것으로, 내부 구현 방법보다 HTTP API, database, 외부 시스템과의 계약, 사용자에게 관찰되는 비즈니스 동작, 보안 및 운영 영향을 작성합니다.

- 작성 원칙: [`tasks/STANDARD.md`](tasks/STANDARD.md)
- 진행 중인 Task: `tasks/open/`
- 완료된 Task: `tasks/completed/`

Task 문서를 생성하거나 수정한 뒤에는 현재 OS와 architecture에 맞는 `tools/task-lint/bin/task-lint-*` binary로 필수 섹션과 내용을 검증해야 합니다. binary가 없거나 source와 일치하지 않으면 다음 명령으로 다시 build합니다.

```bash
docker buildx build --file tools/task-lint/Dockerfile --output type=local,dest=tools/task-lint/bin .
```

자세한 build 및 실행 방법은 [`tools/task-lint/README.md`](tools/task-lint/README.md)를 참고합니다.
