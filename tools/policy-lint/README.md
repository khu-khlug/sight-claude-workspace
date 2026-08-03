# Policy lint

## Binary build

저장소 root에서 다음 명령을 실행한다.

```text
docker buildx build --file tools/policy-lint/Dockerfile --output type=local,dest=tools/policy-lint/bin .
```

생성된 binary는 `tools/policy-lint/bin/`에 저장되며 Git에 포함하지 않는다. Docker build는 unit test를 실행한 뒤 macOS, Windows 및 Linux용 binary를 생성한다.

## 실행

macOS Apple Silicon:

```text
tools/policy-lint/bin/policy-lint-darwin-arm64
```

macOS Intel:

```text
tools/policy-lint/bin/policy-lint-darwin-amd64
```

Windows x64:

```text
tools\policy-lint\bin\policy-lint-windows-amd64.exe
```

Windows ARM64:

```text
tools\policy-lint\bin\policy-lint-windows-arm64.exe
```

Linux x64:

```text
tools/policy-lint/bin/policy-lint-linux-amd64
```

Linux ARM64:

```text
tools/policy-lint/bin/policy-lint-linux-arm64
```

저장소 하위 경로에서 실행하면 `policy/`와 `tools/policy-lint/`가 있는 저장소 root를 상위 경로에서 찾는다. 다른 위치의 저장소를 검사하려면 `--root`에 경로를 지정한다.

```text
tools/policy-lint/bin/policy-lint-darwin-arm64 --root /path/to/sight-workspace
```

lint source가 binary build 이후 변경되면 실행을 중단하고 binary를 다시 build하도록 안내한다.

## CI

`.github/workflows/policy-lint.yml`은 `main` branch를 대상으로 하는 push와 pull request에서 실행된다. lint source와 Dockerfile을 기준으로 binary cache를 사용하고, cache가 없으면 Docker로 binary를 build한 뒤 Linux x64 binary로 정책 구조를 검사한다.

## 검증 범위

`policy/` 바로 아래에는 다음 항목만 존재할 수 있다.

- 일반 파일인 `STANDARD.md`
- 일반 파일인 `GLOSSARY.md`
- 도메인 디렉터리

`STANDARD.md`와 `GLOSSARY.md`는 반드시 존재해야 한다. 각 도메인 디렉터리는 다음 구조만 가질 수 있다.

```text
<도메인>/
├── POLICY.md
└── HISTORY.md
```

- `POLICY.md`는 반드시 존재하는 일반 파일이다.
- `HISTORY.md`는 선택적으로 존재할 수 있는 일반 파일이다.
- 그 밖의 파일, 하위 디렉터리, 심볼릭 링크 및 다른 종류의 항목은 허용하지 않는다.
- 도메인 디렉터리명의 문자 구성은 검사하지 않는다.
- Markdown 문서의 내용과 형식은 검사하지 않는다.

## 종료 코드

- `0`: lint 성공
- `1`: 허용되지 않은 정책 구조 발견
- `2`: 저장소 탐색, 파일 읽기 또는 binary source hash 검증 실패
