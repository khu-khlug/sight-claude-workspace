package lint

import "testing"

func TestRunAcceptsValidHistoryDocumentFormats(t *testing.T) {
	root := createRepository(t)
	writeContent(t, root, "policy/행사/POLICY.md", "# 행사 정책\n")
	writeContent(t, root, "policy/행사/HISTORY.md", `## 참가 확정 기준 변경 (2026-08-08)

기존 정책과 변경된 정책의 차이를 설명한다.

- AS-IS: 참가비 납부만으로 참가를 확정했다.
- TO-BE: 신청 승인과 참가비 납부가 모두 필요하다.

> 당시 판단에 참고한 상황을 인용문으로 작성할 수도 있다.

| 구분 | 내용 |
| --- | --- |
| 이유 | 승인되지 않은 신청의 참가 확정을 방지한다. |

## 참가 취소 기준 변경 (2026-08-09)

같은 날짜의 기록을 여러 개 작성할 수 있다.

## 환불 기준 변경 (2026-08-09)

본문의 Markdown 형식은 heading을 제외하고 제한하지 않는다.

1. 관련 판단을 순서가 있는 목록으로 작성할 수도 있다.

`+"```text"+`
관련 자료의 일부를 기록할 수도 있다.
`+"```"+`
`)

	assertNoErrors(t, root)
}

func TestRunValidatesHistoryDocumentStructure(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "empty document",
			content:  "",
			expected: []string{"하나 이상의 H2 변경 기록이 필요합니다"},
		},
		{
			name:     "content before first entry",
			content:  "변경 이력에 대한 설명이다.\n\n## 기준 변경 (2026-08-09)\n\n변경 내용이다.\n",
			expected: []string{"첫 H2 변경 기록 앞에는 본문을 작성할 수 없습니다"},
		},
		{
			name:     "H1 heading",
			content:  "# 변경 이력\n\n## 기준 변경 (2026-08-09)\n\n변경 내용이다.\n",
			expected: []string{"변경 기록을 구분하는 H2 이외의 heading은 사용할 수 없습니다"},
		},
		{
			name:     "H3 heading",
			content:  "## 기준 변경 (2026-08-09)\n\n### 변경 이유\n\n변경 내용이다.\n",
			expected: []string{"변경 기록을 구분하는 H2 이외의 heading은 사용할 수 없습니다"},
		},
		{
			name:     "nested H2 heading",
			content:  "## 기준 변경 (2026-08-09)\n\n> ## 인용문 안의 제목\n",
			expected: []string{"변경 기록을 구분하는 최상위 H2 이외의 heading은 사용할 수 없습니다"},
		},
		{
			name:     "empty entry body",
			content:  "## 첫 번째 변경 (2026-08-08)\n\n## 두 번째 변경 (2026-08-09)\n\n변경 내용이다.\n",
			expected: []string{"H2 변경 기록의 본문은 비어 있을 수 없습니다"},
		},
		{
			name:     "missing date",
			content:  "## 기준 변경\n\n변경 내용이다.\n",
			expected: []string{"H2 변경 기록은 '변경 내용 (YYYY-MM-DD)' 형식이어야 합니다"},
		},
		{
			name:     "empty title",
			content:  "## (2026-08-09)\n\n변경 내용이다.\n",
			expected: []string{"H2 변경 기록은 '변경 내용 (YYYY-MM-DD)' 형식이어야 합니다"},
		},
		{
			name:     "invalid date format",
			content:  "## 기준 변경 (2026-8-9)\n\n변경 내용이다.\n",
			expected: []string{"H2 변경 기록은 '변경 내용 (YYYY-MM-DD)' 형식이어야 합니다"},
		},
		{
			name:     "invalid calendar date",
			content:  "## 기준 변경 (2026-02-29)\n\n변경 내용이다.\n",
			expected: []string{"H2 변경 기록에 유효한 날짜가 필요합니다"},
		},
		{
			name: "descending dates",
			content: `## 첫 번째 변경 (2026-08-09)

첫 번째 변경 내용이다.

## 두 번째 변경 (2026-08-08)

두 번째 변경 내용이다.
`,
			expected: []string{"변경 기록은 날짜 오름차순으로 작성해야 합니다"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errors := lintHistoryContent(t, test.content)
			assertMessages(t, errors, test.expected...)
		})
	}
}

func lintHistoryContent(t *testing.T, content string) []Error {
	t.Helper()
	root := createRepository(t)
	writeContent(t, root, "policy/테스트/POLICY.md", "# 테스트 정책\n")
	writeContent(t, root, "policy/테스트/HISTORY.md", content)
	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	return errors
}
