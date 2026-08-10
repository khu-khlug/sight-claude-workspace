# Task lint

## Binary build

저장소 root에서 다음 명령을 실행한다.

```text
docker buildx build --file tools/task-lint/Dockerfile --output type=local,dest=tools/task-lint/bin .
```

생성된 binary는 `tools/task-lint/bin/`에 저장되며 Git에 포함하지 않는다.

## 실행

macOS Apple Silicon:

```text
tools/task-lint/bin/task-lint-darwin-arm64
```

macOS Intel:

```text
tools/task-lint/bin/task-lint-darwin-amd64
```

Windows x64:

```text
tools\task-lint\bin\task-lint-windows-amd64.exe
```

Linux x64:

```text
tools/task-lint/bin/task-lint-linux-amd64
```

lint source가 binary build 이후 변경되면 실행을 중단하고 binary를 다시 build하도록 안내한다.

## 검증 범위

Task 문서는 문서 시작의 YAML frontmatter에 다음과 같이 단일 `type`을 선언한다.

```yaml
---
type: backend
---
```

lint는 `tasks/_schema/{type}.yaml`을 해당 Task의 schema로 선택하고 `tasks/_standard/{type}.md`가 함께 존재하는지 확인한다. YAML schema는 frontmatter field와 Markdown `##` section의 필수 여부, type, 순서, 추가 입력 허용 여부와 선택적인 section content 형식을 정의한다. Markdown Standard는 각 입력에 작성할 의미와 예시를 설명하며 lint 기준을 중복 정의하지 않는다.

지원하는 frontmatter field type은 다음과 같다.

- `string`
- `integer`
- `boolean`
- `string-list`

Schema는 다음 형식으로 작성한다.

```yaml
version: 1
type: backend

frontmatter:
  additionalFields: false
  fields:
    - name: type
      type: string
      required: true
      const: backend

sections:
  - name: 작업 개요
    required: true
  - name: 참고 사항
    required: false

additionalSections: true
```

- schema 파일명, schema의 `type`, frontmatter `type` field의 `const`는 서로 같아야 한다.
- `required: true`인 field와 section은 반드시 작성한다.
- `required: false`인 입력은 생략할 수 있지만 작성하면 선언된 type, 중복 및 순서 검사를 받는다.
- `additionalFields`와 `additionalSections`는 schema에 선언되지 않은 입력의 허용 여부를 결정한다.
- 지원하지 않는 schema key는 오류로 처리한다.

### 구조화된 section content

일반 section은 자유로운 Markdown 내용을 허용한다. Section에 `content`를 선언하면 전체 내용이 `oneOf` variant 중 하나와 일치해야 한다.

정확한 단일 값은 `literal`로 선언한다. 반복되는 구조화된 입력은 `records`로 선언한다.

```yaml
sections:
  - name: 허용 및 제한 정책
    required: true
    content:
      oneOf:
        - literal: 정책 변경 없음
        - type: records
          headingLevel: 3
          minItems: 1
          additionalFields: false
          fields:
            - name: 주체
              type: string
              required: true
            - name: 판정
              type: enum
              required: true
              values:
                - "허용"
                - "금지"
```

`records` 내용은 다음 Markdown 형식을 사용한다.

```markdown
## 허용 및 제한 정책

### 권한이 없는 사용자의 변경

- 주체: 일반 회원
- 판정: 금지
```

- `headingLevel` 제목 하나를 record 하나로 인식한다. 지원 범위는 3 이상 6 이하이다.
- Record 제목은 section 안에서 중복될 수 없다.
- Field는 `- 이름: 값` 형식의 한 줄 목록 항목으로 작성한다.
- Schema에 선언된 field끼리의 상대적 순서를 유지해야 한다.
- `string`과 `enum` field type을 지원한다.
- `enum` 값은 schema의 `values` 중 하나와 정확히 일치해야 한다.
- `required: true`인 field는 모든 record에 정확히 한 번 존재하고 값이 비어 있지 않아야 한다.
- `additionalFields`는 schema에 선언되지 않은 field의 허용 여부를 결정한다.
- `literal`은 section의 공백을 제외한 전체 내용과 정확히 일치해야 한다.
- `records` 형식 안에는 fenced code block을 작성할 수 없다.

`tasks/open/`과 `tasks/completed/` 아래의 모든 Markdown 문서는 다음 조건을 만족해야 한다. 그 밖의 `tasks/` 하위 디렉터리는 검증하지 않는다.

- schema의 모든 필수 section이 정확히 한 번 존재한다.
- 선택 section은 생략할 수 있지만 작성하면 정확히 한 번 존재해야 한다.
- schema에 선언된 section끼리의 상대적 순서가 같다.
- 작성된 schema section의 내용은 공백을 제외하고 한 글자 이상이다.
- `content`가 선언된 section의 내용은 `literal` 또는 `records` 형식 중 하나와 일치한다.
- `type`은 소문자로 시작하고 소문자, 숫자, hyphen만 포함하는 단일 값이다.
- `tasks/_schema/{type}.yaml`과 `tasks/_standard/{type}.md`가 존재한다.

추가 `##` 섹션은 schema의 `additionalSections`가 `true`일 때만 허용한다. fenced code block과 YAML frontmatter 안의 `##`는 섹션으로 인식하지 않는다.
