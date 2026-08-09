package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAcceptsValidPolicyDocumentFormats(t *testing.T) {
	root := createRepository(t)
	writeContent(t, root, "policy/빈_도메인/POLICY.md", "# 빈 도메인 정책\n")
	writeContent(t, root, "policy/운영진/POLICY.md", "# 운영진 정책\n")
	writeContent(t, root, "policy/범위만_있는_도메인/POLICY.md", `# 범위만 있는 도메인 정책

## 적용 범위

이 문서는 아직 정책이 정의되지 않은 도메인의 책임 범위를 설명한다.

두 번째 일반 문단도 사용할 수 있다.
`)
	writeContent(t, root, "policy/참가_신청/POLICY.md", `# 참가 신청 정책

## 적용 범위

이 문서는 참가 신청의 생성과 상태 전이를 정의한다.

## 정책

### 참가 신청 상태 전이

| 현재 상태 | 처리 | 변경 상태 |
| --- | --- | --- |
| 대기 | 승인 | 승인 |
| 대기 | 거절 | 거절 |

- 표에 정의되지 않은 상태 전이는 허용하지 않는다.

`+"```mermaid"+`
stateDiagram-v2
    대기 --> 승인: 승인
    대기 --> 거절: 거절
`+"```"+`

- 상태 전이의 판단은 요청 시점의 현재 상태를 기준으로 한다.

### 참가 신청 권한

- [운영진 정책](../운영진/POLICY.md)에서 정의한 운영진은 참가 신청을 승인할 수 있다.
  - 승인할 수 있는 신청은 대기 상태로 제한한다.
    - 신청이 대기 상태가 아니면 상태를 변경하지 않는다.
`)

	assertNoErrors(t, root)
}

func TestRunValidatesPolicyDocumentHeadings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "empty document",
			content:  "",
			expected: []string{"비어 있지 않은 H1 제목이 필요합니다"},
		},
		{
			name:     "content before H1",
			content:  "설명\n\n# 정책\n",
			expected: []string{"문서의 첫 번째 내용은 H1 제목이어야 합니다"},
		},
		{
			name:     "multiple H1 headings",
			content:  "# 첫 번째\n\n# 두 번째\n",
			expected: []string{"H1 제목은 정확히 하나만 존재해야 합니다"},
		},
		{
			name:     "empty H1 heading",
			content:  "#\n",
			expected: []string{"H1 제목은 비어 있을 수 없습니다"},
		},
		{
			name:     "content directly below H1",
			content:  "# 정책\n\n허용되지 않은 본문이다.\n",
			expected: []string{"H1 아래에는 허용된 H2 섹션만 작성할 수 있습니다"},
		},
		{
			name:     "unknown H2",
			content:  "# 정책\n\n## 참고\n\n내용\n",
			expected: []string{"허용되지 않은 H2 섹션입니다"},
		},
		{
			name: "duplicate H2",
			content: `# 정책

## 적용 범위

첫 번째 범위다.

## 적용 범위

두 번째 범위다.
`,
			expected: []string{"H2 섹션은 한 번만 사용할 수 있습니다"},
		},
		{
			name: "wrong H2 order",
			content: `# 정책

## 정책

### 판단

- 규칙이다.

## 적용 범위

범위다.
`,
			expected: []string{`"적용 범위" 섹션은 "정책" 섹션보다 앞에 있어야 합니다`},
		},
		{
			name:     "empty scope",
			content:  "# 정책\n\n## 적용 범위\n",
			expected: []string{`"적용 범위" 섹션은 비어 있을 수 없습니다`},
		},
		{
			name:     "scope with list",
			content:  "# 정책\n\n## 적용 범위\n\n- 범위다.\n",
			expected: []string{`"적용 범위" 섹션에는 일반 문단만 사용할 수 있습니다`},
		},
		{
			name:     "empty policy section",
			content:  "# 정책\n\n## 정책\n",
			expected: []string{`"정책" 섹션에는 하나 이상의 H3 정책 항목이 필요합니다`},
		},
		{
			name:     "content before policy item",
			content:  "# 정책\n\n## 정책\n\n- H3 없이 작성된 규칙이다.\n",
			expected: []string{`"정책" 섹션의 내용은 H3 정책 항목 아래에 작성해야 합니다`},
		},
		{
			name:     "empty H3 heading",
			content:  "# 정책\n\n## 정책\n\n###\n\n- 규칙이다.\n",
			expected: []string{"H3 정책 항목 제목은 비어 있을 수 없습니다"},
		},
		{
			name:     "H3 outside policy section",
			content:  "# 정책\n\n## 적용 범위\n\n범위다.\n\n### 판단\n",
			expected: []string{`H3 정책 항목은 "정책" 섹션 안에서만 사용할 수 있습니다`},
		},
		{
			name: "duplicate H3",
			content: `# 정책

## 정책

### 같은 판단

- 첫 번째 규칙이다.

### 같은 판단

- 두 번째 규칙이다.
`,
			expected: []string{"H3 정책 항목 제목은 중복될 수 없습니다"},
		},
		{
			name:     "H4 heading",
			content:  "# 정책\n\n## 정책\n\n### 판단\n\n- 규칙이다.\n\n#### 하위 판단\n",
			expected: []string{"H4 이하 제목은 사용할 수 없습니다"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errors := lintPolicyContent(t, test.content)
			assertMessages(t, errors, test.expected...)
		})
	}
}

func TestRunValidatesPolicyItemBlocks(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []string
	}{
		{
			name:     "paragraph",
			body:     "정책을 일반 문단으로 작성한다.\n",
			expected: []string{"H3 정책 항목에는 '-' unordered list, 표와 Mermaid code block만 사용할 수 있습니다"},
		},
		{
			name:     "table without list",
			body:     "| 조건 | 결과 |\n| --- | --- |\n| 허용 | 처리 |\n",
			expected: []string{"H3 정책 항목에는 하나 이상의 '-' unordered list가 필요합니다"},
		},
		{
			name:     "ordered list",
			body:     "1. 규칙이다.\n",
			expected: []string{"ordered list는 사용할 수 없습니다"},
		},
		{
			name:     "asterisk marker",
			body:     "* 규칙이다.\n",
			expected: []string{"unordered list marker는 '-'만 사용할 수 있습니다"},
		},
		{
			name:     "nested plus marker",
			body:     "- 상위 규칙이다.\n  + 하위 규칙이다.\n",
			expected: []string{"unordered list marker는 '-'만 사용할 수 있습니다"},
		},
		{
			name:     "empty list item",
			body:     "-\n  - 하위 규칙이다.\n",
			expected: []string{"list item은 비어 있을 수 없습니다"},
		},
		{
			name:     "non Mermaid code block",
			body:     "- 규칙이다.\n\n```text\nvalue\n```\n",
			expected: []string{"Mermaid 이외의 code block은 사용할 수 없습니다"},
		},
		{
			name:     "empty Mermaid",
			body:     "- 규칙이다.\n\n```mermaid\n```\n",
			expected: []string{"Mermaid code block은 비어 있을 수 없습니다"},
		},
		{
			name:     "table without data row",
			body:     "- 규칙이다.\n\n| 조건 | 결과 |\n| --- | --- |\n",
			expected: []string{"표에는 비어 있지 않은 data row가 하나 이상 필요합니다"},
		},
		{
			name:     "table with empty cell",
			body:     "- 규칙이다.\n\n| 조건 | 결과 |\n| --- | --- |\n| 허용 | |\n",
			expected: []string{"표의 cell은 비어 있을 수 없습니다"},
		},
		{
			name:     "table with empty header cell",
			body:     "- 규칙이다.\n\n| 조건 | |\n| --- | --- |\n| 허용 | 처리 |\n",
			expected: []string{"표의 cell은 비어 있을 수 없습니다"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := "# 정책\n\n## 정책\n\n### 판단\n\n" + test.body
			errors := lintPolicyContent(t, content)
			assertMessages(t, errors, test.expected...)
		})
	}
}

func TestRunValidatesPolicyDocumentLinks(t *testing.T) {
	t.Run("valid relative and external links", func(t *testing.T) {
		root := createRepository(t)
		writeContent(t, root, "policy/운영진/POLICY.md", "# 운영진 정책\n")
		writeContent(t, root, "policy/기능/POLICY.md", `# 기능 정책

## 정책

### 사용 권한

- [운영진 정책](../운영진/POLICY.md)을 참조한다.
- [외부 문서](https://example.com/reference)를 참조할 수 있다.
`)
		assertNoErrors(t, root)
	})

	tests := []struct {
		name        string
		destination string
		expected    string
	}{
		{
			name:        "missing local target",
			destination: "../운영진/POLICY.md",
			expected:    "로컬 Markdown link 대상이 존재하지 않습니다",
		},
		{
			name:        "repository absolute path",
			destination: "/policy/운영진/POLICY.md",
			expected:    "로컬 Markdown link는 상대경로를 사용해야 합니다",
		},
		{
			name:        "outside repository",
			destination: "../../../outside.md",
			expected:    "로컬 Markdown link는 repository 밖을 가리킬 수 없습니다",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := "# 기능 정책\n\n## 정책\n\n### 사용 권한\n\n- [참조](" + test.destination + ")를 사용한다.\n"
			errors := lintPolicyContent(t, content)
			assertMessages(t, errors, test.expected)
		})
	}

	t.Run("POLICY fragment", func(t *testing.T) {
		root := createRepository(t)
		writeContent(t, root, "policy/운영진/POLICY.md", "# 운영진 정책\n")
		writeContent(t, root, "policy/기능/POLICY.md", "# 기능 정책\n\n## 정책\n\n### 사용 권한\n\n- [참조](../운영진/POLICY.md#권한)를 사용한다.\n")

		errors, err := Run(root)
		if err != nil {
			t.Fatal(err)
		}
		assertMessages(t, errors, "다른 정책 문서는 POLICY.md 전체를 참조해야 하며 fragment를 사용할 수 없습니다")
	})
}

func lintPolicyContent(t *testing.T, content string) []Error {
	t.Helper()
	root := createRepository(t)
	writeContent(t, root, "policy/테스트/POLICY.md", content)
	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	return errors
}

func writeContent(t *testing.T, root string, relativePath string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
